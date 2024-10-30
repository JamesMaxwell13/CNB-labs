package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"
	"go.bug.st/serial"
	"lab_1/gui"
	"lab_1/rs232"
	"sort"
	"sync"
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
					u.TransmittedBytes += len([]byte(string(newChar)))
					u.UpdateStatus()
				}
				prevText = currentText
			}
		}
	}
}

func ReceiveData(u *gui.UserInterface, mutex *sync.Mutex) {
	for {
		mutex.Lock()
		if u.OutputEntry != nil && u.OutputPort.SerialPort != nil {
			data, err := u.OutputPort.ReadBytes()
			mutex.Unlock()
			if err != nil && err.Error() != "Port has been closed" {
				gui.ErrorWindow(err, u.App)
				continue
			}
			if len(data) > 0 {
				u.OutputEntry.SetText(u.OutputEntry.Text + string(data))
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
	sort.Sort(rs232.ByNumericSuffix(ports))
	u.InputPort = new(rs232.Port)
	u.OutputPort = new(rs232.Port)
	u.TransmittedBytes = 0
	u.InitEntries()
	u.InitSelects(ports)
	u.UpdateStatus()
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
