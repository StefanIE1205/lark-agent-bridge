.PHONY: build test clean release run

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
LDFLAGS := -s -w -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o lab.exe ./cmd/lab/

test:
	go test ./...

test-verbose:
	go test ./... -v

run:
	go run ./cmd/lab/ --config config.toml

release: build
	@echo "Release build: lab.exe ($(VERSION))"
	@ls -la lab.exe

clean:
	rm -f lab.exe
	go clean -cache

deps:
	go mod tidy
	go mod verify
