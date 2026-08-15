# 📐 Probe Language Specification (v0.3 — Draft)

*Companion to the [PRD](./index.md) and the [Roadmap](./roadmap.md). This is
the formal reference for the `.probe` file format: both its base HTTP syntax
and its Probe-specific extensions.*

Status: **Draft, v1 scope only.** Not yet implemented. Control flow
(loops/conditionals) is explicitly deferred — see §11.

**v0.2 changes**: added built-in functions to interpolation (§6), multipart
/ file-upload bodies (§4.4), teardown blocks (§5.1), expanded `@assert`
comparators for length/existence/schema (§7.1), and the `probe.toml` project
configuration file (§13). These were identified as grammar-affecting gaps
that needed resolving before the Probe-extensions layer could be finalized —
see [roadmap §7](./roadmap.md#7-future-considerations-not-yet-scoped) for the
broader list of deferred, non-grammar items tracked alongside these.

**v0.3 changes**: `@use ... as <alias>` (§7.3) — namespaced imports, closing
a variable-shadowing hole in the flat import model v0.1–v0.2 shipped with
(two `@use`d dependencies, or a dependency and its importer, saving the same
identifier would silently overwrite one another with no error).

## 1. Notation

Grammar is given in ABNF ([RFC 5234](https://www.rfc-editor.org/rfc/rfc5234)),
the same notation [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110) uses.
Core rules (`ALPHA`, `DIGIT`, `VCHAR`, `SP`, `HTAB`, `CRLF`, `OWS`) are as
defined in RFC 5234 Appendix B.1 and RFC 9110 §5.6.3. Where a production is
identical to one in RFC 9110, it's referenced by name rather than
copied — e.g. `method`, `field-name`, `field-value` from RFC 9110 §9, §5.1,
§5.5.

```abnf
line-ending = CRLF / LF   ; source files may use either; wire requests are
                          ; normalized to CRLF by the implementation
```

## 2. File Structure Overview

```abnf
probe-file    = *blank-line [ *comment-line ] *use-directive request-block
                *( *blank-line separator *blank-line request-block ) *blank-line
request-block = *comment-line http-request *directive-line [ teardown-block ]
blank-line    = line-ending   ; an empty line — carries no meaning here
```

A `.probe` file is one or more **request blocks**, each a plain RFC 9110
HTTP request optionally preceded by comments and followed by Probe
directive lines, separated by an explicit `###` delimiter. Blank lines
around a `separator` are optional and insignificant — see §5. A file may
also declare dependencies on other files via leading `@use` directives
(§7.3) before its first request-block. A request-block may additionally
carry its own **teardown block** (§5.1), run after it regardless of outcome.

## 3. Comments

```abnf
comment-line = "#" *(VCHAR / SP / HTAB) line-ending
```

- A comment is a line whose first non-whitespace character is `#`.
- Comments are only valid **before the request-line** of a block (leading
  comments) — not inside the header block or body, to avoid ambiguity with
  RFC 9110 header-field syntax (which has no comment production). A `#` line
  found after the request-line but before the blank line that ends headers
  is a syntax error.

## 4. Request Blocks

### 4.1 Request Line

```abnf
request-line   = method SP request-target [ SP HTTP-version ] line-ending
method         = token           ; RFC 9110 §9.1 — any token, not just the
                                  ; standard verbs, so extension methods parse
request-target = absolute-form / interpolated-origin-form
```

- `HTTP-version` is **optional**, unlike raw RFC 9110 (which requires it on
  the wire). If omitted, the implementation assumes `HTTP/1.1`. This is the
  one deliberate deviation from RFC 9110 syntax — Probe files are a DSL over
  HTTP, not a literal wire capture.
- `request-target` is expected to be `absolute-form` (a full URL, RFC 9110
  §7.1) in the common case, since base hosts are typically supplied via
  `{{variable}}` interpolation (§6) rather than a separate `Host` line, e.g.
  `{{baseUrl}}/login`.

### 4.2 Header Fields

```abnf
header-block = *( field-line line-ending )
field-line   = field-name ":" OWS field-value OWS   ; RFC 9110 §5.1, §5.5
```

Standard RFC 9110 header-field syntax, unmodified. `field-value` may contain
`{{variable}}` interpolation tokens (§6), resolved before the request is
sent — the grammar treats them as opaque text; substitution is a semantic
(execution-time) concern, not a parse-time one.

### 4.3 Message Body

```abnf
http-request = request-line header-block line-ending [ message-body ]
message-body = *OCTET   ; to end of block (next separator, directive-line, or EOF)
```

Standard RFC 9110 framing: a blank line ends the header block; everything
after it, up to the next `@`-directive line, `###` separator, or EOF, is the
body. Bodies may contain `{{variable}}` interpolation tokens.

### 4.4 Multipart & File-Upload Bodies

```abnf
multipart-body = 1*( multipart-part line-ending )
multipart-part = field-part / file-part
field-part     = "@field" SP identifier SP "=" SP ( string / interpolation )
file-part      = "@file" SP identifier SP "=" SP quoted-path
                  [ SP "as" SP string ]
```

- If a request-block's headers include `Content-Type: multipart/form-data`
  (with or without an explicit `boundary`; the implementation generates one),
  `message-body` (§4.3) is parsed as a `multipart-body` instead of opaque
  octets: a sequence of `@field` / `@file` lines, one part per line.
- `@field` declares a text form field; its value follows the same
  string/interpolation rules as an `@assert` value (§7.1).
- `@file` declares a file part: `quoted-path` (§7.3) is resolved relative to
  the `.probe` file, read at execution time, and sent as the named part's
  content. The optional `as <string>` clause sets that part's
  `Content-Type` explicitly; if omitted, it's inferred from the file
  extension.
- A request whose `Content-Type` is not `multipart/form-data` keeps the
  existing `message-body = *OCTET` framing (§4.3) — `@field`/`@file` lines
  are only meaningful in a multipart body and are a syntax error elsewhere.

```
POST {{baseUrl}}/upload
Content-Type: multipart/form-data
Authorization: Bearer {{token}}

@field description = "profile picture"
@file avatar = "./fixtures/avatar.png" as "image/png"

@assert status == 201
```

## 5. Request Separators

```abnf
separator = "###" [ 1*SP request-name ] line-ending
request-name = 1*(VCHAR / SP)
```

- `###` on its own line starts a new request block. Text after `###` on the
  same line is that request's name (used in CLI/report output and as a
  reference target for future chaining features).
- The **first** request block in a file has no leading `###` — the
  separator is only needed *between* requests.
- **No blank line is required** on either side of a `separator`. Both of
  these are valid and equivalent:

  ```
  @assert status == 200
  ### Create user
  POST {{baseUrl}}/users
  ```

  ```
  @assert status == 200

  ### Create user

  POST {{baseUrl}}/users
  ```

  Blank lines there are purely cosmetic (`blank-line` in §2) — write them or
  not. The **one** blank line that is never optional is the one ending a
  request's header block, per RFC 9110 framing (§4.3) — that's what
  separates headers from the body, not what separates requests.
- This supersedes the bare back-to-back example shown in the PRD (§7.2) —
  that example predates this spec and should be read with an implicit
  `###` between the two requests.

### 5.1 Teardown Blocks

```abnf
teardown-block = teardown-sep *blank-line http-request *directive-line
teardown-sep   = "~~~" [ 1*SP request-name ] line-ending
```

- `~~~` introduces a **teardown block**: a request (with its own
  `@assert`/`@save` directives) that belongs to the request-block
  immediately above it, rather than starting an independent test. It shares
  the same variable scope, so it can reference values the owning block
  `@save`d (e.g. an id to delete).
- A teardown block runs once, immediately after its owning request-block
  finishes — **regardless of whether that block's `@assert`s passed or
  failed** — as long as the owning request was actually sent (an unresolved
  `{{variable}}` that prevented sending skips the teardown too; there's
  nothing to tear down).
- A teardown block's own `@assert` failures are reported as a **warning
  attached to the owning request**, not as a separate failed test — a
  cleanup hiccup (e.g. the resource was already gone) shouldn't flip an
  otherwise-passing test to failed, but it must still be visible.
- A request-block has at most one teardown block. Chaining cleanup steps is
  out of scope for v1 — see §11.

```
POST {{baseUrl}}/users
Content-Type: application/json

{"name": "New User", "email": "new@example.com"}

@assert status == 201
@save userId = body.id

~~~ cleanup
DELETE {{baseUrl}}/users/{{userId}}
Authorization: Bearer {{token}}

@assert status == 204
```

## 6. Variables & Interpolation

```abnf
interpolation = "{{" *SP interp-expr *SP "}}"
interp-expr   = identifier / function-call
identifier    = ALPHA *( ALPHA / DIGIT / "_" / "." )
function-call = fn-name "(" [ fn-arg *( *SP "," *SP fn-arg ) ] ")"
fn-name       = "uuid" / "now" / "base64" / "hmac_sha256"
fn-arg        = identifier / string / number
```

`{{identifier}}` may appear anywhere in a request-target, header field-value,
or message body. Dots in an `identifier` carry meaning in exactly one place:
the `alias.name` form produced by an aliased `@use ... as alias` (§7.3).
Everywhere else — including a bare identifier that happens to contain a dot
because you named it that way — dots are just part of the name; there's no
general dotted-path lookup into a variable's value (that's what `body.x.y`
`assert-target` dot notation, §7.1, is for, and it's a separate grammar).

`{{function-call()}}` computes a value at execution time instead of looking
one up; arguments may themselves be `identifier`s (resolved per the order
below before the function runs), not nested function calls — v1 keeps
function calls non-composable to avoid needing a general expression grammar.
Built-in v1 functions:

| Function | Args | Produces |
|---|---|---|
| `uuid()` | none | A random UUID v4 string |
| `now()` / `now(format)` | optional format `string` | Current timestamp; ISO 8601 if `format` omitted |
| `base64(value)` | `string` or identifier | Base64-encoded `value` |
| `hmac_sha256(key, value)` | two `string`s or identifiers | Hex-encoded HMAC-SHA256, for signed-request headers |

This set is deliberately small — covers the recurring "idempotency key /
timestamp header / signed request" cases named in
[roadmap §7](./roadmap.md#7-future-considerations-not-yet-scoped). Additional
functions are additive (new `fn-name` alternatives), not breaking changes.

**Resolution order** (first match wins):
1. Values captured by a prior `@save` in the same run (§7.2), resolved
   through the *resolving file's own `@use` view* (§7.3): its own
   request-blocks' saves, plus its dependencies' saves — either flattened
   in (unaliased `@use`) or reachable only as `alias.name` (aliased
   `@use ... as alias`). This view is assembled per file, not one
   indiscriminate global table — see §7.3 for exactly how.
2. Values from the `--env <file>` loaded at CLI startup
3. The active `[env.<name>]` table in `probe.toml` (§13), if one was
   selected and a project config file was found
4. Process environment variables

`function-call`s are evaluated fresh each time they're encountered — they
don't participate in this lookup order themselves, only their arguments do.

An unresolved `{{identifier}}`, or a `function-call` with an unresolvable
argument, at execution time is a runtime error (not a parse error) that
fails the containing request.

## 7. Directives

```abnf
directive-line = ( assert-directive / save-directive ) line-ending
```

Directive lines follow the request's body (or its blank-line terminator, if
there's no body) and precede the next `separator` or EOF. `@assert` and
`@save` (§7.1, §7.2) are **request-scoped**: they attach to the request
immediately above them. `@use` (§7.3) is **file-scoped**: it's a leading
directive that appears before any request-block (§2) and declares a
dependency on another file, rather than describing a single request.

### 7.1 `@assert`

```abnf
assert-directive  = value-assert / length-assert / existence-assert / schema-assert
value-assert      = "@assert" SP assert-target SP comparator SP assert-value
length-assert     = "@assert" SP assert-target SP "length" SP comparator SP number
existence-assert  = "@assert" SP assert-target SP ( "exists" / "not" SP "exists" )
schema-assert     = "@assert" SP "body" SP "matches" SP "schema" SP quoted-path

assert-target     = "status" / "duration" / header-target / body-target
header-target     = "headers." field-name
body-target       = "body" *( "." json-key / "[" 1*DIGIT "]" )
json-key          = 1*( ALPHA / DIGIT / "_" )
comparator        = "==" / "!=" / "<=" / ">=" / "<" / ">" / "contains" / "matches"
assert-value      = json-literal / interpolation
json-literal      = "null" / "true" / "false" / number / string
```

- `body` targets use dot notation for object keys and bracket notation for
  array indices, e.g. `body.user.id`, `body.items[0].name`.
- `matches` (in `value-assert`) compares against a regular expression
  `string` value.
- `length-assert` applies `comparator` to the element count of an array or
  object `assert-target`, or the character count of a string target — e.g.
  `@assert body.items length == 3`, `@assert body.items length >= 1`. It's
  a syntax error against `status`/`duration` (scalars have no length).
- `existence-assert` checks whether `assert-target` is present in the
  response at all — distinct from `value-assert ... != null`, which requires
  the key to be present *and* non-null. `@assert body.deletedAt not exists`
  passes if the key is absent; `@assert body.deletedAt != null` would error
  (unresolvable target) if the key is absent rather than just `null`.
- `schema-assert` validates the entire response body against a JSON Schema
  file at `quoted-path` (§7.3), resolved relative to the `.probe` file.
  This is the one `@assert` form that doesn't take an `assert-target`
  narrower than the whole `body` — it's a contract-level check, not a
  field-level one.
- A failed `@assert` (any of the four forms) fails the enclosing request;
  execution continues with the next request block (per PRD §5.3).

Examples:
```
@assert status == 200
@assert duration < 500
@assert headers.Content-Type == "application/json"
@assert body.id != null
@assert body.items[0].name matches "^user-"
@assert body.items length == 3
@assert body.user.email exists
@assert body.deletedAt not exists
@assert body matches schema "./schemas/user.schema.json"
```

### 7.2 `@save`

```abnf
save-directive = "@save" SP identifier SP "=" SP assert-target
```

Captures a value from the response into `identifier`, available to later
requests via `{{identifier}}` interpolation (§6) for the rest of the run.

```
@save token = body.token
@save requestId = headers.X-Request-Id
```

### 7.3 `@use`

```abnf
use-directive      = "@use" SP quoted-path [ SP "as" SP simple-identifier ] line-ending
quoted-path        = DQUOTE *(%x20-21 / %x23-7E) DQUOTE   ; any char but DQUOTE
simple-identifier  = ALPHA *( ALPHA / DIGIT / "_" )        ; no dots — see below
```

Declares that this file depends on another `.probe` file, which must run to
completion before this file's own request-blocks do. `quoted-path` is
resolved relative to the file containing the directive. A file may have
multiple `@use` lines; dependencies run in declaration order.

```
@use "./login.probe"

POST {{baseUrl}}/users
Authorization: Bearer {{token}}
...
```

**Execution semantics** (this is a run-time concern, not just syntax — it
changes what running a file means):

- Before a file's own request-blocks execute, each of its `@use`d files
  runs first, in declaration order. `@use` is transitive: if a dependency
  itself has `@use` lines, those resolve depth-first before it runs.
- A given file, resolved to an absolute path, executes **at most once per
  `probe run`**, no matter how many other files `@use` it. If both
  `users.probe` and `orders.probe` `@use "./login.probe"` in the same
  invocation, `login.probe` runs once; both reuse its captured `@save`
  results (each through its own view, per below — aliasing them
  differently, or not at all, doesn't force a second run).
- A dependency cycle (`A` uses `B` uses `A`) is a hard error at resolution
  time, reported with the offending cycle path. This formalizes the
  "circular dependencies" edge case already named in the PRD (§5.3).
- If a dependency's requests or assertions fail, every file that
  transitively depends on it fails too, attributed to the dependency —
  dependents are not run against a broken fixture. How this is *reported*
  in CLI output is a test-runner (M3) concern, not specified here.

**Namespacing (`as alias`) and variable visibility:**

Each file builds its own **view** of `{{identifier}}`-resolvable names —
resolution-order tier 1 (§6) — out of two sources: its own request-blocks'
`@save`s, and its `@use`d dependencies' `@save`s, combined per directive:

- **`@use "path" as alias`** (namespaced): the dependency's saved values are
  reachable *only* as `{{alias.name}}`, never as bare `{{name}}`. This is
  the only place in the grammar a dotted `identifier` is meaningful (§6).
  Namespacing is local to the importing file — it doesn't rename anything
  inside the dependency itself, and it doesn't propagate: if that dependency
  has its own `@use`s, *their* names aren't re-exposed through `alias.*`
  unless the dependency chooses to re-export them (not a v1 concept — see
  §11).
- **`@use "path"`** (unaliased, the v1-original form): the dependency's
  saved values flatten directly into the importing file's own flat
  namespace, alongside its own `@save`s — unchanged from pre-v0.3 behavior.
- **Collision rule for flattened (unaliased) names**: if two different
  `@use`d files — neither one a dependency of the other — would both
  contribute the same flat name to the same importing file's view, that's a
  **hard error at resolution time**, naming both source files and the
  conflicting identifier. Fix it by aliasing at least one of them.
  Rebinding is *not* an error when it happens within one dependency chain —
  a file's own `@save` may always override a flat name it imported (the
  importer's own value wins), and a later request in the same file may
  freely re-`@save` a name it (or its flat imports) already bound — e.g. a
  token-refresh request re-saving `token` after login is normal and allowed.
- **Alias collisions**: two `@use` lines in the same file may not declare
  the same `alias` — that's a syntax-adjacent error caught at resolution
  time (same phase as the flat-name collision above).

```
# orders.probe — same dependency imported twice, safely, via aliasing
@use "./login-as-customer.probe" as customer
@use "./login-as-admin.probe" as admin

DELETE {{baseUrl}}/orders/{{orderId}}
Authorization: Bearer {{admin.token}}

@assert status == 204
```

Without the `as` clauses here, both dependencies would `@save token`, and
`{{token}}` downstream would silently resolve to whichever one ran last —
exactly the shadowing hazard aliasing exists to close.

- Selectively importing only *some* of a dependency's variables (as opposed
  to all of them, whether flattened or under one alias) is still deferred —
  see §11.

## 8. Consolidated Grammar (Appendix)

```abnf
probe-file        = *blank-line [ *comment-line ] *use-directive request-block
                    *( *blank-line separator *blank-line request-block ) *blank-line
request-block     = *comment-line http-request *directive-line [ teardown-block ]
blank-line        = line-ending
separator         = "###" [ 1*SP request-name ] line-ending
request-name      = 1*(VCHAR / SP)
teardown-block    = teardown-sep *blank-line http-request *directive-line
teardown-sep      = "~~~" [ 1*SP request-name ] line-ending

http-request      = request-line header-block line-ending [ message-body ]
request-line      = method SP request-target [ SP HTTP-version ] line-ending
method            = token
request-target    = absolute-form / interpolated-origin-form
header-block      = *( field-line line-ending )
field-line        = field-name ":" OWS field-value OWS
message-body      = *OCTET / multipart-body   ; multipart-body iff Content-Type: multipart/form-data

multipart-body    = 1*( multipart-part line-ending )
multipart-part    = field-part / file-part
field-part        = "@field" SP identifier SP "=" SP ( string / interpolation )
file-part         = "@file" SP identifier SP "=" SP quoted-path [ SP "as" SP string ]

comment-line      = "#" *(VCHAR / SP / HTAB) line-ending

use-directive     = "@use" SP quoted-path [ SP "as" SP simple-identifier ] line-ending
quoted-path       = DQUOTE *(%x20-21 / %x23-7E) DQUOTE
simple-identifier = ALPHA *( ALPHA / DIGIT / "_" )

directive-line    = ( assert-directive / save-directive ) line-ending
assert-directive  = value-assert / length-assert / existence-assert / schema-assert
value-assert      = "@assert" SP assert-target SP comparator SP assert-value
length-assert     = "@assert" SP assert-target SP "length" SP comparator SP number
existence-assert  = "@assert" SP assert-target SP ( "exists" / "not" SP "exists" )
schema-assert     = "@assert" SP "body" SP "matches" SP "schema" SP quoted-path
save-directive    = "@save" SP identifier SP "=" SP assert-target
assert-target     = "status" / "duration" / header-target / body-target
header-target     = "headers." field-name
body-target       = "body" *( "." json-key / "[" 1*DIGIT "]" )
comparator        = "==" / "!=" / "<=" / ">=" / "<" / ">" / "contains" / "matches"
assert-value      = json-literal / interpolation
json-literal      = "null" / "true" / "false" / number / string
json-key          = 1*( ALPHA / DIGIT / "_" )

interpolation     = "{{" *SP interp-expr *SP "}}"
interp-expr       = identifier / function-call
identifier        = ALPHA *( ALPHA / DIGIT / "_" / "." )
function-call     = fn-name "(" [ fn-arg *( *SP "," *SP fn-arg ) ] ")"
fn-name           = "uuid" / "now" / "base64" / "hmac_sha256"
fn-arg            = identifier / string / number

line-ending       = CRLF / LF
```

*(`method`, `field-name`, `field-value`, `OWS`, `token`, `absolute-form`,
`HTTP-version` are as defined in RFC 9110; `number`, `string` follow
standard JSON literal syntax, RFC 8259 §6–7.)*

## 9. Complete Example

```
# Login and create user test

POST {{baseUrl}}/login
Content-Type: application/json

{"username": "test", "password": "pass"}

@assert status == 200
@assert body.token != null
@save token = body.token

### Create user

POST {{baseUrl}}/users
Authorization: Bearer {{token}}
Content-Type: application/json

{"name": "New User", "email": "new@example.com"}

@assert status == 201
@assert body.id != null
```

Cross-file chaining (§7.3), split across two files instead of one:

```
# login.probe
POST {{baseUrl}}/login
Content-Type: application/json

{"username": "test", "password": "pass"}

@assert status == 200
@save token = body.token
```

```
# users.probe
@use "./login.probe"

POST {{baseUrl}}/users
Authorization: Bearer {{token}}
Content-Type: application/json

{"name": "New User", "email": "new@example.com"}

@assert status == 201
@assert body.id != null
```

Running `probe run users.probe` runs `login.probe` first automatically and
makes `{{token}}` available — no separate invocation needed.

## 10. Deviations from RFC 9110

| Construct | RFC 9110 | Probe v1 |
|---|---|---|
| `HTTP-version` on request-line | Required | Optional, defaults to `HTTP/1.1` |
| Comments | Not defined | `#` lines, leading position only |
| Multiple messages per stream | Not defined (one message per connection direction) | `###`-delimited blocks per file |
| Variable interpolation | Not defined | `{{identifier}}` / `{{function()}}`, resolved before send |
| Cross-file composition | No notion of "file" at all — out of scope for the RFC | `@use "<path>"`, run before the declaring file |
| Multipart body framing | `multipart/form-data` body is opaque octets on the wire (RFC 9110 defers to RFC 7578) | `@field`/`@file` directive lines (§4.4) — the implementation assembles the actual wire encoding |
| Request-block cleanup | No notion of "cleanup" — out of scope for the RFC | `~~~` teardown block (§5.1), tied to and run after its owning request |

## 11. Non-Goals for v1 (deferred)

- **Control flow** (loops, conditionals) — PRD lists it as a Core goal but
  gives no syntax; deliberately out of scope until a v2 addendum to this
  spec, per the [roadmap](./roadmap.md#6-open-decisions-log).
- Request-to-request references by `request-name` (only forward capture via
  `@save`/`{{var}}` is specified).
- Selective (partial) variable import on `@use` — importing only *some* of
  a dependency's `@save`d names rather than all of them. `@use ... as alias`
  (§7.3) namespaces the whole set to avoid collisions; it doesn't let you
  cherry-pick a subset.
- Generic file-scoped constant declarations (i.e. a value not derived from
  a response) — use `--env` for that; `@use` only covers importing another
  file's *response-derived* variables.
- `@assert` forms beyond value/length/existence/schema (§7.1) — e.g. full
  deep-equality against a JSON fixture, array-as-set comparisons ("contains
  exactly these elements, any order").
- Chaining more than one teardown block per request-block (§5.1), and
  file-level (as opposed to request-level) teardown.
- Composable/nested function calls in interpolation (§6) — v1's
  `function-call` args are limited to `identifier`/literal, not another
  `function-call`.
- Everything in
  [roadmap §7](./roadmap.md#7-future-considerations-not-yet-scoped) that
  isn't a grammar change (data-driven runs, retries, cookie jars, import
  from other tools, etc.) — tracked there rather than here since it doesn't
  affect this document's grammar.

## 12. Versioning

This document versions independently of the PRD. Breaking grammar changes
bump the major/minor version in the title; any implementation of this
grammar (covering both the base HTTP syntax and the Probe-specific
extensions) should track which spec version it implements. `probe.toml`
(§13) is a companion file format, not part of the `.probe` grammar itself,
and versions independently — see §13.

## 13. Project Configuration File (`probe.toml`)

*Not part of the `.probe` grammar (§1–§10 above) — this section specifies a
companion TOML file that supplies defaults `.probe` files rely on but don't
declare themselves, chiefly per-environment `baseUrl` values.*

### 13.1 Discovery

Running `probe run <file>` searches for `probe.toml` starting in `<file>`'s
directory and walking up ancestor directories (the same walk-to-root pattern
`.git` discovery uses), stopping at the first one found. No `probe.toml` is
not an error — defaults are simply empty.

### 13.2 Structure

```toml
[defaults]
timeout_ms = 5000

[defaults.headers]
"User-Agent" = "probe/0.1"

[env.dev]
baseUrl = "http://localhost:3000"

[env.staging]
baseUrl = "https://staging.api.example.com"
```

- `[defaults]` applies to every run regardless of which environment is
  selected. `timeout_ms` and `[defaults.headers]` (merged into every
  request's headers, request-level headers win on conflict) are the v1
  fields; more may be added additively.
- `[env.<name>]` tables hold environment-specific values. Exactly one is
  **active** per run, selected by `--env <name>` on the CLI. Its keys
  (`baseUrl` and any other identifiers a file interpolates) become resolvable
  via `{{identifier}}` at resolution-order tier 3 (§6).
- If `--env <name>` is passed but no matching `[env.<name>]` table exists,
  that's a startup error, not a silent fall-through to defaults.

### 13.3 Relationship to `--env <file>` (PRD §7.3)

The PRD's existing `--env <file>` flag (a flat key=value dotenv-style file)
and this `probe.toml`'s `[env.<name>]` tables overlap in purpose. **Open
decision, not resolved by this spec**: whether `--env <file>` is kept as a
lower-precedence override mechanism (as reflected in §6's resolution order,
tier 2 vs. tier 3) or is subsumed entirely by `probe.toml` before M3. Tracked
in [roadmap §6](./roadmap.md#6-open-decisions-log).
