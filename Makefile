.PHONY: build test

build:
	go build ./cmd/seidor-aws-cli

test:
	go test ./... -coverprofile=coverage.out
