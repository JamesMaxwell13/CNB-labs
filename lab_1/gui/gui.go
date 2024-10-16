package gui

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"image/color"
	"lab_1/rs232"
	"log"
	"strings"
)

type LogWriter struct {
	entry *widget.Entry
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	lw.entry.SetText(lw.entry.Text + string(p))
	return len(p), nil
}

type СustomTheme struct {
	fyne.Theme
}

func (c СustomTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameInputBorder || name == theme.ColorNameDisabled ||
		name == theme.ColorNameForeground {
		return color.White
	} else {
		return c.Theme.Color(name, variant)
	}
}

func ErrorWindow(err error, a fyne.App) {
	w := a.NewWindow("Error")
	ok := widget.NewButton("Ok", func() { w.Close() })
	ok.Resize(fyne.NewSize(100, 50))
	w.SetContent(container.NewVBox(
		container.NewCenter(
			widget.NewLabel(
				err.Error())), ok),
	)
	w.SetFixedSize(true)
	w.Show()
}

type UserInterface struct {
	App              fyne.App
	InputPort        *rs232.Port
	OutputPort       *rs232.Port
	TransmittedBytes int
	ReceivedBytes    int
	InputEntry       *widget.Entry
	OutputEntry      *widget.Entry
	StatusEntry      *widget.Entry
	DebugEntry       *widget.Entry
	SelectInputPort  *widget.Select
	SelectOutputPort *widget.Select
	Grid             *fyne.Container
}

func (u *UserInterface) InitSelects(ports []string) {
	u.SelectInputPort = widget.NewSelect(
		ports,
		func(s string) {
			if u.InputPort.SerialPort != nil {
				err := u.InputPort.ClosePort()
				if err != nil {
					ErrorWindow(err, u.App)
				}
			}
			_, err := u.InputPort.OpenPort(s)
			if err != nil {
				ErrorWindow(err, u.App)
			}
			err = UpdatePorts(u.SelectInputPort, u.SelectOutputPort, ports)
			if err != nil {
				ErrorWindow(err, u.App)
			}
		},
	)
	u.SelectInputPort.PlaceHolder = "Transmitter"
	u.SelectOutputPort = widget.NewSelect(
		ports,
		func(s string) {
			if u.OutputPort.SerialPort != nil {
				err := u.OutputPort.ClosePort()
				if err != nil {
					ErrorWindow(err, u.App)
				}
			}
			_, err := u.OutputPort.OpenPort(s)
			if err != nil {
				ErrorWindow(err, u.App)
			}
			err = UpdatePorts(u.SelectInputPort, u.SelectOutputPort, ports)
			if err != nil {
				ErrorWindow(err, u.App)
			}
		},
	)
	u.SelectOutputPort.PlaceHolder = "Receiver"
}

func (u *UserInterface) InitEntries() {
	u.InputEntry = widget.NewMultiLineEntry()
	u.InputEntry.PlaceHolder = "Write some text..."
	prevText := u.InputEntry.Text
	u.InputEntry.OnChanged = func(newText string) {
		if len(newText) < len(prevText) || !strings.HasPrefix(newText, prevText) {
			u.InputEntry.SetText(prevText)
			u.InputEntry.CursorRow = len(newText)
		} else {
			prevText = newText
		}
	}
	u.OutputEntry = InitReadOnlyEntry()
	u.StatusEntry = InitReadOnlyEntry()
	u.DebugEntry = InitReadOnlyEntry()
	log.SetFlags(log.Ltime)
	log.SetOutput(&LogWriter{entry: u.DebugEntry})
}

func InitReadOnlyEntry() *widget.Entry {
	entry := widget.NewMultiLineEntry()
	entry.SetText("")
	entry.Disable()
	return entry
}

func (u *UserInterface) MakeGrid() {
	statusBorder := container.NewBorder(
		container.NewCenter(widget.NewLabel("Status")),
		nil, nil, nil,
		u.StatusEntry,
	)
	debugBorder := container.NewBorder(
		container.NewCenter(widget.NewLabel("Debug")),
		nil, nil, nil,
		u.DebugEntry,
	)
	column1 := container.NewGridWithRows(2,
		statusBorder,
		debugBorder,
	)
	column2 := container.NewBorder(
		container.NewVBox(u.SelectInputPort,
			container.NewCenter(widget.NewLabel("Transmitted data"))),
		nil, nil, nil,
		u.InputEntry)
	column3 := container.NewBorder(
		container.NewVBox(u.SelectOutputPort,
			container.NewCenter(widget.NewLabel("Received data"))),
		nil, nil, nil,
		u.OutputEntry)
	u.Grid = container.New(
		layout.NewGridLayoutWithColumns(3),
		column1,
		column2,
		column3)
}

func (u *UserInterface) UpdateStatus() {
	mode := rs232.DefaultConfig()
	status := fmt.Sprintf("Ports parameters:\n"+
		"Baudrate - %d\nData bits - %d\nStop bits - %d\n"+
		"Parity bits - %d\nBytes transmitted - %d\nBytes raceived - %d", mode.BaudRate, mode.DataBits,
		int(mode.StopBits)+1, mode.Parity, u.TransmittedBytes, u.ReceivedBytes)
	u.StatusEntry.SetText(status)
}

func UpdatePorts(selectInputPort, selectOutputPort *widget.Select, allPorts []string) error {
	newPorts, err := rs232.RemovePorts(allPorts)
	if err != nil {
		return err
	}
	selectInputPort.Options = newPorts
	selectOutputPort.Options = newPorts
	selectInputPort.Refresh()
	selectOutputPort.Refresh()
	log.Printf("Ports list updated successful\n")
	return nil
}
