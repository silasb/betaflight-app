# Betaflight PID App

A Betaflight PID changer via a tablet

## Getting started

    yarn install

Building `hyperapp.js`

    rollup -i .\www\vendor\hyperapp.js -o www/vendor/hyperapp2.min.js -m -f umd -n hyperapp

Building the JS on demand:

    yarn run watch

Go stuff
:g
    go generate -tags js
    vgo build -tags js