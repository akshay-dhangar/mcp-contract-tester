package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	core "github.com/akshay-dhangar/mcp-contract-tester"
)

type report struct {
	Version int                  `json:"report_version"`
	Check   string               `json:"check"`
	Passed  bool                 `json:"passed"`
	Server  *core.ServerIdentity `json:"server,omitempty"`
	Stage   string               `json:"stage,omitempty"`
	Code    string               `json:"code,omitempty"`
	Message string               `json:"message,omitempty"`
	Hint    string               `json:"hint,omitempty"`
}

func makeReport(id core.ServerIdentity, err error) report {
	r := report{Version: 1, Check: "stdio_initialization", Passed: err == nil}
	if err == nil {
		r.Server = &id
		return r
	}
	r.Stage = "unknown"
	var pe *core.ProbeError
	if errors.As(err, &pe) {
		r.Stage = pe.Stage
	}
	r.Code = "probe_failed"
	if errors.Is(err, context.DeadlineExceeded) {
		r.Code = "deadline_exceeded"
	}
	if errors.Is(err, context.Canceled) {
		r.Code = "canceled"
	}
	r.Message = err.Error()
	switch r.Stage {
	case "configuration":
		r.Hint = "Set timeout greater than zero and at most one minute."
	case "startup":
		r.Hint = "Check executable path, permissions, and runtime dependencies."
	case "initialize":
		r.Hint = "Check newline-delimited JSON-RPC on stdout, response ID 1, protocol 2025-11-25, capabilities, and server identity."
	case "shutdown":
		r.Hint = "Ensure the server exits successfully when stdin closes within the total timeout."
	default:
		r.Hint = "Inspect the diagnostic and rerun the trusted server locally."
	}
	return r
}

func run(ctx context.Context, args []string, out, diagnostics io.Writer) int {
	flags := flag.NewFlagSet("mcp-probe", flag.ContinueOnError)
	flags.SetOutput(diagnostics)
	timeout := flags.Duration("timeout", 5*time.Second, "total handshake and shutdown deadline")
	jsonReport := flags.Bool("report", false, "emit a versioned JSON report on success or failure")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if flags.NArg() == 0 {
		fmt.Fprintln(diagnostics, "usage: mcp-probe [-report] [-timeout 5s] -- executable [args...]")
		return 1
	}
	identity, err := core.Probe(ctx, *timeout, flags.Arg(0), flags.Args()[1:]...)
	var output any = identity
	if *jsonReport {
		output = makeReport(identity, err)
	}
	if err == nil || *jsonReport {
		if writeErr := json.NewEncoder(out).Encode(output); writeErr != nil {
			fmt.Fprintln(diagnostics, writeErr)
			return 1
		}
	}
	if err != nil {
		if !*jsonReport {
			fmt.Fprintln(diagnostics, err)
		}
		return 1
	}
	return 0
}

func main() { os.Exit(run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)) }
