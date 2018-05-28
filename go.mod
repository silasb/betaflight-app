module github.com/silasb/betaflight-pid-app

replace github.com/fiam/msp-tool v0.0.0-20180218210101-e86acb5a412e => github.com/silasb/msp-tool v0.0.0-break-up-packages

replace github.com/bugst/go-serial v0.0.0-20170728081230-eae1344f9f90 => go.bug.st/serial.v1 v0.0.0-20170728081230-eae1344f9f90

require (
	github.com/bugst/go-serial v0.0.0-20170728081230-eae1344f9f90
	github.com/creack/goselect v0.0.0-20180501195510-58854f77ee8d
	github.com/fiam/msp-tool v0.0.0-20180218210101-e86acb5a412e
	github.com/tarm/serial v0.0.0-20180114052751-eaafced92e96
	github.com/zserge/webview v0.0.0-20180509070823-016c6ffd99f3
	go.bug.st/serial.v1 v0.0.0-20170728081230-eae1344f9f90
	golang.org/x/sys v0.0.0-20180522224204-88eb85aaee56
)
