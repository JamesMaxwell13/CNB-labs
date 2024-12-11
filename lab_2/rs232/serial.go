package rs232

import (
	"errors"
	"fmt"
	"go.bug.st/serial"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type Port struct {
	Name       string
	Number     int
	SerialPort serial.Port
}

func DefaultConfig() *serial.Mode {
	return &serial.Mode{
		BaudRate: 115200,
		DataBits: 8,
		StopBits: serial.OneStopBit,
		Parity:   serial.NoParity,
	}
}

func (p *Port) OpenPort(name string) (serial.Port, error) {
	port, err := serial.Open(name, DefaultConfig())
	if err != nil {
		log.Printf("Port %s open failed\n", name)
		return nil, err
	}
	p.Name = name
	p.SerialPort = port
	p.Number = extractNum(name)
	log.Printf("Port %s opened successful\n", name)
	return port, nil
}

func (p *Port) ClosePort() error {
	if p.SerialPort != nil {
		_ = p.SerialPort.ResetOutputBuffer()
		_ = p.SerialPort.ResetInputBuffer()
		err := p.SerialPort.Close()
		if err != nil {
			log.Printf("Port %s close failed\n", p.Name)
			return err
		}
		log.Printf("Port %s closed successfull\n", p.Name)
		p.SerialPort = nil
	}
	return nil
}

func (p *Port) WriteBytes(data []byte) error {
	if p.SerialPort == nil {
		return errors.New("Serial port is not open")
	}
	n, err := p.SerialPort.Write(data)
	if err != nil {
		return err
	}
	log.Printf("Written %d bytes to port %s\n", n, p.Name)
	//err = p.SerialPort.ResetOutputBuffer()
	//if err != nil {
	//	return err
	//}
	return nil
}

func (p *Port) ReadBytes() ([]byte, error) {
	if p.SerialPort == nil {
		return nil, errors.New("Serial port is not open")
	}
	buff := make([]byte, 256)
	n, err := p.SerialPort.Read(buff)
	if err != nil {
		return nil, err
	}
	//err = p.SerialPort.ResetInputBuffer()
	//if err != nil {
	//	return nil, err
	//}
	log.Printf("Read %d bytes from port %s\n", n, p.Name)
	return buff[:n], nil
}

func (p *Port) PortNumber() (int, error) {
	var numberStr string
	for i := len(p.Name) - 1; i >= 0; i-- {
		if p.Name[i] >= '0' && p.Name[i] <= '9' {
			numberStr = string(p.Name[i]) + numberStr
		} else {
			break
		}
	}
	if numberStr == "" {
		return 0, errors.New("No number found in port name")
	}
	return strconv.Atoi(numberStr)
}

func PortIsOpen(name string) bool {
	cmd := exec.Command("fuser", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	processes := strings.Fields(string(output))
	return len(processes) > 1
}

func PortIsOpenThisProcess(name string) bool {
	pid := os.Getpid()
	cmd := exec.Command("fuser", name)
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	processes := strings.Fields(string(output))
	for _, process := range processes {
		if process == fmt.Sprintf("%d", pid) {
			return true
		}
	}
	return false
}

type ByNumber []string

func (s ByNumber) Len() int {
	return len(s)
}

func (s ByNumber) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByNumber) Less(i, j int) bool {
	numI := extractNum(s[i])
	numJ := extractNum(s[j])
	return numI < numJ
}

func extractNum(s string) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] < '0' || s[i] > '9' {
			if i == len(s)-1 {
				return 0
			}
			num, err := strconv.Atoi(s[i+1:])
			if err != nil {
				return 0
			}
			return num
		}
	}
	num, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return num
}

func RemovePorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	sort.Sort(ByNumber(ports))
	if err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, errors.New("No serial ports found!")
	}
	//availablePorts := make([]string, 0)
	//for i := 0; i < len(ports); i++ {
	//	if !PortIsOpen(ports[i]) {
	//		availablePorts = append(availablePorts, ports[i])
	//	} else {
	//		if PortIsOpenThisProcess(ports[i]) {
	//			if i%2 == 0 {
	//				i++
	//			} else {
	//				if !PortIsOpen(ports[i-1]) && len(availablePorts) > 0 {
	//					availablePorts = availablePorts[:len(availablePorts)-1]
	//				}
	//			}
	//		}
	//	}
	//}
	return ports, nil
}
