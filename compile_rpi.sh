#!/bin/sh

export GOOS=linux
export GOARCH=arm
export GOARM=5

go build -o rpi_mon *.go

