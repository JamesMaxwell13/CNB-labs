package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"go.bug.st/serial"
	"lab_1/gui"
	"lab_1/rs232"
	"time"
)

func TransmitData(u *gui.UserInterface) {
	prevText := ""
	for {
		if u.InputEntry != nil && u.InputPort.SerialPort != nil {
			currentText := u.InputEntry.Text
			if len(currentText) > len(prevText) {
				newChars := []rune(currentText[len(prevText):])
				for _, newChar := range newChars {
					err := u.InputPort.WriteBytes([]byte(string(newChar)))
					if err != nil {
						gui.ErrorWindow(err, u.App)
					}
					u.TransmittedBytes++
					u.UpdateStatus()
				}
				prevText = currentText
			}
		}
	}
}

func ReceiveData(u *gui.UserInterface) {
	for {
		if u.OutputEntry != nil && u.OutputPort.SerialPort != nil {
			data, err := u.OutputPort.ReadBytes()
			if err != nil {
				if err.Error() != "Port has been closed" {
					gui.ErrorWindow(err, u.App)
				}
				continue
			}
			if len(data) > 0 {
				u.OutputEntry.SetText(u.OutputEntry.Text + string(data))
				u.ReceivedBytes++
				u.UpdateStatus()
			}
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
		gui.ErrorWindow(errors.New("No rs232 ports found"), u.App)
		time.Sleep(time.Minute)
		panic("No rs232 ports found!")
	}
	u.InputPort = new(rs232.Port)
	u.OutputPort = new(rs232.Port)
	u.TransmittedBytes = 0
	u.ReceivedBytes = 0
	u.InitEntries()
	u.InitSelects(ports)
	u.UpdateStatus()
	u.MakeGrid()
	w.SetContent(u.Grid)
	w.Resize(fyne.NewSize(600, 450))
	go func() {
		for {
			if u.InputPort.SerialPort == nil || u.OutputPort.SerialPort == nil {
				err := gui.UpdatePorts(u.SelectInputPort, u.SelectOutputPort, ports)
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
