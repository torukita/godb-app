.PHONY: build linux
GOOS ?= linux
GOARCH ?= amd64

build:
	go build

linux:
	GOOS=linux GOARCH=amd64 go build -o build/godb-app-linux main.go

