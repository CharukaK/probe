.PHONY: install generate build test vet run playground

install:
	$(MAKE) -C tree-sitter-probe install

generate:
	$(MAKE) -C tree-sitter-probe generate
	cd cli/internal/parser && go generate

build:
	$(MAKE) -C tree-sitter-probe build
	go build ./cli/...

test:
	$(MAKE) -C tree-sitter-probe test-grammar test-rust
	go test ./cli/... ./tree-sitter-probe/...

vet:
	go vet ./cli/... ./tree-sitter-probe/...

run:
	go run ./cli/cmd $(ARGS)

playground:
	$(MAKE) -C tree-sitter-probe playground
