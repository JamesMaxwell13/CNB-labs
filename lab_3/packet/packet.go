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
	FCS         [4]byte
}

func NewPacket(source int, data string) Packet {
	if source > 15 {
		source %= 16
	}
	return Packet{
		Flag:        [8]byte{1, 0, 0, 0, 0, 1, 1, 1},
		Destination: [4]byte{0, 0, 0, 0},
		Source:      [4]byte(StrToByte(fmt.Sprintf("%04b", source))),
		Data:        [7]byte(StrToByte(data)),
		FCS:         [4]byte{0, 0, 0, 0},
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
	log.Printf("Serialize packet:\n%s", strings.ReplaceAll(DataToStr(packet.ToRaw()), "\n", "\\n"))
	return stuffedPacket, formattedPacket, nil
}

func ParseRawData(rawData []byte) (string, error) {
	newText := ""
	for len(rawData) >= 27 {
		rawPacket := rawData[:27]
		rawData = rawData[27:]
		for len(rawData) >= 27 && !bytes.Equal(rawData[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			rawPacket = append(rawPacket, rawData[0])
			rawData = rawData[1:]
		}
		if !bytes.Equal(rawPacket[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			continue
		}
		if len(rawData) < 27 {
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
	if len(rawPacket) < 27 {
		return "", errors.New("Packet is too short")
	}
	deStuffedPacket, err := DeBitStuffing(rawPacket)
	if err != nil {
		return "", err
	}
	log.Printf("Deserialize packet:\n%s", strings.ReplaceAll(DataToStr(rawPacket), "\n", "\\n"))
	deStuffedPacket.CleanDistortion()
	data := DataToStr(deStuffedPacket.Data[:])
	return data, err
}

func BitStuffing(packet Packet) []byte {
	stuffedPacket := packet.ToRaw()
	for i := 7; i < len(stuffedPacket)-7; i++ {
		if bytes.Equal(stuffedPacket[i:i+8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			stuffedPacket = append(stuffedPacket[:i+7],
				append([]byte{0}, stuffedPacket[i+7:]...)...)
			i += 7
		}
	}
	return stuffedPacket
}

func DeBitStuffing(packet []byte) (Packet, error) {
	if len(packet) < 27 || !bytes.Equal(packet[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
		return Packet{}, errors.New("Invalid packet")
	} else {
		deStuffedPacket := NewPacket(0, "0000000")
		for i := 7; i < len(packet); i++ {
			if len(packet) == 27 {
				break
			}
			if i+8 <= len(packet) && bytes.Equal(packet[i:i+8], []byte{1, 0, 0, 0, 0, 1, 1, 0}) {
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
			p.Data[bitError] = 1
		}
	}
	return *p
}

func (p *Packet) GetHammingFCS() [4]byte {
	data := p.Data
	for i := range data {
		if data[i] == '\n' {
			data[i] = 0
		}
	}
	p.FCS[0] = data[0] ^ data[1] ^ data[3] ^ data[4] ^ data[6]
	p.FCS[1] = data[0] ^ data[2] ^ data[3] ^ data[5] ^ data[6]
	p.FCS[2] = data[1] ^ data[2] ^ data[3]
	p.FCS[3] = data[4] ^ data[5] ^ data[6]
	return p.FCS
}

func (p *Packet) CleanDistortion() [7]byte {
	code := make([]byte, 4)
	code[0] = 0
	code[1] = 0
	code[2] = p.Data[0]
	code[3] = 0
	code = append(code[:4], p.Data[1:4]...)
	code = append(code, 0)
	code = append(code[:8], p.Data[4:]...)
	for i := 2; i > len(code); i++ {
		if code[i] == '\n' {
			code[i] = 0
		}
	}
	code[0] = code[2] ^ code[4] ^ code[6] ^ code[8] ^ code[10]
	code[1] = code[2] ^ code[5] ^ code[6] ^ code[9] ^ code[10]
	code[3] = code[4] ^ code[5] ^ code[6]
	code[7] = code[8] ^ code[9] ^ code[10]
	pos := 0

	if code[0] != p.FCS[0] {
		pos += 1
	}
	if code[1] != p.FCS[1] {
		pos += 2
	}
	if code[3] != p.FCS[2] {
		pos += 4
	}
	if code[7] != p.FCS[3] {
		pos += 8
	}
	if pos != 0 {
		if code[pos-1] == 0 {
			code[pos-1] = 1
		} else {
			code[pos-1] = 0
		}
	}

	buff := make([]byte, 1)
	buff[0] = code[2]
	buff = append(buff, code[4:7]...)
	buff = append(buff, code[8:]...)
	for i := range p.Data {
		if p.Data[i] == '\n' {
			buff[i] = '\n'
		}
	}
	copy(p.Data[:], buff)
	return p.Data
}

//			Hamming code
// data: 0   1 2 3   4 5 6
// 	 p p 1 p 0 1 0 p 1 0 1
//	 0 1 2 3 4 5 6 7 8 9 10
// 0 х   х   х   х   х   х	1
// 1   х х     х х     х х	1
// 2       х х х х			1
// 3				 х х х	0
