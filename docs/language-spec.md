# Probe Language Specification (v0.5, Draft)

*This is the formal reference for the `.probe` file format: both its base
HTTP syntax and its Probe-specific extensions.*

Status: **Draft, v1 scope only.** Not yet implemented. Control flow
(loops/conditionals) is explicitly deferred; see §11.

**v0.2 changes**: added built-in functions to interpolation (§6), multipart
/ file-upload bodies (§4.4), teardown blocks (§5), expanded `@assert`
comparators for length/existence/schema (§7.1), and the `probe.toml` project
configuration file (§13). These were identified as grammar-affecting gaps
that needed resolving before the Probe-extensions layer could be finalized.

**v0.3 changes**: `@use ... as <alias>` (§7.3), namespaced imports that close
a variable-shadowing hole in the flat import model v0.1–v0.2 shipped with
(two `@use`d dependencies, or a dependency and its importer, saving the same
identifier would silently overwrite one another with no error).

**v0.4 changes**: dropped the standalone `comment-line` (`#`) construct —
nothing in the PRD or roadmap ever required it, and `separator` (`###`,
§3) already carries optional free-form trailing text, so it does the
labeling/documentation job a `#` comment would have. `separator` is now
also allowed (optionally) before a file's first request block, not just
between blocks, so that first block can carry a label too. `~~~` teardown
blocks (§5) are unaffected and unchanged.

**v0.5 changes**: unified interpolation's dotted names and `@assert`'s
`body.x.y[0]` targets under one `path`/`accessor` grammar (§6), instead of
treating them as two independent dotted-path mechanisms. `{{alias.name}}`
is now just the one-`accessor` case of `path`, rather than a special
single-dot-only form; deeper chains (`{{a.b.c}}`) are syntactically valid
wherever a `path` is, with resolution failures deferred to execution time
like any other unresolved reference. `body-target` (§7.1) now reads as
`path` rooted at `body` instead of restating the same chain grammar under
`json-key`.

## 1. Notation

Grammar is given in ABNF ([RFC 5234](https://www.rfc-editor.org/rfc/rfc5234)),
the same notation [RFC 9110](https://www.rfc-editor.org/rfc/rfc9110) uses.
Core rules (`ALPHA`, `DIGIT`, `VCHAR`, `SP`, `HTAB`, `CRLF`, `OWS`) are as
defined in RFC 5234 Appendix B.1 and RFC 9110 §5.6.3. Where a production is
identical to one in RFC 9110, it's referenced by name rather than
copied, e.g. `method`, `field-name`, `field-value` from RFC 9110 §9, §5.1,
§5.5.

```abnf
line-ending = CRLF / LF   ; source files may use either; wire requests are
                          ; normalized to CRLF by the implementation
```

## 2. File structure overview

```abnf
probe-file    = *blank-line [ separator ] *use-directive request-block
                *( *blank-line separator *blank-line request-block ) *blank-line
request-block = http-request *directive-line [ teardown-block ]
blank-line    = line-ending   ; an empty line, carries no meaning here
```

A `.probe` file is one or more **request blocks**, each a plain RFC 9110
HTTP request followed by Probe directive lines, separated by an explicit
`###` delimiter (§3). There's no separate comment construct — a
`separator`'s optional trailing text doubles as that block's label, so it's
also where block-level documentation goes. Between blocks the `separator`
is required; before the file's first block it's optional, allowed purely to
label that first block (§3). A file may also declare dependencies on other
files via leading `@use` directives (§7.3) before its first request-block.
A request-block may additionally carry its own **teardown block** (§5), run
after it regardless of outcome.

## 3. Request separators

```abnf
separator    = "###" [ 1*SP request-name ] line-ending
request-name = 1*(VCHAR / SP)
```

- `###` on its own line starts a new request block. Text after `###` on the
  same line is that request's name (used in CLI/report output, as a
  reference target for future chaining features, and as a free-form label —
  there's no separate comment construct; a `###` line's trailing text is
  where block-level documentation goes).
- Between two request blocks, `###` is **required** — it's the only thing
  that marks where one block ends and the next begins.
- Before the file's **first** request block, `###` is **optional**: nothing
  precedes it there to separate, but a lone `### <name>` line is still
  allowed purely as a label for that first block. This is the one place a
  `separator` doesn't imply an actual boundary.
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

  Blank lines there are purely cosmetic (`blank-line` in §2); write them or
  not. The **one** blank line that is never optional is the one ending a
  request's header block, per RFC 9110 framing (§4.3). That's what
  separates headers from the body, not what separates requests.

## 4. Request blocks

### 4.1 Request line

```abnf
request-line   = method SP request-target [ SP HTTP-version ] line-ending
method         = token           ; RFC 9110 §9.1, any token, not just the
                                  ; standard verbs, so extension methods parse
request-target = absolute-form / interpolated-origin-form
```

- `HTTP-version` is **optional**, unlike raw RFC 9110 (which requires it on
  the wire). If omitted, the implementation assumes `HTTP/1.1`. This is the
  one deliberate deviation from RFC 9110 syntax. Probe files are a DSL over
  HTTP, not a literal wire capture.
- `request-target` is expected to be `absolute-form` (a full URL, RFC 9110
  §7.1) in the common case, since base hosts are typically supplied via
  `{{variable}}` interpolation (§6) rather than a separate `Host` line, e.g.
  `{{baseUrl}}/login`.

### 4.2 Header fields

```abnf
header-block = *( field-line line-ending )
field-line   = field-name ":" OWS field-value OWS   ; RFC 9110 §5.1, §5.5
```

Standard RFC 9110 header-field syntax, unmodified. `field-value` may contain
`{{variable}}` interpolation tokens (§6), resolved before the request is
sent. The grammar treats them as opaque text; substitution is a semantic
(execution-time) concern, not a parse-time one.

### 4.3 Message body

```abnf
http-request = request-line header-block line-ending [ message-body ]
message-body = *OCTET   ; to end of block (next separator, directive-line, or EOF)
```

Standard RFC 9110 framing: a blank line ends the header block; everything
after it, up to the next `@`-directive line, `###` separator, or EOF, is the
body. Bodies may contain `{{variable}}` interpolation tokens.

### 4.4 Multipart & file-upload bodies

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
  existing `message-body = *OCTET` framing (§4.3); `@field`/`@file` lines
  are only meaningful in a multipart body and are a syntax error elsewhere.

```
POST {{baseUrl}}/upload
Content-Type: multipart/form-data
Authorization: Bearer {{token}}

@field description = "profile picture"
@file avatar = "./fixtures/avatar.png" as "image/png"

@assert status == 201
```

## 5. Teardown blocks

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
  finishes, **regardless of whether that block's `@assert`s passed or
  failed**, as long as the owning request was actually sent (an unresolved
  `{{variable}}` that prevented sending skips the teardown too; there's
  nothing to tear down).
- A teardown block's own `@assert` failures are reported as a **warning
  attached to the owning request**, not as a separate failed test. A
  cleanup hiccup (e.g. the resource was already gone) shouldn't flip an
  otherwise-passing test to failed, but it must still be visible.
- A request-block has at most one teardown block. Chaining cleanup steps is
  out of scope for v1; see §11.

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

## 6. Variables & interpolation

```abnf
interpolation = "{{" *SP interp-expr *SP "}}"
interp-expr   = path / function-call
path          = identifier *accessor
accessor      = "." identifier / "[" 1*DIGIT "]"
identifier    = ALPHA *( ALPHA / DIGIT / "_" )
function-call = fn-name "(" [ fn-arg *( *SP "," *SP fn-arg ) ] ")"
fn-name       = "uuid" / "now" / "base64" / "hmac_sha256"
fn-arg        = path / string / number
```

`{{path}}` may appear anywhere in a request-target, header field-value, or
message body. `path` is the one accessor-chain grammar used everywhere a
value needs walking into by name or by index — it's the same `accessor`
production `body-target` (§7.1) uses for `body.x.y[0]`, just rooted at a
plain `identifier` instead of the literal `body`. There's deliberately no
second, separate dotted-path notation: `{{auth.token}}`'s `.token` and
`@assert`'s `body.token` are the same kind of step.

A bare `identifier` (no accessors) resolves directly per the lookup order
below. Each `accessor` after it — `.name` or `[index]` — walks one level
into whatever the previous step resolved to; it's a **runtime** error (not
a parse error) if that value isn't a structured value in the first place,
or the key/index isn't present, the same way an unresolved bare identifier
already is (see below). The `alias.name` form produced by an aliased `@use
... as alias` (§7.3) is just the one-accessor case of this: `alias`
resolves to that dependency's saved-value view, and `.name` looks a name up
in it.

`{{function-call()}}` computes a value at execution time instead of looking
one up; arguments may themselves be `path`s (resolved per the order below
before the function runs), not nested function calls. v1 keeps function
calls non-composable to avoid needing a general expression grammar.
Built-in v1 functions:

| Function | Args | Produces |
|---|---|---|
| `uuid()` | none | A random UUID v4 string |
| `now()` / `now(format)` | optional format `string` | Current timestamp; ISO 8601 if `format` omitted |
| `base64(value)` | `string` or `path` | Base64-encoded `value` |
| `hmac_sha256(key, value)` | two `string`s or `path`s | Hex-encoded HMAC-SHA256, for signed-request headers |

This set is deliberately small: it covers the recurring "idempotency key /
timestamp header / signed request" cases. Additional functions are additive
(new `fn-name` alternatives), not breaking changes.

**Resolution order** (first match wins):
1. Values captured by a prior `@save` in the same run (§7.2), resolved
   through the *resolving file's own `@use` view* (§7.3): its own
   request-blocks' saves, plus its dependencies' saves, either flattened
   in (unaliased `@use`) or reachable only as `alias.name` (aliased
   `@use ... as alias`). This view is assembled per file, not one
   indiscriminate global table; see §7.3 for exactly how.
2. Values from the `--env <file>` loaded at CLI startup
3. The active `[env.<name>]` table in `probe.toml` (§13), if one was
   selected and a project config file was found
4. Process environment variables

`function-call`s are evaluated fresh each time they're encountered; they
don't participate in this lookup order themselves, only their arguments do.

An unresolved `{{path}}` — its root `identifier`, or any `accessor` step
after it — or a `function-call` with an unresolvable argument, at execution
time is a runtime error (not a parse error) that fails the containing
request.

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
body-target       = "body" *accessor   ; accessor from §6's `path` grammar
comparator        = "==" / "!=" / "<=" / ">=" / "<" / ">" / "contains" / "matches"
assert-value      = json-literal / interpolation
json-literal      = "null" / "true" / "false" / number / string
```

- `body-target` is `path` (§6) rooted at the literal `body` instead of an
  `identifier` — the same accessor-chain grammar, not a second one. Dot
  notation walks object keys, bracket notation walks array indices, e.g.
  `body.user.id`, `body.items[0].name`.
- `matches` (in `value-assert`) compares against a regular expression
  `string` value.
- `length-assert` applies `comparator` to the element count of an array or
  object `assert-target`, or the character count of a string target, e.g.
  `@assert body.items length == 3`, `@assert body.items length >= 1`. It's
  a syntax error against `status`/`duration` (scalars have no length).
- `existence-assert` checks whether `assert-target` is present in the
  response at all. This is distinct from `value-assert ... != null`, which
  requires the key to be present *and* non-null. `@assert body.deletedAt not
  exists` passes if the key is absent; `@assert body.deletedAt != null`
  would error (unresolvable target) if the key is absent rather than just
  `null`.
- `schema-assert` validates the entire response body against a JSON Schema
  file at `quoted-path` (§7.3), resolved relative to the `.probe` file.
  This is the one `@assert` form that doesn't take an `assert-target`
  narrower than the whole `body`; it's a contract-level check, not a
  field-level one.
- A failed `@assert` (any of the four forms) fails the enclosing request;
  execution continues with the next request block.

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
simple-identifier  = ALPHA *( ALPHA / DIGIT / "_" )        ; no dots, see below
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

**Execution semantics** (this is a run-time concern, not just syntax; it
changes what running a file means):

- Before a file's own request-blocks execute, each of its `@use`d files
  runs first, in declaration order. `@use` is transitive: if a dependency
  itself has `@use` lines, those resolve depth-first before it runs.
- A given file, resolved to an absolute path, executes **at most once per
  `probe run`**, no matter how many other files `@use` it. If both
  `users.probe` and `orders.probe` `@use "./login.probe"` in the same
  invocation, `login.probe` runs once; both reuse its captured `@save`
  results (each through its own view, per below; aliasing them
  differently, or not at all, doesn't force a second run).
- A dependency cycle (`A` uses `B` uses `A`) is a hard error at resolution
  time, reported with the offending cycle path. This formalizes the
  "circular dependencies" edge case.
- If a dependency's requests or assertions fail, every file that
  transitively depends on it fails too, attributed to the dependency;
  dependents are not run against a broken fixture. How this is *reported*
  in CLI output is a test-runner concern, not specified here.

**Namespacing (`as alias`) and variable visibility:**

Each file builds its own **view** of `{{identifier}}`-resolvable names,
resolution-order tier 1 (§6), out of two sources: its own request-blocks'
`@save`s, and its `@use`d dependencies' `@save`s, combined per directive:

- **`@use "path" as alias`** (namespaced): the dependency's saved values are
  reachable *only* as `{{alias.name}}`, never as bare `{{name}}`. This is
  the one-`accessor` case of the general `path` grammar (§6) — `alias`
  resolves to the dependency's view, `.name` looks a name up in it, same
  as any other accessor step. Namespacing is local to the importing file;
  it doesn't rename anything inside the dependency itself, and it doesn't
  propagate: if that dependency
  has its own `@use`s, *their* names aren't re-exposed through `alias.*`
  unless the dependency chooses to re-export them (not a v1 concept; see
  §11).
- **`@use "path"`** (unaliased, the v1-original form): the dependency's
  saved values flatten directly into the importing file's own flat
  namespace, alongside its own `@save`s. This is unchanged from pre-v0.3
  behavior.
- **Collision rule for flattened (unaliased) names**: if two different
  `@use`d files, neither one a dependency of the other, would both
  contribute the same flat name to the same importing file's view, that's a
  **hard error at resolution time**, naming both source files and the
  conflicting identifier. Fix it by aliasing at least one of them.
  Rebinding is *not* an error when it happens within one dependency chain.
  A file's own `@save` may always override a flat name it imported (the
  importer's own value wins), and a later request in the same file may
  freely re-`@save` a name it (or its flat imports) already bound, e.g. a
  token-refresh request re-saving `token` after login is normal and allowed.
- **Alias collisions**: two `@use` lines in the same file may not declare
  the same `alias`. That's a syntax-adjacent error caught at resolution
  time (same phase as the flat-name collision above).

```
### orders.probe: same dependency imported twice, safely, via aliasing
@use "./login-as-customer.probe" as customer
@use "./login-as-admin.probe" as admin

DELETE {{baseUrl}}/orders/{{orderId}}
Authorization: Bearer {{admin.token}}

@assert status == 204
```

Without the `as` clauses here, both dependencies would `@save token`, and
`{{token}}` downstream would silently resolve to whichever one ran last.
That's exactly the shadowing hazard aliasing exists to close.

- Selectively importing only *some* of a dependency's variables (as opposed
  to all of them, whether flattened or under one alias) is still deferred;
  see §11.

## 8. Consolidated grammar (appendix)

```abnf
probe-file        = *blank-line [ separator ] *use-directive request-block
                    *( *blank-line separator *blank-line request-block ) *blank-line
request-block     = http-request *directive-line [ teardown-block ]
blank-line        = line-ending
separator         = "###" [ 1*SP request-name ] line-ending
teardown-block    = teardown-sep *blank-line http-request *directive-line
teardown-sep      = "~~~" [ 1*SP request-name ] line-ending
request-name      = 1*(VCHAR / SP)

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
body-target       = "body" *accessor
comparator        = "==" / "!=" / "<=" / ">=" / "<" / ">" / "contains" / "matches"
assert-value      = json-literal / interpolation
json-literal      = "null" / "true" / "false" / number / string

interpolation     = "{{" *SP interp-expr *SP "}}"
interp-expr       = path / function-call
path              = identifier *accessor
accessor          = "." identifier / "[" 1*DIGIT "]"
identifier        = ALPHA *( ALPHA / DIGIT / "_" )
function-call     = fn-name "(" [ fn-arg *( *SP "," *SP fn-arg ) ] ")"
fn-name           = "uuid" / "now" / "base64" / "hmac_sha256"
fn-arg            = path / string / number

line-ending       = CRLF / LF
```

*(`method`, `field-name`, `field-value`, `OWS`, `token`, `absolute-form`,
`HTTP-version` are as defined in RFC 9110; `number`, `string` follow
standard JSON literal syntax, RFC 8259 §6–7.)*

## 9. Complete example

```
### Login and create user test

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
### login.probe
POST {{baseUrl}}/login
Content-Type: application/json

{"username": "test", "password": "pass"}

@assert status == 200
@save token = body.token
```

```
### users.probe
@use "./login.probe"

POST {{baseUrl}}/users
Authorization: Bearer {{token}}
Content-Type: application/json

{"name": "New User", "email": "new@example.com"}

@assert status == 201
@assert body.id != null
```

Running `probe run users.probe` runs `login.probe` first automatically and
makes `{{token}}` available; no separate invocation needed.

## 10. Deviations from RFC 9110

| Construct | RFC 9110 | Probe v1 |
|---|---|---|
| `HTTP-version` on request-line | Required | Optional, defaults to `HTTP/1.1` |
| Multiple messages per stream | Not defined (one message per connection direction) | `###`-delimited blocks per file |
| Variable interpolation | Not defined | `{{identifier}}` / `{{function()}}`, resolved before send |
| Cross-file composition | No notion of "file" at all; out of scope for the RFC | `@use "<path>"`, run before the declaring file |
| Multipart body framing | `multipart/form-data` body is opaque octets on the wire (RFC 9110 defers to RFC 7578) | `@field`/`@file` directive lines (§4.4); the implementation assembles the actual wire encoding |
| Request-block cleanup | No notion of "cleanup"; out of scope for the RFC | `~~~` teardown block (§5), tied to and run after its owning request |

## 11. Non-goals for v1 (deferred)

- **Control flow** (loops, conditionals): deliberately out of scope until a
  v2 addendum to this spec.
- Request-to-request references by `request-name` (only forward capture via
  `@save`/`{{var}}` is specified).
- Selective (partial) variable import on `@use`: importing only *some* of
  a dependency's `@save`d names rather than all of them. `@use ... as alias`
  (§7.3) namespaces the whole set to avoid collisions; it doesn't let you
  cherry-pick a subset.
- Generic file-scoped constant declarations (i.e. a value not derived from
  a response). Use `--env` for that; `@use` only covers importing another
  file's *response-derived* variables.
- `@assert` forms beyond value/length/existence/schema (§7.1), e.g. full
  deep-equality against a JSON fixture, array-as-set comparisons ("contains
  exactly these elements, any order").
- Chaining more than one teardown block per request-block (§5), and
  file-level (as opposed to request-level) teardown.
- Composable/nested function calls in interpolation (§6): v1's
  `function-call` args are limited to `path`/literal, not another
  `function-call`.
- Operational features that don't affect this document's grammar:
  data-driven runs, retries, cookie jars, importing from other tools, and
  similar.

## 12. Versioning

This document has its own version, tracked separately. Breaking grammar
changes bump the major/minor version in the title; any implementation of
this grammar (covering both the base HTTP syntax and the Probe-specific
extensions) should track which spec version it implements. `probe.toml`
(§13) is a companion file format, not part of the `.probe` grammar itself,
and versions independently; see §13.

## 13. Project configuration file (`probe.toml`)

*Not part of the `.probe` grammar (§1–§10 above). This section specifies a
companion TOML file that supplies defaults `.probe` files rely on but don't
declare themselves, chiefly per-environment `baseUrl` values.*

### 13.1 Discovery

Running `probe run <file>` searches for `probe.toml` starting in `<file>`'s
directory and walking up ancestor directories (the same walk-to-root pattern
`.git` discovery uses), stopping at the first one found. No `probe.toml` is
not an error; defaults are simply empty.

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

### 13.3 Relationship to `--env <file>`

The existing `--env <file>` flag (a flat key=value dotenv-style file) and
this `probe.toml`'s `[env.<name>]` tables overlap in purpose. **Open
decision, not resolved by this spec**: whether `--env <file>` is kept as a
lower-precedence override mechanism (as reflected in §6's resolution order,
tier 2 vs. tier 3) or is subsumed entirely by `probe.toml`.

## 14. Multi-protocol method classification (proposed)

*Status: proposed, from a 2026-08-23 design conversation. Not decided, not
implemented, and out of the "v1 scope only" line at the top of this
document; v1 is HTTP-only. Accepting this would extend the grammar past
what the PRD's non-goals currently allow (PRD §1.5, §4.3 list WebSocket and
gRPC support as explicitly out of scope), and it would earn a real version
bump under §12's rule once it lands, not just this section. Recorded here
so the design isn't lost, not as a commitment to build it.*

The core idea: `method` is already `token` (RFC 9110 §9.1), unrestricted
beyond the standard verbs (§4.1). Four method values are reserved to select
a protocol handler instead of an HTTP verb: `GRAPHQL`, `GRPC`, `WS`, `WSS`.
`request-line` itself doesn't change. No new grammar there, just four
tokens an implementation treats specially. A reserved method is an exact
uppercase match; `graphql`, `Graphql`, etc. are ordinary (if unusual) HTTP
extension methods, not the reserved form.

Nothing about `request-target` changes either. `absolute-form` already
covers a GraphQL endpoint URL, a gRPC method path
(`scheme://host:port/package.Service/Method`, gRPC's actual addressing
scheme over HTTP/2), and a `ws://`/`wss://` socket URL. If the URL's scheme
and the method disagree (e.g. `WSS` against an `http://` target), the
method wins.

Each reserved method synthesizes the transport-level headers its protocol
needs (`Content-Type`, `Upgrade`/`Sec-WebSocket-*`, and so on) instead of
requiring them written out by hand. An explicit `field-line` for the same
header in the request block overrides the synthesized value, the same
precedence multipart bodies already give an explicit `boundary` over the
generated one (§4.4).

### 14.1 `GRAPHQL`

Synthesizes `POST` and `Content-Type: application/json`. Adds a fourth
`message_body` alternative alongside `octet_body` and `multipart_body`
(§4.3, §4.4):

```abnf
graphql-body         = graphql-query-text [ *blank-line variables-directive ]
graphql-query-text   = *OCTET   ; to variables-directive, separator,
                                ; directive-line, or EOF, same framing
                                ; octet-body already uses
variables-directive  = "@variables" SP json-value line-ending
```

`json-value` is a JSON value per RFC 8259 §3 (so, unlike `assert-value`'s
`json-literal`, an object or array is allowed; `@variables` is normally an
object). Strings inside it may carry `{{variable}}` interpolation tokens,
resolved the same as anywhere else in a body (§4.3, §6).

```
GRAPHQL {{baseUrl}}/graphql

query GetUser($id: ID!) {
  user(id: $id) { id name }
}

@variables {"id": "{{userId}}"}

@assert status == 200
@assert body.errors not exists
@save userName = body.data.user.name
```

`@assert`/`@save` need no new grammar: `body.data.*` and `body.errors` are
ordinary `body-target` dot-paths (§7.1). Because GraphQL returns `200` on
most errors, `@assert body.errors not exists` (an `existence-assert`, §7.1)
is the actual pass/fail check, not `status`.

Whether `@variables` under a non-`GRAPHQL` method is a parse error or a
semantic one follows the existing pattern: `graphql-body` parses
structurally under any method, same as `multipart_body` does today (§4.4
already allows `@field`/`@file` to parse and then rejects them
semantically outside a multipart body); a mismatched method is caught at
validation, not at parse time.

**Implementation note, not spec.** The external scanner
(`tree-sitter-probe/src/scanner.c`) currently reserves `###`, `~~~`,
`@assert`, `@save`, `@field`, `@file` as boundary markers `OCTET_BODY`
stops before. `@variables` would need to join that set before
`graphql-query-text` could coexist with the existing body productions
without ambiguity.

### 14.2 `GRPC`

Synthesizes an HTTP/2 request with `content-type: application/grpc+json`
(assuming JSON transcoding of the Protobuf message, grpc-gateway/grpcurl
style, rather than requiring the raw binary encoding to be hand-written).
The body stays a plain JSON `octet_body`; no new body grammar.

gRPC's real result code is a trailer, independent of the HTTP/2 `:status`
pseudo-header (which is `200` on most calls, failed or not). Reusing
`status` for both would mean the same token silently means two different
things depending on which method preceded it, so this adds a distinct
target instead of overloading `status`:

```abnf
assert-target = "status" / "duration" / "grpc-status" / header-target / body-target
```

`grpc-status` is a scalar like `status`/`duration`: a `length-assert`
against it is a syntax error for the same reason (§7.1).

```
GRPC {{baseUrl}}/user.UserService/GetUser
Authorization: Bearer {{token}}

{"id": "{{userId}}"}

@assert grpc-status == 0
@save userName = body.name
```

Not addressed here, and not solved by method classification alone:

- **Message shape.** JSON transcoding needs a Protobuf descriptor source
  (server reflection, or a compiled `.proto` file) to know what a method's
  request/response even look like. How a `.probe` file would declare that
  source (a `@proto "<path>"` directive, mirroring `@use`'s `quoted-path`,
  is the obvious shape) isn't specified.
- **Streaming RPCs.** Client-streaming, server-streaming, and bidi calls
  don't fit one request block producing one response. See §14.4.

### 14.3 `WS` / `WSS`

Synthesizes `GET` plus `Upgrade: websocket`, `Connection: Upgrade`, and
`Sec-WebSocket-Version: 13`. The handshake itself needs no new grammar.
It's an ordinary request block, and a successful upgrade is an ordinary
`@assert status == 101`:

```
WSS {{baseUrl}}/socket
Authorization: Bearer {{token}}

@assert status == 101
```

What happens on the connection after the handshake is out of scope for
this section. See §14.4.

### 14.4 Deferred: streaming sessions

A WebSocket connection past the handshake, and a streaming gRPC call, are
both long-lived and bidirectional. Neither fits the "one request block,
one response" shape `@assert`/`@save` are built around. Both need the same
kind of thing: a directive family for sending and expecting further
messages against an already-open connection, e.g. (illustrative, not
specified):

```
@send {"type": "subscribe", "channel": "orders"}
@expect body.type == "ack" timeout 2000
```

This is the largest unresolved piece of the whole proposal and isn't
designed past this paragraph. It would need its own grammar section (send/
expect directive syntax, ordering and timeout semantics) and its own
execution model (a request block that, once opened, becomes a session
other directive lines act within). Until it exists, `WS`/`WSS` covers the
handshake only, and `GRPC` covers unary calls only.
