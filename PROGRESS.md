# Progress

## Required-presence milestone

Added a bounded offline JSON CLI around the existing required-key check. Deterministic positive/negative fixtures demonstrate null-as-present semantics and sorted, deduplicated missing keys. Tests cover malformed/null/wrong-type envelopes, unknown fields, multiple JSON values, oversize input and CLI exit codes 0/1/2.

No MCP protocol or JSON Schema conformance is claimed. Next: supported schema type validation, then a local transport adapter after checking official protocol documentation. Full MVP acceptance remains incomplete.

## Type-keyword milestone

Added a bounded standalone type checker for seven JSON types and unions. Exact rational numeric handling distinguishes fractional values beyond float64 precision. Unsupported schema keywords and numeric limits fail closed. Tests cover integer decimals/exponents, nullable unions, primitives, malformed schemas and unsupported constraints. Official JSON Schema type documentation consulted. Next: local MCP transport adapter; full schema and protocol conformance remain unimplemented.

## 5 November — stdio framing foundation

Implemented bounded newline-delimited UTF-8 JSON-RPC reading/writing after consulting the versioned MCP transport specification. Tests cover request/notification/result/error round trips, consecutive frames, EOF, invalid envelopes, oversized input, truncated input and short writes. This completes only the framing portion of Day 3. Subprocess lifecycle, initialization, response matching and deadlines remain next; Day 3 stays unchecked.

## 6 November — local stdio adapter

The mcp-probe CLI starts a trusted local executable and verifies a version-pinned initialization handshake. Subprocess tests exercise notification delivery and stdin EOF, response ID mismatch and a silent server hitting its deadline. Day 3 is complete for the one-shot probe. Next, expand protocol failure cases and report diagnostics for Day 4.
