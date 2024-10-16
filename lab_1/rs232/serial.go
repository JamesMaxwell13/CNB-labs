package rs232

import (
	"errors"
	"fmt"
	"go.bug.st/serial"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Port struct {
	Name       string
	SerialPort serial.Port
}

func DefaultConfig() *serial.Mode {
	return &serial.Mode{
		BaudRate: 256000,
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
		log.Printf("Port %s closed successfully\n", p.Name)
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
	err = p.SerialPort.ResetOutputBuffer()
	if err != nil {
		return err
	}
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
	err = p.SerialPort.ResetInputBuffer()
	if err != nil {
		return nil, err
	}
	log.Printf("Read %d bytes from port %s\n", n, p.Name)
	return buff[:n], nil
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

func RemovePorts(allPorts []string) ([]string, error) {
	if len(allPorts) == 0 {
		return nil, errors.New("No rs232 ports found!")
	}
	availablePorts := make([]string, 0)
	for i := 0; i < len(allPorts); i++ {
		if !PortIsOpen(allPorts[i]) {
			availablePorts = append(availablePorts, allPorts[i])
		} else {
			if PortIsOpenThisProcess(allPorts[i]) {
				if i%2 == 0 {
					i++
				} else {
					if !PortIsOpen(allPorts[i-1]) {
						availablePorts = append(availablePorts[0 : len(availablePorts)-1])
					}
				}
			}
		}
	}
	return availablePorts, nil
}
