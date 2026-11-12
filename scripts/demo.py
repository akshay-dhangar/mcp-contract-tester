#!/usr/bin/env python3
"""Build real CLIs and assert deterministic local outcomes; no external services."""
import json
from pathlib import Path
import subprocess
import tempfile

root = Path(__file__).resolve().parents[1]

def run(args, expected, payload=None):
    result = subprocess.run([str(x) for x in args], input=payload, text=True,
                            capture_output=True, timeout=15, cwd=root)
    assert result.returncode == expected, (args, result.returncode, result.stderr)
    return json.loads(result.stdout)

with tempfile.TemporaryDirectory(prefix="mcp-demo-") as directory:
    binaries = {}
    for command in ("mcp-probe", "example-server", "contract-check", "type-check"):
        binary = Path(directory) / command
        subprocess.run(["go", "build", "-o", str(binary), "./cmd/" + command],
                       cwd=root, check=True, timeout=120)
        binaries[command] = binary
    for mode in ("success", "wrong-id", "wrong-version", "rpc-error"):
        report = run([binaries["mcp-probe"], "-report", "-timeout", "5s", "--",
                      binaries["example-server"], "-mode", mode], 0 if mode == "success" else 1)
        assert report["passed"] == (mode == "success")
        assert report["report_version"] == 1
        if mode == "success":
            assert report["server"]["name"] == "deterministic-fixture"
        else:
            assert report["stage"] == "initialize"
            expected = {"wrong-id": "ID mismatch", "wrong-version": "unsupported protocol version",
                        "rpc-error": "RPC error -32602"}[mode]
            assert expected in report["message"]
        print(json.dumps({"scenario": mode, "report": report}, sort_keys=True))
    for fixture, code in (("present", 0), ("missing", 2)):
        report = run([binaries["contract-check"]], code,
                     (root / "examples" / (fixture + ".json")).read_text())
        assert report["valid"] == (code == 0)
        print(json.dumps({"scenario": fixture, "report": report}, sort_keys=True))
    report = run([binaries["type-check"]], 2,
                 json.dumps({"schema": {"type": "integer"}, "value": "wrong"}))
    assert not report["valid"]
    print(json.dumps({"scenario": "type-mismatch", "report": report}, sort_keys=True))
print("PASS: 7 deterministic scenarios; no live model or external MCP server evaluated.")
