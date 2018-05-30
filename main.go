package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	serial "go.bug.st/serial.v1"
	//serial "github.com/bugst/go-serial"

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
var ticker *time.Ticker

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
	UpdatingApp          bool                      `json:"updating_app"`
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

	ticker.Stop()

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

func (c *Betaflight) Disconnect() {
	err := fc.Close()
	if err != nil {
		c.Flash = "Could not disconnect serial port"
		return
	}

	// restart serial port ticker
	ticker = time.NewTicker(time.Second)
	go watchSerialPorts(ticker)

	c.ConnectedSerialPort = ""
	c.Flash = "Serial port disconnected"
}


func (c *Betaflight) ExportPids(path string) error {
	b, _ := json.MarshalIndent(betaFlight.FlightSurfaces, "", "  ")
	b = append(b, '\n')

	err := ioutil.WriteFile(path, b, 0644)

	fmt.Printf("Wrote %+v\n", path)

	return err
}

func (c *Betaflight) ImportPids(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &betaFlight.FlightSurfaces)

	sync()

	fmt.Printf("Read %+v\n", path)

	return err
}

func watchSerialPorts(ticker *time.Ticker) {
	for range ticker.C {
		ports, err := serial.GetPortsList()
		if err != nil {
			log.Fatal(err)
		}
		betaFlight.SerialPortsAvailable = ports

		w.Dispatch(func() {
			sync()
		})
	}
}

func handleRPC(w webview.WebView, data string) {
	switch {
	case data == "close":
		w.Terminate()
	case data == "fullscreen":
		w.SetFullscreen(true)
	case data == "unfullscreen":
		w.SetFullscreen(false)
	case data == "load":
		path := w.Dialog(webview.DialogTypeOpen, 0, "Open file", "")
		err := betaFlight.ImportPids(path)
		if err != nil {
			log.Println("Error", err)
		}
	case data == "opendir":
		log.Println("open", w.Dialog(webview.DialogTypeOpen, webview.DialogFlagDirectory, "Open directory", ""))
	case data == "dump":
		path := w.Dialog(webview.DialogTypeSave, 0, "Save file", "")
		err := betaFlight.ExportPids(path)
		if err != nil {
			log.Println("Error", err)
		}
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
	var ports []string

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

	if ok, version := HasNewerVersion(VERSION); ok {
		betaFlight.UpdatingApp = true
		go func() {
			err := UpdateBinary(version)
			if err != nil {
				panic(err)
			}
		}()
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

	// ticker used for watching the serial ports
	ticker = time.NewTicker(time.Second)
	go watchSerialPorts(ticker)

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
