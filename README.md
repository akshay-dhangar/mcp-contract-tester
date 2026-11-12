# mcp-contract-tester

Test MCP tool contracts and failure behavior.

**Status: scoped local MVP implemented. This is not a full MCP conformance suite.**

Implemented now: offline required-argument and type checks, bounded stdio framing, and a one-shot MCP initialization probe with lifecycle diagnostics.

Backend: Go. This seed has no live model calls, credentials, external writes, or deployment.

## Verify

```sh
go vet ./...
go test -race ./...
```

## One-week MVP

See [ROADMAP.md](ROADMAP.md) for seven bounded development days and acceptance criteria. Work is scheduled sequentially across the portfolio, not on every repository every day. The week starts when this project becomes active.

## Evidence

The current automated checks validate the small implemented component only. They do not establish end-to-end AI quality, scalability, or production readiness. Later demos and LinkedIn write-ups should cite measured behavior and disclosed limitations.

## Required-presence CLI

```sh
go build -o /tmp/contract-check ./cmd/contract-check
/tmp/contract-check < examples/present.json
/tmp/contract-check < examples/missing.json
```

The second example emits sorted missing names and exits 2. Exit 0 means all named keys are present; exit 1 means malformed/oversized input or an output failure. Input is capped at 1 MiB and must contain exactly one object with `required` (string array) and `arguments` (object). Unknown envelope fields are rejected. Duplicate required names are deduplicated. Empty property names are allowed. Null, false and zero values count as present: this is not type validation.

This offline fixture envelope is our own CLI format, not an MCP wire message or JSON Schema implementation. No server is contacted. Full schema validation, tool-call testing and report packaging remain future milestones.

## Type keyword checker

`go run ./cmd/type-check < examples/types.json` checks a JSON value against a schema containing **only** `type`. Supports string, number, integer, boolean, null, object and array, plus unions of these names. Integers include 1.0 and 1e2; numbers are checked exactly rather than rounded through float64. Null is a value distinct from a missing value field. Exit codes are 0 match, 2 mismatch, 1 unsupported/malformed input.

Unsupported schema keywords are rejected, never silently ignored. This is not full JSON Schema validation or MCP protocol testing. It does not validate nested properties/items, required keys, bounds, references or formats. Input is limited to 1 MiB; top-level numeric tokens to 256 characters and exponent magnitude 1000. Values outside these supported limits are errors, not type mismatches.

Semantics reference: [official JSON Schema type documentation](https://json-schema.org/understanding-json-schema/reference/type). Type-only checks intentionally implement a limited subset.

## Stdio framing foundation

The Go `FrameReader`/`WriteFrame` API reads and writes newline-delimited UTF-8 JSON-RPC envelopes with a 1 MiB message limit. It rejects malformed envelopes, embedded literal newlines, invalid IDs, conflicting result/error fields and truncated frames. Discard the reader after an error. Writers detect short writes; readers propagate EOF and I/O errors.

This is the framing layer used by the local transport adapter. It does not launch subprocesses, enforce I/O deadlines, initialize MCP sessions, match response IDs, or validate method-specific payloads. It is not a complete MCP client. The caller must supply trusted streams and manage their lifecycle. Null response IDs are deliberately unsupported in this MCP-focused subset.

Reference: [MCP 2025-11-25 stdio transport](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports). This milestone targets that documented version rather than claiming support for every protocol version.

## Probe a local MCP server

Run `go run ./cmd/mcp-probe -timeout 5s -- /absolute/path/to/trusted-server [args...]`. The probe launches that executable directly with the current user's environment, checks the response ID, protocol version, capabilities object and server identity, sends the initialized notification, and closes stdin. A successful exit prints the server name and version as JSON. The deadline covers the whole exchange including server exit. The probe can terminate its direct child on failure or timeout.

This discovery probe targets version 2025-11-25 and expects the initialize reply as the first incoming frame. Tool calls, interleaved messages and process-tree isolation are outside its scope. See the [MCP lifecycle specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/lifecycle).

### Diagnose probe failures

Errors identify the failing lifecycle stage: `configuration` for invalid deadlines, `startup` for launch errors, `initialize` for handshake failures, and `shutdown` for failure to exit cleanly after initialization. For example, `initialize: unsupported protocol version; expected 2025-11-25` means the server must support the pinned version; `shutdown: context deadline exceeded` means initialization succeeded but the server did not exit within the total timeout. Check that it exits when stdin closes.

Go callers can inspect `*ProbeError` using `errors.As` and detect deadlines/cancellation with `errors.Is`. Server RPC errors include their numeric code but omit their message/data. This is not a general redaction guarantee for all OS errors. Only the direct child is terminated; run trusted executables. Automated failure tests include malformed replies, unsupported versions, cancellation, early EOF, nonzero exits and shutdown hangs. Structured JSON reports are available with `-report`.

### Save a probe report

```sh
go build -o /tmp/mcp-probe ./cmd/mcp-probe
/tmp/mcp-probe -report -timeout 5s -- /absolute/path/to/trusted-server > report.json
```

`-report` emits one JSON object on stdout on probe success or failure: `report_version: 1`, `check: "stdio_initialization"`, and `passed`. Success includes `server`; failure includes `stage`, `code`, `message` and a stage-specific `hint`. Codes are `deadline_exceeded`, `canceled`, or `probe_failed`; stage and code are intended for automation, while message wording may change. Exit 0 means success; exit 1 means probe, usage or output failure. Invalid CLI flags and missing executables in the command line (no positional argument) produce usage diagnostics on stderr, not a report. An output-write failure also goes to stderr.

Without `-report`, the existing server-identity JSON and stderr diagnostics are preserved. Reports omit executable arguments and environment values, but OS errors can include local paths and successful reports include server-supplied identity. Review artifacts before sharing them. This report covers only initialization and shutdown; it does not establish tool-call or full schema conformance.

## Reproduce the scoped MVP

Requires Go 1.25, Python 3 and a POSIX shell for the existing subprocess tests. From the repository root, run `python3 scripts/demo.py`. It builds four CLIs into a temporary directory, checks seven deterministic scenarios and removes the binaries afterward. The same demo runs in CI. See [docs/DEMO.md](docs/DEMO.md) for expected results and limits.

The `example-server` executable is a purpose-built handshake fixture with `-mode success`, `wrong-id`, `wrong-version` or `rpc-error`. It expects this probe's ID 1 and exits after stdin closes. It does not implement tools or serve as an independent conformance oracle.
