.PHONY: install generate build test test-grammar test-go test-rust playground

install:
	npm ci

generate:
	./node_modules/.bin/tree-sitter generate

build:
	npm run build

test: test-grammar test-go test-rust

test-grammar:
	./node_modules/.bin/tree-sitter test

test-go:
	go test ./...

test-rust:
	cargo test

playground:
	npm start
