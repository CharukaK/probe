GO ?= go
BINARY := probe
CMD_DIR := ./cmd

.PHONY: all build run test vet fmt tidy clean

all: build

build:
	$(GO) build -o build/$(BINARY) $(CMD_DIR)

run: build
	./build/$(BINARY)

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

fmt:
	$(GO) fmt ./...

tidy:
	$(GO) mod tidy

clean:
	$(RM) build/$(BINARY)
