package csma_cd

import (
	"bytes"
	"lab_4/gui"
	"lab_4/packet"
	"lab_4/rs232"
	"log"
	"math"
	"math/rand"
	"time"
)

func ChannelBusy() bool {
	return packet.Chance(70)
}

func Collision() bool {
	return packet.Chance(30)
}

func Delay(attempts int) {
	//log.Printf("Random delay for %d attempts", attempts)
	if attempts > 10 {
		attempts = 10
	}
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	times := random.Intn(int(math.Pow(2, float64(attempts))))
	log.Printf("Random delay: %d ms", times)
	time.Sleep(time.Duration(times) * time.Millisecond)
}

func Transmitter(rawPacket []byte, formattedPacket string, u *gui.UserInterface) {
	collisionInfo := ""
	transmittedBytes := 0
	for transmittedBytes < len(rawPacket) {
		attempts := 0
		for attempts <= 16 {
			if !ChannelBusy() {
				//log.Printf("Channel is free")
				err := u.InputPort.WriteBytes(rawPacket[transmittedBytes : transmittedBytes+1])
				if err != nil {
					gui.ErrorWindow(err, u.App)
				}
				if Collision() {
					attempts++
					//log.Printf("Transmitting collision")
					collisionInfo += "!"
					u.UpdateStatus(formattedPacket + " " + collisionInfo)
					err = u.InputPort.WriteBytes([]byte("j"))
					if err != nil {
						gui.ErrorWindow(err, u.App)
					}
					Delay(attempts)
				} else {
					collisionInfo += ". "
					transmittedBytes++
					break
				}
			}
		}
		u.TransmittedBytes++
		u.UpdateStatus(formattedPacket + " " + collisionInfo)
	}
}

func CleanPacketPrefix(rawData []byte) []byte {
	for i := 0; i < len(rawData)-7; i++ {
		if bytes.Equal(rawData[i:i+8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			rawData = rawData[i:]
		}
	}
	return rawData
}

func CheckPacket(rawData []byte) bool {
	length := len(rawData)
	lenPacket := 26
	if length >= 26 {
		rawData = CleanPacketPrefix(rawData)
		length = len(rawData)
		if length <= 28 && bytes.Equal(rawData[:8], []byte{1, 0, 0, 0, 0, 1, 1, 1}) {
			for i := 7; i < length; i++ {
				if i+8 <= length && bytes.Equal(rawData[i:i+7], []byte{1, 0, 0, 0, 0, 1, 1}) {
					lenPacket += 1
					i += 6
				}
			}
			if length == lenPacket {
				return true
			}
		}
	}
	return false
}

func Receiver(outputPort *rs232.Port) (string, error) {
	newText := ""
	rawPacket := make([]byte, 0)
	for !CheckPacket(rawPacket) {
		rawData, err := outputPort.ReadBytes()
		if err != nil && err.Error() != "Port has been closed" {
			return "", err
		}
		if len(rawData) < 1 {
			continue
		}
		for bitNum := range rawData {
			if rawData[bitNum] == 'j' && len(rawPacket) > 0 {
				//log.Printf("Receiving collision")
				rawPacket = rawPacket[0 : len(rawPacket)-1]
			} else {
				rawPacket = append(rawPacket, rawData[bitNum])
			}
		}
	}
	rawPacket = CleanPacketPrefix(rawPacket)
	data, err := packet.DeserializePacket(rawPacket)
	if err != nil {
		return newText, err
	}
	newText += data
	return newText, nil
}
