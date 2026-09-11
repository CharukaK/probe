GO ?= go
BINARY := probe
CMD_DIR := ./cmd/probe

.PHONY: all build generate run test vet fmt tidy clean

all: build

generate:
	$(GO) generate ./internal/parser/...
	$(GO) fmt ./internal/parser/...

build: generate
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
