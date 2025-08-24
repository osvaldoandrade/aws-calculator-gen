.RECIPEPREFIX := >
.PHONY: build run test fmt lint

build:
>go build ./cmd/seidor-tools

run:
>go run ./cmd/seidor-tools $(ARGS)

test:
>go test ./...

fmt:
>gofmt -w $(shell find . -name *.go)

lint:
>gofmt -l $(shell find . -name *.go)
