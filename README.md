# Probe

> Pre-alpha: Probe is a work in progress and not yet an HTTP client.

Probe is a text-based API-testing language with an HTTP-like request format. It has a tree-sitter grammar and a Go command-line interface. The [language specification](docs/language-spec.md) describes the intended language; it is not a list of implemented CLI features.

## Get started

Build the CLI, create a request file, then parse it:

```sh
cd cli
make build
cat > example.probe <<'EOF'
GET https://example.com/ HTTP/1.1
Accept: text/html

EOF
./build/probe run example.probe
```

The command currently parses the file and prints its request target. It does not send the request.

## Current status

The grammar parses HTTP-style request blocks, headers, bodies, request separators, interpolation syntax, and several directives. The CLI reads and parses a `.probe` file, but request execution, response handling, directive evaluation, assertions, variable interpolation, and multipart upload construction are not implemented.

See the [language specification](docs/language-spec.md) for planned features such as multiple requests, `@let`, `@assert`, `@save`, and `@use`.

## Repository layout

| Path | Purpose |
| --- | --- |
| [cli/](cli/) | Go CLI (`probe`) |
| [tree-sitter-probe/](tree-sitter-probe/) | Tree-sitter grammar and bindings |
| [docs/language-spec.md](docs/language-spec.md) | Draft language specification |

The root [go.work](go.work) connects the CLI and the grammar's Go binding.

## Development

The workspace declares Go 1.26.5. Node.js and npm are needed for grammar development.

```sh
cd cli
make build
make test
make vet
make fmt
make tidy
```

To test the grammar corpus and Node binding:

```sh
cd tree-sitter-probe
npm install
npx tree-sitter test  # grammar corpus
npm test               # Node binding
```

## License

Apache License 2.0. See [LICENSE](LICENSE).
