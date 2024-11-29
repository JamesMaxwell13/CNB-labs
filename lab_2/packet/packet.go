package packet

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"strings"
)

type Packet struct {
	Flag        [8]byte
	Destination [4]byte
	Source      [4]byte
	Data        [7]byte
	FCS         byte
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
		FCS:         0,
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

func (packet *Packet) PacketToRaw() []byte {
	var rawPacket []byte
	fields := [][]byte{
		packet.Flag[:],
		packet.Destination[:],
		packet.Source[:],
		packet.Data[:],
		{packet.FCS},
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
	for len(rawData) >= 24 {
		rawPacket := rawData[:24]
		rawData = rawData[24:]
		for len(rawData) >= 24 && !bytes.Equal(rawData[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			rawPacket = append(rawPacket, rawData[0])
			rawData = rawData[1:]
		}
		if !bytes.Equal(rawPacket[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			continue
		}
		if len(rawData) < 24 {
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
	log.Printf("Deserialize packet: %s", DataToStr(rawPacket))
	data := DataToStr(deStuffedPacket.Data[:])
	return data, err
}

func BitStuffing(packet Packet) []byte {
	stuffedPacket := packet.PacketToRaw()
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
	if len(packet) < 24 || !bytes.Equal(packet[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
		return Packet{}, errors.New("Invalid packet")
	} else {
		deStuffedPacket := NewPacket(0, "0000000")
		for i := 7; i < len(packet); i++ {
			if len(packet) == 24 {
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
		deStuffedPacket.FCS = packet[23]
		return deStuffedPacket, nil
	}
}

func FindStuffedBits(packet []byte) string {
	strPacket := DataToStr(packet)
	formattedPacket := strPacket[:7]
	stuffedBits := 0
	for i := 7; i < len(packet); i++ {
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
