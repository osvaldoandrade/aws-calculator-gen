.PHONY: build run test fmt lint

BINARY := seidor-tools
CMD_DIR := ./cmd/$(BINARY)
ARGS ?=

build:
	@mkdir -p bin
	go build -o bin/$(BINARY) $(CMD_DIR)

run:
	go run $(CMD_DIR) $(ARGS)

test:
	go test ./...

fmt:
	go fmt ./...

lint:
	@out=$$(gofmt -l -s .); if [ -n "$$out" ]; then echo "$$out"; exit 1; fi