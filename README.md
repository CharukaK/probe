# Probe

![status: pre-alpha](https://img.shields.io/badge/status-pre--alpha-lightgrey)

**A text-based API testing tool built on HTTP message syntax.**

Requests are written as plain `.probe` files, in the same shape as a raw
HTTP message: a request line, headers, and a body. The CLI parses one,
sends it over the wire, and prints the response.

[`docs/language-spec.md`](docs/language-spec.md) describes the full
design: variable interpolation, assertions, multi-request files, project
config, and more. This README covers only what's implemented today; the
spec is about where the language is headed, not what runs yet.

## Status

Pre-alpha. The grammar and CLI currently handle a single HTTP request end
to end; most of the DSL described in the language spec isn't wired up.

**Works today:**

- Parsing one request per `.probe` file: method, request target, optional
  HTTP version, headers, and a raw body.
- `probe run <file>.probe`: parses the file, sends the request over real
  HTTP (`net/http`), and prints the response body.
- `-v` / `--verbose`: prints the raw request and response, wire format, as
  they cross the transport.

**Designed, not built yet:**

- Multiple requests per file (`###`-separated): the grammar currently
  parses one request block per file.
- `{{variable}}` interpolation: braces aren't special yet.
- `@assert`, `@save`, `@use`, and `probe.toml` project config.
- Sending an actual `multipart/form-data` body: `@field` / `@file` lines
  parse into structured data, but the CLI doesn't build a request body from
  them yet, so a file using them currently sends an empty body.

## Repo layout

Two independent packages:

| Path | What it is |
|---|---|
| [`tree-sitter-probe/`](tree-sitter-probe/) | The grammar: Node tooling, multi-language bindings |
| [`cli/`](cli/) | The Go CLI (`probe`), built on the grammar's Go binding |

A root `go.work` ties `cli/` and `tree-sitter-probe/`'s Go binding together.

## Building

### Requirements

- Go 1.26+, to build and run the CLI
- Node.js + npm, only if you're changing the grammar itself

### CLI

```sh
cd cli
make build          # -> cli/build/probe
```

(equivalent to `go build -o build/probe ./cmd/probe`)

### Grammar

```sh
cd tree-sitter-probe
npm install
tree-sitter generate   # regenerate the parser from grammar.js
tree-sitter test       # run the corpus in test/corpus/*.txt
```

## Examples

### GET request

`cli/public.probe`:

```http
GET https://www.alphavantage.co/query?function=TIME_SERIES_MONTHLY&symbol=IBM&apikey=demo HTTP/1.1
Accept: application/json
```

```sh
cd cli
./build/probe run public.probe
```

Sends the request and prints the response body. This is a real call to
Alpha Vantage's public demo endpoint.

### POST with a JSON body, verbose

```http
POST https://httpbin.org/post HTTP/1.1
Content-Type: application/json

{"name": "probe", "kind": "api test"}
```

```sh
./build/probe run post-example.probe -v
```

`-v` prints the request and response as they're sent:

```text
> POST /post HTTP/1.1
> Host: httpbin.org
> Content-Type: application/json
>
> {"name": "probe", "kind": "api test"}

< HTTP/2.0 200 OK
< Content-Type: application/json
< ...
```

The response body follows on stdout, same as the non-verbose run.

## License

MIT, per `tree-sitter-probe/package.json`. *(No top-level `LICENSE` file yet.)*
