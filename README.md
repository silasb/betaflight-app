# Betaflight App

A quick Betaflight PID changer specifically for a tablet.

![Screen1](.images/screen1.png?raw=true "Screen")
![Screen2](.images/screen2.png?raw=true "Screen")

## Getting started

Getting `yarn` and `go` packages install:

    yarn install
    vgo get

Build the world:

    yarn build-amd64

Examine the `package.json` to understand the build commands.

## Todo

- [x] Auto update binary
- [ ] Auto refresh serial ports
- [ ] Test with esp8266
- [ ] Load different profiles
- [ ] Allow changing rates