.PHONY: build test

build:
	go build -o seidor-cloud ./cmd/seidor-cloud

test:
	go test ./...

