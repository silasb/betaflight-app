package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	// serial "github.com/bugst/go-serial"

	serial "github.com/bugst/go-serial"

	_fc "github.com/fiam/msp-tool/fc"
	"github.com/zserge/webview"
)

const (
	windowWidth  = 600
	windowHeight = 900
)

var VERSION string
var fc *_fc.FC
var betaFlight *Betaflight
var sync func()

var w webview.WebView

type Pid struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func (p *Pid) Incr(n int) {
	p.Value = p.Value + int(n)
}
func (p *Pid) Dec(n int) {
	p.Value = p.Value - int(n)
}

type FlightSurface struct {
	Name string          `json:"name"`
	Pids map[string]*Pid `json:"pids"`
}

type Betaflight struct {
	Flash                string                    `json:"flash"`
	SerialPortsAvailable []string                  `json:"serialPortsAvailable"`
	ConnectedSerialPort  string                    `json:"connectedSerialPort"`
	FlightSurfaces       map[string]*FlightSurface `json:"flightSurfaces"`
}

func (c *Betaflight) SetFlash(value string) {
	c.Flash = value
}

func (c *Betaflight) IncrPid(n int, flightSurface string, pid string) {
	pidStruct := c.FlightSurfaces[flightSurface].Pids[pid]
	pidStruct.Incr(n)
}

func (c *Betaflight) DecPid(n int, flightSurface string, pid string) {
	pidStruct := c.FlightSurfaces[flightSurface].Pids[pid]
	pidStruct.Dec(n)
}

func (c *Betaflight) SavePids() {
	fc.SetPIDs(convertLocalPidsToFCPids(betaFlight.FlightSurfaces))

	c.Flash = "PIDs saved!"
}

func convertLocalPidsToFCPids(flightSurfaces map[string]*FlightSurface) []uint8 {

	output := make([]uint8, 8)

	for _, flightSurface := range []string{"yaw", "pitch", "roll"} {
		fs := flightSurfaces[flightSurface]

		var loopOver []string

		if len(fs.Pids) == 3 {
			loopOver = []string{"d", "i", "p"}
		} else if len(fs.Pids) == 2 {
			loopOver = []string{"i", "p"}
		}

		for _, pid := range loopOver {
			p := fs.Pids[pid]

			output = append([]uint8{uint8(p.Value)}, output...)
		}
	}

	return output
}

func (c *Betaflight) Connect(serialPort string) {
	var err error

	var pidCb MyPIDReceiver

	opts := _fc.FCOptions{
		PortName: serialPort,
		BaudRate: 115200,
		// Stdout:           km,
		EnableDebugTrace: true,
	}
	fc, err = _fc.NewFC(opts)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		fc.StartUpdating(pidCb)
	}()

	c.ConnectedSerialPort = serialPort
	c.Flash = "Connected!"

	fc.GetPIDs()

}

func handleRPC(w webview.WebView, data string) {
	switch {
	case data == "close":
		w.Terminate()
	case data == "fullscreen":
		w.SetFullscreen(true)
	case data == "unfullscreen":
		w.SetFullscreen(false)
	case data == "open":
		log.Println("open", w.Dialog(webview.DialogTypeOpen, 0, "Open file", ""))
	case data == "opendir":
		log.Println("open", w.Dialog(webview.DialogTypeOpen, webview.DialogFlagDirectory, "Open directory", ""))
	case data == "save":
		log.Println("save", w.Dialog(webview.DialogTypeSave, 0, "Save file", ""))
	case data == "message":
		w.Dialog(webview.DialogTypeAlert, 0, "Hello", "Hello, world!")
	case data == "info":
		w.Dialog(webview.DialogTypeAlert, webview.DialogFlagInfo, "Hello", "Hello, info!")
	case data == "warning":
		w.Dialog(webview.DialogTypeAlert, webview.DialogFlagWarning, "Hello", "Hello, warning!")
	case data == "error":
		w.Dialog(webview.DialogTypeAlert, webview.DialogFlagError, "Hello", "Hello, error!")
	case strings.HasPrefix(data, "changeTitle:"):
		w.SetTitle(strings.TrimPrefix(data, "changeTitle:"))
	case strings.HasPrefix(data, "changeColor:"):
		hex := strings.TrimPrefix(strings.TrimPrefix(data, "changeColor:"), "#")
		num := len(hex) / 2
		if !(num == 3 || num == 4) {
			log.Println("Color must be RRGGBB or RRGGBBAA")
			return
		}
		i, err := strconv.ParseUint(hex, 16, 64)
		if err != nil {
			log.Println(err)
			return
		}
		if num == 3 {
			r := uint8((i >> 16) & 0xFF)
			g := uint8((i >> 8) & 0xFF)
			b := uint8(i & 0xFF)
			w.SetColor(r, g, b, 255)
			return
		}
		if num == 4 {
			r := uint8((i >> 24) & 0xFF)
			g := uint8((i >> 16) & 0xFF)
			b := uint8((i >> 8) & 0xFF)
			a := uint8(i & 0xFF)
			w.SetColor(r, g, b, a)
			return
		}
	}
}

type MyPIDReceiver struct {
}

func (p MyPIDReceiver) ReceivedPID(pids map[string]*_fc.Pid) error {

	interestingPids := map[string]bool{"roll": true, "pitch": true, "yaw": true}

	for flightSurface, pid := range pids {
		if !interestingPids[flightSurface] {
			continue
		}

		if len(pid.Value) == 3 {
			for i, p := range []string{"p", "i", "d"} {
				betaFlight.FlightSurfaces[flightSurface].Pids[p].Value = int(pid.Value[i])
			}
		} else if len(pid.Value) == 2 {
			for i, p := range []string{"p", "i"} {
				betaFlight.FlightSurfaces[flightSurface].Pids[p].Value = int(pid.Value[i])
			}
		}
	}

	fmt.Println("finished updating pids")

	w.Dispatch(func() {
		sync()
	})

	return nil
}

func main() {
	ports, err := serial.GetPortsList()
	if err != nil {
		log.Fatal(err)
	}

	CheckForNewUpdates(VERSION)

	betaFlight = &Betaflight{
		SerialPortsAvailable: ports,
		// ConnectedSerialPort:  nil,
		FlightSurfaces: map[string]*FlightSurface{
			"roll": &FlightSurface{
				Name: "Roll",
				Pids: map[string]*Pid{
					"p": &Pid{
						Name:  "P",
						Value: 0,
					},
					"i": &Pid{
						Name:  "I",
						Value: 0,
					},
					"d": &Pid{
						Name:  "D",
						Value: 0,
					},
				},
			},
			"pitch": &FlightSurface{
				Name: "Pitch",
				Pids: map[string]*Pid{
					"p": &Pid{
						Name:  "P",
						Value: 0,
					},
					"i": &Pid{
						Name:  "I",
						Value: 0,
					},
					"d": &Pid{
						Name:  "D",
						Value: 0,
					},
				},
			},
			"yaw": &FlightSurface{
				Name: "Yaw",
				Pids: map[string]*Pid{
					"p": &Pid{
						Name:  "P",
						Value: 0,
					},
					"i": &Pid{
						Name:  "I",
						Value: 0,
					},
				},
			},
		},
	}

	w = webview.New(webview.Settings{
		Width:  windowWidth,
		Height: windowHeight,
		Title:  fmt.Sprintf("Betaflight PID App %s", VERSION),
		ExternalInvokeCallback: handleRPC,
		Resizable:              true,
		Debug:                  true,
		URL:                    injectHTML(string(MustAsset("www/index.html"))),
		// URL:                    "data:text/html," + url.PathEscape(indexHTML),
	})

	defer w.Exit()
	w.Dispatch(func() {
		var err error
		// Inject controller
		sync, err = w.Bind("betaflight", betaFlight)
		if err != nil {
			panic(err)
		}

		// Inject CSS
		w.InjectCSS(string(MustAsset("www/styles.css")))

		// Inject web UI framework and app UI code
		loadUIFramework(w)
	})

	w.Run()
}
