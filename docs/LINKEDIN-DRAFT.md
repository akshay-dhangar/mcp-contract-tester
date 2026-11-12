# LinkedIn draft — not published

I built a small Go toolkit to make MCP initialization failures easier to reproduce.

It checks a local server's initialization response, bounds the subprocess lifecycle with a timeout, and emits JSON reports identifying the failure stage with a suggested next step. Separate offline commands check required arguments and supported JSON type constraints.

The reproducible demo builds the CLIs and checks seven deterministic scenarios: a successful handshake, wrong ID, unsupported protocol version, RPC rejection, present/missing required arguments and a type mismatch. Local vet and race checks passed too.

The boundary matters: this is a scoped local MVP, not a complete MCP conformance suite. Tool calls, full schema validation and independent external-server interoperability are still future work.

Code and reproduction instructions: https://github.com/akshay-dhangar/mcp-contract-tester
