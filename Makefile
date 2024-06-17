SHELL:=/bin/bash

.PHONY: run lint build

run:
	go run main.go

lint:
	go fmt ./internal/... && go fmt ./cmd/... && go fmt ./main.go

build:
	go build -o ./build/devops

build-arm:
	GOOS=linux GOARCH=arm64 go build -o ./build/devops-arm64

build-x86:
	GOOS=linux GOARCH=amd64 go build -o ./build/devops-x86_64
