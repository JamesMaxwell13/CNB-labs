package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"go.bug.st/serial"
	"lab_2/gui"
	"lab_2/packet"
	"lab_2/rs232"
	"sync"
	"time"
)

func TransmitData(u *gui.UserInterface) {
	prevText := ""
	for {
		if u.InputEntry != nil && u.InputPort.SerialPort != nil {
			currentText := u.InputEntry.Text
			if len(currentText)-len(prevText) == 7 {
				rawPacket, formattedPacket, err := packet.SerializePacket(
					currentText[len(prevText):],
					u.InputPort.Number,
				)
				if err != nil {
					gui.ErrorWindow(err, u.App)
				}
				err = u.InputPort.WriteBytes(rawPacket)
				if err != nil {
					gui.ErrorWindow(err, u.App)
				}
				u.TransmittedBytes += 7
				u.UpdateStatus(formattedPacket)
				prevText = currentText
			}
		}
	}
}

func ReceiveData(u *gui.UserInterface, mutex *sync.Mutex) {
	for {
		mutex.Lock()
		if u.OutputEntry != nil && u.OutputPort.SerialPort != nil {
			rawData, err := u.OutputPort.ReadBytes()
			mutex.Unlock()
			if err != nil && err.Error() != "Port has been closed" {
				gui.ErrorWindow(err, u.App)
				continue
			}
			if len(rawData) > 0 {
				data, errPacket := packet.DeserializePacket(rawData)
				if errPacket != nil {
					gui.ErrorWindow(errPacket, u.App)
					continue
				}
				u.OutputEntry.SetText(u.OutputEntry.Text + data)
			}
		} else {
			mutex.Unlock()
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
	mutex := new(sync.Mutex)
	go TransmitData(u)
	go ReceiveData(u, mutex)
	w.ShowAndRun()
}
