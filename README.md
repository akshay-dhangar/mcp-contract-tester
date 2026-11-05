# mcp-contract-tester

Test MCP tool contracts and failure behavior.

**Status: initial working component; full MVP is planned, not complete.**

Implemented now: Check required tool arguments; full MCP transport/schema support is planned.

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

This offline fixture envelope is our own CLI format, not an MCP wire message or JSON Schema implementation. No server is contacted. Full schema validation, MCP transport, protocol errors/timeouts and report packaging remain future milestones.

## Type keyword checker

`go run ./cmd/type-check < examples/types.json` checks a JSON value against a schema containing **only** `type`. Supports string, number, integer, boolean, null, object and array, plus unions of these names. Integers include 1.0 and 1e2; numbers are checked exactly rather than rounded through float64. Null is a value distinct from a missing value field. Exit codes are 0 match, 2 mismatch, 1 unsupported/malformed input.

Unsupported schema keywords are rejected, never silently ignored. This is not full JSON Schema validation or MCP protocol testing. It does not validate nested properties/items, required keys, bounds, references or formats. Input is limited to 1 MiB; top-level numeric tokens to 256 characters and exponent magnitude 1000. Values outside these supported limits are errors, not type mismatches.

Semantics reference: [official JSON Schema type documentation](https://json-schema.org/understanding-json-schema/reference/type). Type-only checks intentionally implement a limited subset.
