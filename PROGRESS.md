# Progress

## Required-presence milestone

Added a bounded offline JSON CLI around the existing required-key check. Deterministic positive/negative fixtures demonstrate null-as-present semantics and sorted, deduplicated missing keys. Tests cover malformed/null/wrong-type envelopes, unknown fields, multiple JSON values, oversize input and CLI exit codes 0/1/2.

No MCP protocol or JSON Schema conformance is claimed. Next: supported schema type validation, then a local transport adapter after checking official protocol documentation. Full MVP acceptance remains incomplete.

## Type-keyword milestone

Added a bounded standalone type checker for seven JSON types and unions. Exact rational numeric handling distinguishes fractional values beyond float64 precision. Unsupported schema keywords and numeric limits fail closed. Tests cover integer decimals/exponents, nullable unions, primitives, malformed schemas and unsupported constraints. Official JSON Schema type documentation consulted. Next: local MCP transport adapter; full schema and protocol conformance remain unimplemented.
