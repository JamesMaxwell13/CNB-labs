package packet

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"
)

type Packet struct {
	Flag        [8]byte
	Destination [4]byte
	Source      [4]byte
	Data        [7]byte
	FCS         [3]byte
}

func NewPacket(source int, data string) Packet {
	if source > 15 {
		source %= 16
	}
	fmt.Println(StrToByte(data))
	return Packet{
		Flag:        [8]byte{1, 0, 0, 0, 0, 1, 1, 1},
		Destination: [4]byte{0, 0, 0, 0},
		Source:      [4]byte(StrToByte(fmt.Sprintf("%04b", source))),
		Data:        [7]byte(StrToByte(data)),
		FCS:         [3]byte{0, 0, 0},
	}
}

func StrToByte(str string) []byte {
	var rawBytes []byte
	for _, char := range str {
		if char != '\n' {
			char -= '0'
		}
		rawBytes = append(rawBytes, byte(char))
	}
	return rawBytes
}

func DataToStr(rawBytes []byte) string {
	str := ""
	for _, rawByte := range rawBytes {
		if rawByte != '\n' {
			rawByte += '0'
		}
		str += string(rawByte)
	}
	return str
}

func (packet *Packet) ToRaw() []byte {
	var rawPacket []byte
	fields := [][]byte{
		packet.Flag[:],
		packet.Destination[:],
		packet.Source[:],
		packet.Data[:],
		packet.FCS[:],
	}
	for _, field := range fields {
		rawPacket = append(rawPacket, field...)
	}
	return rawPacket
}

func SerializePacket(data string, source int) ([]byte, string, error) {
	if len(data) != 7 {
		return nil, "", errors.New("Wrong data in packet")
	}
	packet := NewPacket(source, data)
	packet.GetHammingFCS()
	packet.Distortion()
	stuffedPacket := BitStuffing(packet)
	formattedPacket := FindStuffedBits(stuffedPacket)
	log.Printf("Serialize packet:\n%s", strings.ReplaceAll(formattedPacket, "\n", "\\n"))
	return stuffedPacket, formattedPacket, nil
}

func ParseRawData(rawData []byte) (string, error) {
	newText := ""
	for len(rawData) >= 26 {
		rawPacket := rawData[:26]
		rawData = rawData[26:]
		for len(rawData) >= 26 && !bytes.Equal(rawData[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			rawPacket = append(rawPacket, rawData[0])
			rawData = rawData[1:]
		}
		if !bytes.Equal(rawPacket[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			continue
		}
		if len(rawData) < 26 {
			rawPacket = append(rawPacket, rawData...)
		}
		data, err := DeserializePacket(rawPacket)
		if err != nil {
			return newText, err
		}
		newText += data
	}
	return newText, nil
}

func DeserializePacket(rawPacket []byte) (string, error) {
	if len(rawPacket) < 26 {
		return "", errors.New("Packet is too short")
	}
	log.Printf("Deserialize packet:\n%s", strings.ReplaceAll(DataToStr(rawPacket), "\n", "\\n"))
	deStuffedPacket, err := DeBitStuffing(rawPacket)
	if err != nil {
		return "", err
	}
	deStuffedPacket.CleanDistortion()
	data := DataToStr(deStuffedPacket.Data[:])
	return data, err
}

func BitStuffing(packet Packet) []byte {
	stuffedPacket := packet.ToRaw()
	for i := 7; i < len(stuffedPacket)-7; i++ {
		if bytes.Equal(stuffedPacket[i:i+7], []byte{1, 0, 0, 0, 0, 1, 1}) {
			stuffedPacket = append(stuffedPacket[:i+7],
				append([]byte{0}, stuffedPacket[i+7:]...)...)
			i += 7
		}
	}
	return stuffedPacket
}

func DeBitStuffing(packet []byte) (Packet, error) {
	if len(packet) < 26 || !bytes.Equal(packet[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
		return Packet{}, errors.New("Invalid packet")
	} else {
		deStuffedPacket := NewPacket(0, "0000000")
		for i := 7; i < len(packet); i++ {
			if len(packet) == 26 {
				break
			}
			if i+8 <= len(packet) && bytes.Equal(packet[i:i+7], []byte{1, 0, 0, 0, 0, 1, 1}) {
				packet = append(packet[:i+7], packet[i+8:]...)
				i += 6
			}
		}
		copy(deStuffedPacket.Flag[:], packet[:8])
		copy(deStuffedPacket.Destination[:], packet[8:12])
		copy(deStuffedPacket.Source[:], packet[12:16])
		copy(deStuffedPacket.Data[:], packet[16:23])
		copy(deStuffedPacket.FCS[:], packet[23:])
		return deStuffedPacket, nil
	}
}

func FindStuffedBits(packet []byte) string {
	strPacket := DataToStr(packet)
	formattedPacket := strPacket[:7]
	stuffedBits := 0
	for i := 7; i < len(packet); i++ {
		if i == 23+stuffedBits {
			formattedPacket += " "
		}
		if i+8 < len(packet) && strPacket[i:i+8] == "10000110" {
			formattedPacket += strPacket[i : i+7]
			formattedPacket += "-" + string(strPacket[i+7]) + "-"
			stuffedBits++
			i += 7
		} else {
			formattedPacket += string(strPacket[i])
		}
	}
	formattedPacket = strings.ReplaceAll(formattedPacket, "\n", "\\n")
	return formattedPacket
}

func Chance(percent int) bool {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	chance := random.Intn(100)
	return chance < percent
}

func (p *Packet) Distortion() Packet {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	bitError := random.Intn(7)
	if Chance(30) {
		if p.Data[bitError] == 1 {
			p.Data[bitError] = 0
		} else {
			if p.Data[bitError] == 0 {
				p.Data[bitError] = 1
			}
		}
	}
	return *p
}

func (p *Packet) GetHammingFCS() [3]byte {
	data := p.Data
	for i := range data {
		if data[i] != 0 && data[i] != 1 {
			data[i] = 0
		}
	}
	p.FCS[0] = data[0] ^ data[2] ^ data[4] ^ data[6]
	p.FCS[1] = data[1] ^ data[2] ^ data[5] ^ data[6]
	p.FCS[2] = data[3] ^ data[4] ^ data[5] ^ data[6]
	return p.FCS
}

//	 Hamming code
//	 1 1 1 1 1 1 1
//	 0 1 2 3 4 5 6
//
// 0 х   х   х   х 0
// 1   х х     х х 0
// 2       х х х х 0

func (p *Packet) CleanDistortion() [7]byte {
	var newFCS [3]byte
	data := p.Data
	for i := range data {
		if data[i] != 0 && data[i] != 1 {
			data[i] = 0
		}
	}
	newFCS[0] = data[0] ^ data[2] ^ data[4] ^ data[6]
	newFCS[1] = data[1] ^ data[2] ^ data[5] ^ data[6]
	newFCS[2] = data[3] ^ data[4] ^ data[5] ^ data[6]
	pos := 0

	if newFCS[0] != p.FCS[0] {
		pos += 1
	}
	if newFCS[1] != p.FCS[1] {
		pos += 2
	}
	if newFCS[2] != p.FCS[2] {
		pos += 4
	}

	if pos >= 1 {
		if p.Data[pos-1] == 0 {
			p.Data[pos-1] = 1
		} else {
			p.Data[pos-1] = 0
		}
	}
	return p.Data
}
