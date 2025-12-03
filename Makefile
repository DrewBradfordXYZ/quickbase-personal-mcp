.PHONY: build run clean test

build:
	go build -o quickbase-personal-mcp

run:
	go run main.go

clean:
	rm -f quickbase-personal-mcp

test:
	go test -v ./...

install-deps:
	go mod download

.DEFAULT_GOAL := build
