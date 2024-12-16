package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"go.bug.st/serial"
	"lab_4/csma_cd"
	"lab_4/gui"
	"lab_4/packet"
	"lab_4/rs232"
	"runtime"
	"strconv"
	"time"
)

func TransmitData(u *gui.UserInterface) {
	prevText := ""
	for {
		if u.InputEntry != nil && u.InputPort.SerialPort != nil &&
			u.InputPort.Number != 0 && rs232.PortIsOpen("/dev/ttyS"+strconv.Itoa(u.InputPort.Number+1)) {
			currentText := u.InputEntry.Text
			for len(currentText)-len(prevText) >= 7 {
				dataChunk := currentText[len(prevText) : len(prevText)+7]
				rawPacket, formattedPacket, err := packet.SerializePacket(dataChunk, u.InputPort.Number)
				if err != nil {
					gui.ErrorWindow(err, u.App)
					return
				}
				csma_cd.Transmitter(rawPacket, formattedPacket, u)
				prevText += dataChunk
			}
		} else {
			if u.InputEntry.Text != "" && u.TransmittedBytes == 0 {
				u.InputEntry.Text = ""
				gui.ErrorWindow(errors.New("Open the both ports of the pair"), u.App)
			}
		}
		runtime.GC()
	}
}

func ReceiveData(u *gui.UserInterface) {
	for {
		if u.OutputEntry != nil && u.OutputPort.SerialPort != nil {
			data, err := csma_cd.Receiver(u.OutputPort)
			if err != nil && err.Error() != "Port has been closed" {
				gui.ErrorWindow(err, u.App)
				continue
			}
			u.OutputEntry.SetText(u.OutputEntry.Text + data)
		}
	}
}

func main() {
	u := new(gui.UserInterface)
	u.App = app.New()
	w := u.App.NewWindow("Serial port communication")
	u.App.Settings().SetTheme(&gui.СustomTheme{Theme: theme.DefaultTheme()})
	ports, err := serial.GetPortsList()
	if err != nil || len(ports) == 0 {
		gui.ErrorWindow(errors.New("No serial ports found"), u.App)
		time.Sleep(time.Minute)
		panic("No serial ports found!")
	}
	u.InputPort = new(rs232.Port)
	u.OutputPort = new(rs232.Port)
	u.TransmittedBytes = 0
	u.InitEntries()
	u.InitSelects(ports)
	u.UpdateStatus("")
	u.MakeGrid()
	w.SetContent(u.Grid)
	w.Resize(fyne.NewSize(675, 475))
	go func() {
		for {
			if u.InputPort.SerialPort == nil || u.OutputPort.SerialPort == nil {
				err := gui.UpdatePorts(u.SelectInputPort, u.SelectOutputPort)
				if err != nil {
					gui.ErrorWindow(err, u.App)
				}
			} else {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
	go TransmitData(u)
	go ReceiveData(u)
	w.ShowAndRun()
}
