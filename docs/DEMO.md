# Local MVP demo

Run from the repository root:

```sh
go vet ./...
go test -race ./...
python3 scripts/demo.py
```

The demo builds actual executables, launches the fixture through the probe, asserts JSON results and exit statuses, and prints each report. No API keys, network services or persistent output files are needed after Go is installed.

| Scenario | Expected exit | Evidence |
|---|---:|---|
| Successful initialization | 0 | Server identity is deterministic-fixture |
| Wrong response ID | 1 | initialize stage, ID mismatch |
| Unsupported version | 1 | initialize stage, unsupported version |
| RPC rejection | 1 | initialize stage, RPC code -32602 |
| Required keys present | 0 | valid true |
| Missing required keys | 2 | valid false; limit and query missing |
| String for integer schema | 2 | valid false; actual string |

Final line: `PASS: 7 deterministic scenarios; no live model or external MCP server evaluated.`

Measured locally on 12 November 2026: all seven scenarios passed; vet and race tests passed. CI repeats the commands. Timeout, cancellation, malformed-frame and shutdown cases are covered in Go regression tests, not all in this seven-scenario demo.

## Accepted scope and remaining work

This MVP combines separate offline required/type checks with a bounded, version-pinned initialization probe and JSON reports. It does not fetch tool schemas from a server or invoke tools. It is not full JSON Schema or MCP conformance, authentication testing, recovery testing, production isolation or an AI-quality benchmark. The deterministic fixture shares framing code with the probe; passing it is integration evidence, not independent interoperability proof. External server interoperability and tool-call validation remain future work.

The demo builds binaries locally; installers, releases and hosted deployment are outside this scoped MVP. A LinkedIn draft is saved beside this document and has not been published.
