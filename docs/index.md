# 🧭 Product Requirements Document (PRD)

## 1. Overview

### 1.1 Product Name

*What is this product/tool called?*
Probe

### 1.2 Summary

*One paragraph describing what the product does, who it’s for, and why it exists.*
Probe is a API testing tool that is fully text based, leveraging the HTTP
syntax from [RFC 9110](https://datatracker.ietf.org/doc/html/rfc9110) and added
in some more syntax make it function as a Language for writing API tests

Probe is primarily targetting backend developers and QAs

This is an low effort way of writing Network tests, scriptable, easier to share and can easily able to write e2e api 
tests that can run in the pipeline

### 1.3 Background / Context

*Why are you building this now? What problem or opportunity triggered it?*
I have tried similar tools but there is too much friction for it to be
functional and effortless, trying to bring ideas and improve and build upon
them

### 1.4 Goals

*List 3–5 concrete goals this release must achieve. Example:*
- Core
    - http syntax parser
    - language parser probe injecting http syntax parser in it
    - test runner as a CLI tool

- Utilities leveraging Core utils to give a better experience
    - LSP for probe language
    - VS Code and NeoVim Integration through plugin connecting to the probe lang

* Reduce manual testing
- Easily create test suites

### 1.5 Non-Goals

*Clearly state what’s **not** in scope, so you don’t waste time building unnecessary features.*

- Advanced reporting dashboards (e.g., visual graphs, charts)
- Team collaboration features (e.g., shared test suites)
- CI/CD pipeline integration
- Support for non-HTTP protocols (e.g., WebSocket, gRPC)
- Mobile app or dedicated GUI

---

## 2. Problem Definition

### 2.1 User Pain Points

*Describe the problem in the user’s terms — what’s broken or inefficient about their current workflow?*

- Writing API tests often requires complex setup and scripting knowledge.
- Existing tools are either too rigid or too complex for quick test creation.
- Lack of a standardized, shareable format for API tests.
- Manual testing is time-consuming and error-prone.
- Limited integration with developer workflows (e.g., CLI, editors).

### 2.2 Why Now

*What makes this problem urgent or valuable to solve right now?*

- Similar tools exist but none have found the right balance of simplicity and power
- HTTP syntax from RFC 9110 provides a solid, standardized foundation to build upon
- Modern LSP and editor plugin ecosystems make it easier to provide great developer experience
- Backend developers and QAs need scriptable, text-based tools that fit into their existing workflows

### 2.3 Success Metrics

*Define measurable outcomes (quantitative or qualitative):*

- Developers can write their first API test in under 5 minutes
- Test execution is faster than manual testing for 3+ endpoint workflows
- 80% of beta users prefer Probe over their current tool for simple API testing
- Users successfully integrate Probe into CI/CD pipelines
- LSP provides helpful autocomplete and validation in supported editors

---

## 3. Target Users

### 3.1 Primary User Personas

*Describe key user groups (include short bios if useful).*

* **Backend Developer** – Works in terminal/editor, builds REST APIs, wants quick way to test endpoints during development without leaving their workflow
* **QA Engineer** – Writes automated test suites, needs readable and maintainable test files that can be version controlled and reviewed like code

### 3.2 User Scenarios / Use Cases

*List the key workflows you must support.*

- Developer writes HTTP request in .probe file using familiar HTTP syntax
- Run single test file from CLI: `probe run api.probe`
- Run test suite with multiple files
- Write assertions to validate responses
- Chain requests together (use response from one request in another)
- Get editor support (syntax highlighting, autocomplete, validation) through LSP

---

## 4. Product Scope

### 4.1 Core Features (MVP)

*List and describe each feature that must be built to call this v1 done.*

* **HTTP Syntax Parser** – Parse standard HTTP request/response syntax from RFC 9110
* **Probe Language Parser** – Extended syntax for variables, assertions, test blocks, and control flow
* **CLI Test Runner** – Execute .probe files and report results
* **LSP Server** – Language server providing editor integration features
* **Editor Plugins** – VS Code and NeoVim extensions connecting to LSP

### 4.2 Nice-to-Have Features (Post-MVP)

*These are important but not blockers for initial release.*

- HTML/JSON test report generation
- Test coverage reporting
- Watch mode for continuous test running during development
- Plugins/extensions system for custom assertions or transformers
- Integration with popular test frameworks (Jest, pytest patterns)
- Parallel test execution

### 4.3 Out of Scope (for now)

*Explicitly list features you're deferring or avoiding.*

- Graphical user interface or desktop application
- Performance/load testing features
- Mock server or stubbing capabilities
- WebSocket, gRPC, or other non-HTTP protocol support
- Cloud-based test execution service

---

## 5. Functional Requirements

### 5.1 User Flows

*Describe or sketch the flow of key tasks (CLI → API → UI).*

1. Developer writes .probe file with HTTP requests and test assertions
2. Run `probe run <file>` from CLI
3. Parser reads file, builds AST
4. Test runner executes requests sequentially or in defined order
5. Assertions are evaluated against responses
6. Results printed to terminal with pass/fail status
7. Exit code indicates overall success/failure for CI integration

### 5.2 System Behavior

*What should happen on each user action? (e.g., running a test triggers X, logs Y, stores Z.)*

- **Running tests**: Parse file → Execute HTTP requests → Evaluate assertions → Display results
- **Syntax errors**: Display error with line number, stop execution
- **Network failures**: Display error with context, mark test as failed
- **Variable extraction**: Store values from responses for use in subsequent requests
- **Editor actions**: LSP provides diagnostics, autocomplete, hover information in real-time

### 5.3 Edge Cases / Constraints

*What happens if the network fails? How are partial results handled?*

- **Network failures**: Display clear error message, mark test as failed, continue with remaining tests
- **Timeouts**: Configurable timeout per request, fail test if exceeded
- **Invalid responses**: Handle non-JSON/malformed responses gracefully
- **Circular dependencies**: Detect and prevent infinite loops in variable references
- **Large responses**: Stream or truncate very large response bodies in output
- **File not found**: Clear error if .probe file doesn't exist

---

## 6. Technical Design

### 6.1 Architecture Overview

*Describe your system at a high level — use diagrams if possible (API, DB, CLI, UI relationships).*

```
┌─────────────┐
│ .probe file │
└──────┬──────┘
       │
       v
┌─────────────────┐
│  HTTP Parser    │ ◄── RFC 9110 compliant
└────────┬────────┘
         │
         v
┌─────────────────┐
│ Probe Parser    │ ◄── Extensions (variables, assertions)
└────────┬────────┘
         │
         v
┌─────────────────┐
│      AST        │
└────┬────────┬───┘
     │        │
     v        v
┌─────────┐  ┌──────────┐
│CLI Tool │  │LSP Server│ ◄── Editor integration
└─────────┘  └────┬─────┘
                  │
                  v
            ┌──────────────┐
            │Editor Plugins│
            └──────────────┘
```

### 6.2 Key Components

* **HTTP Parser**: Tokenizes and parses HTTP message syntax
* **Probe Parser**: Handles Probe-specific directives (@assert, @save, etc.)
* **Test Executor**: Runs HTTP requests, evaluates assertions
* **Variable Store**: Manages variable scope and substitution
* **LSP Server**: Implements Language Server Protocol for editor features
* **CLI**: User-facing command-line interface

### 6.3 Dependencies

*List frameworks, libraries, APIs, or 3rd-party services.*

- **Parser generator** (pest, nom, tree-sitter, or hand-written)
- **HTTP client library** (for making requests)
- **LSP library** (tower-lsp, vscode-languageserver, etc.)
- **CLI framework** (clap, commander, etc.)
- Standard library crypto for any hashing needs
- No external services - fully local execution

### 6.4 Scalability & Reliability

*How will this system scale and remain stable as users grow?*

- **Local execution**: Each user runs independently, no central service to scale
- **Efficient parsing**: Fast parser to handle large test files
- **Streaming**: Stream large responses to avoid memory issues
- **Parallel execution**: Optional parallel test execution for performance
- **Incremental parsing** in LSP for responsive editor experience

### 6.5 Security & Compliance

*Authentication, data encryption, access control, logging, etc.*

- No central authentication - tool runs locally with user's permissions
- Secrets management: Support environment variables, warn against hardcoding
- HTTPS/TLS for actual HTTP requests from tests
- No telemetry or data collection without explicit opt-in
- Open source for transparency and security auditing

---

## 7. User Experience (UX)

### 7.1 Design Principles

*What kind of experience are you going for? (fast, minimal, dev-friendly, etc.)*

- **Low friction**: Minimal setup, start writing tests immediately
- **Familiar syntax**: Use standard HTTP syntax as foundation
- **Text-based**: Everything is plain text, easily version controlled and diffable
- **Editor-native**: First-class support in developer's existing editor
- **Scriptable**: Easy to integrate into any workflow or pipeline

### 7.2 Wireframes / Mockups

*Include low- or high-fidelity sketches of the UI or CLI output.*

- **Example .probe file**:
  ```
  # Login and create user test
  POST https://api.example.com/login
  Content-Type: application/json
  
  {"username": "test", "password": "pass"}
  
  @assert status == 200
  @save token = body.token
  
  POST https://api.example.com/users
  Authorization: Bearer {{token}}
  Content-Type: application/json
  
  {"name": "New User", "email": "new@example.com"}
  
  @assert status == 201
  @assert body.id != null
  ```

- **CLI Output Example**:
  ```
  $ probe run api.probe
  ✓ POST /login (243ms)
  ✓ POST /users (189ms)
  
  2 passed, 0 failed
  ```

### 7.3 Interaction Notes

*Describe how users interact with each component, especially CLI prompts, flags, or web UI behavior.*

- **CLI**:
  - `probe run <file>` - Execute single test file
  - `probe run <directory>` - Execute all .probe files in directory
  - `--verbose` - Show full request/response details
  - `--env <file>` - Load environment variables
  - Exit code 0 for pass, non-zero for failures (CI-friendly)

- **Editor**:
  - Syntax highlighting for .probe files
  - Autocomplete for HTTP methods, headers, Probe directives
  - Inline error highlighting for syntax errors
  - Hover documentation for Probe features

---

## 8. Project Plan

### 8.1 Milestones

| Milestone | Deliverable                       | Target Date |
| --------- | --------------------------------- | ----------- |
| M1        | HTTP syntax parser (RFC 9110)     | TBD         |
| M2        | Probe language extensions parser  | TBD         |
| M3        | CLI test runner MVP               | TBD         |
| M4        | LSP server implementation         | TBD         |
| M5        | VS Code & NeoVim plugins          | TBD         |

### 8.2 Risks & Mitigations

| Risk                         | Impact | Mitigation                                    |
| ---------------------------- | ------ | --------------------------------------------- |
| Parser complexity            | High   | Start with subset of HTTP, expand gradually   |
| Language design decisions    | High   | Early feedback from target users, iterations  |
| LSP implementation effort    | Medium | Focus on core features first (syntax, errors) |
| Editor plugin maintenance    | Medium | Keep plugins thin, logic in LSP server        |
| Adoption vs existing tools   | Medium | Clear docs, examples, better DX               |

### 8.3 Team & Ownership

*Who is responsible for each part (you can still list yourself for all if solo).*

- **Core Parser Development**: [Your Name]
- **CLI Tool**: [Your Name]
- **LSP Server**: [Your Name]
- **Editor Plugins**: [Your Name]
- **Documentation**: [Your Name]

---

## 9. Launch Plan

### 9.1 Beta Strategy

*Who will test early builds? How will feedback be gathered?*

- **Beta Testers**: Backend developers and QA engineers who currently use tools like Postman, curl scripts, or similar
- **Feedback Channels**: GitHub issues, direct conversations, usage analytics (if users opt-in)
- **Testing Goals**: Validate that syntax feels natural, tool fits into workflows, identify missing features
- **Dogfooding**: Use Probe to test Probe itself

### 9.2 Release Plan

*How will you release — public beta, staged rollout, etc.*

- **Alpha Release**: CLI tool with basic parsing, gather initial feedback
- **Beta Release**: Add LSP and editor plugins, expand to more testers
- **v1.0**: Stable release with core features complete and documented
- **Versioning**: Semantic versioning
- **Distribution**: GitHub releases, package managers (homebrew, npm, cargo, etc.)

### 9.3 Marketing / Adoption Plan

*How will people discover and start using it? (Docs, tutorials, CLI guide, etc.)*

- **Documentation**: Clear getting-started guide, language reference, examples
- **Example Tests**: Repository of example .probe files for common scenarios
- **Blog Posts**: Write about design decisions, comparisons with other tools
- **Developer Communities**: Share on Reddit (r/programming, r/webdev), Hacker News, dev.to
- **GitHub**: Open source, good README, examples, clear contribution guidelines

---

## 10. Appendix

### 10.1 Glossary

- **RFC 9110**: HTTP Semantics specification defining HTTP syntax
- **LSP**: Language Server Protocol - standard for editor integration
- **.probe file**: Text file containing HTTP requests and Probe test directives
- **AST**: Abstract Syntax Tree - parsed representation of code

### 10.2 Reference Links

- [RFC 9110 - HTTP Semantics](https://datatracker.ietf.org/doc/html/rfc9110)
- Language Server Protocol: https://microsoft.github.io/language-server-protocol/
- Competitor analysis: Postman, Insomnia, HTTPie, curl, REST Client (VS Code)

### 10.3 Similar Tools Comparison

- **Postman**: GUI-heavy, proprietary format, great for manual testing
- **Insomnia**: Similar to Postman, JSON-based collections
- **HTTPie**: CLI, great syntax but not test-oriented
- **REST Client (VS Code)**: Close to Probe's vision but limited to VS Code, no formal language
- **curl scripts**: Flexible but verbose, no assertions built-in

