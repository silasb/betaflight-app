package main

import (
	"log"
	"strconv"
	"strings"

	"github.com/zserge/webview"
)

const (
	windowWidth  = 600
	windowHeight = 900
)

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
	FlightSurfaces map[string]*FlightSurface `json:"flightSurfaces"`
}

func (c *Betaflight) IncrPid(n int, flightSurface string, pid string) {
	pidStruct := c.FlightSurfaces[flightSurface].Pids[pid]
	pidStruct.Incr(n)
}

func (c *Betaflight) DecPid(n int, flightSurface string, pid string) {
	pidStruct := c.FlightSurfaces[flightSurface].Pids[pid]
	pidStruct.Dec(n)
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

func main() {
	w := webview.New(webview.Settings{
		Width:  windowWidth,
		Height: windowHeight,
		Title:  "Click counter: " + uiFrameworkName,
		ExternalInvokeCallback: handleRPC,
		Resizable:              true,
		Debug:                  true,
		URL:                    injectHTML(string(MustAsset("www/index.html"))),
		// URL:                    "data:text/html," + url.PathEscape(indexHTML),
	})

	defer w.Exit()

	betaFlight := &Betaflight{
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
		},
	}

	w.Dispatch(func() {
		// Inject controller
		w.Bind("betaflight", betaFlight)

		// Inject CSS
		w.InjectCSS(string(MustAsset("www/styles.css")))

		// Inject web UI framework and app UI code
		loadUIFramework(w)
	})
	w.Run()
}
