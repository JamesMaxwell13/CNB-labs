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

func CompareFlag(cmpBits []byte) bool {
	if len(cmpBits) == 8 &&
		(bytes.Equal(cmpBits, []byte{1, 0, 0, 0, 0, 1, 1, 1}) ||
			bytes.Equal(cmpBits, []byte{1, '\n', 0, 0, 0, 1, 1, 1}) ||
			bytes.Equal(cmpBits, []byte{1, 0, '\n', 0, 0, 1, 1, 1}) ||
			bytes.Equal(cmpBits, []byte{1, 0, 0, '\n', 0, 1, 1, 1}) ||
			bytes.Equal(cmpBits, []byte{1, 0, 0, 0, '\n', 1, 1, 1})) {
		return true
	}
	return false
}

func CompareStuffedFlag(cmpBits []byte) bool {
	if len(cmpBits) == 8 &&
		(bytes.Equal(cmpBits, []byte{1, 0, 0, 0, 0, 1, 1, 0}) ||
			bytes.Equal(cmpBits, []byte{1, '\n', 0, 0, 0, 1, 1, 0}) ||
			bytes.Equal(cmpBits, []byte{1, 0, '\n', 0, 0, 1, 1, 0}) ||
			bytes.Equal(cmpBits, []byte{1, 0, 0, '\n', 0, 1, 1, 0}) ||
			bytes.Equal(cmpBits, []byte{1, 0, 0, 0, '\n', 1, 1, 0})) {
		return true
	}
	return false
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

func (packet *Packet) PacketToRaw() []byte {
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
	stuffedPacket := BitStuffing(packet)
	formattedPacket := FindStuffedBits(stuffedPacket)
	log.Printf("Serialize packet: %s", formattedPacket)
	return stuffedPacket, formattedPacket, nil
}

func ParseRawData(rawData []byte) (string, error) {
	newText := ""
	for len(rawData) >= 26 {
		rawPacket := rawData[:26]
		rawData = rawData[26:]
		for len(rawData) >= 26 && !CompareFlag(rawData[:8]) {
			rawPacket = append(rawPacket, rawData[0])
			rawData = rawData[1:]
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
	if len(rawPacket) < 24 {
		return "", errors.New("Packet is too short")
	}
	deStuffedPacket, err := DeBitStuffing(rawPacket)
	if err != nil {
		return "", err
	}
	log.Printf("Deserialize packet: %s", rawPacket)
	data := DataToStr(deStuffedPacket.Data[:])
	return data, err
}

func BitStuffing(packet Packet) []byte {
	stuffedPacket := packet.PacketToRaw()
	for i := 7; i < len(stuffedPacket)-7; i++ {
		if CompareFlag(stuffedPacket[i : i+8]) {
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
			if CompareStuffedFlag(packet[i : i+8]) {
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
		if i+8 < len(packet) && (strPacket[i:i+8] == "10000110" ||
			strPacket[i:i+8] == "1\n000110" || strPacket[i:i+8] == "10\n00110" ||
			strPacket[i:i+8] == "100\n0110" || strPacket[i:i+8] == "1000\n110") {
			formattedPacket += strPacket[i : i+7]
			formattedPacket += "-" + string(strPacket[i+7]) + "-"
			stuffedBits++
			i += 7
		} else {
			if strPacket[i] == '\n' {
				formattedPacket += "\\n"
			} else {
				formattedPacket += string(strPacket[i])
			}
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

func Destortion(packet Packet) Packet {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	bitError := random.Intn(7)
	if Chance(30) && packet.Data[bitError] == 1 {
		packet.Data[bitError] = 0
	} else {
		packet.Data[bitError] = 1
	}
	return packet
}

func GetHammingFCS(data [7]byte) [3]byte {
	return [3]byte(data[:3])
}

func EliminatingDistortion(data [7]byte, FCS [3]byte) [7]byte {
	return data
}
