package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	core "github.com/akshay-dhangar/mcp-contract-tester"
	"testing"
)

func TestReports(t *testing.T) {
	for _, tc := range []struct {
		cause error
		code  string
	}{{context.DeadlineExceeded, "deadline_exceeded"}, {context.Canceled, "canceled"}, {errors.New("bad reply"), "probe_failed"}} {
		r := makeReport(core.ServerIdentity{}, fmt.Errorf("wrapped: %w", &core.ProbeError{Stage: "initialize", Err: tc.cause}))
		if r.Passed || r.Server != nil || r.Stage != "initialize" || r.Code != tc.code || r.Hint == "" {
			t.Fatalf("%+v", r)
		}
	}
}
func TestRunReports(t *testing.T) {
	script := `read -r req; echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{},"serverInfo":{"name":"fixture","version":"1"}}}'; read -r ack`
	for _, tc := range []struct {
		name  string
		args  []string
		code  int
		stage string
	}{
		{"success", []string{"-report", "--", "sh", "-c", script}, 0, ""},
		{"configuration", []string{"-report", "-timeout", "0s", "--", "unused"}, 1, "configuration"},
		{"startup", []string{"-report", "--", "/nonexistent/mcp-fixture"}, 1, "startup"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var out, diag bytes.Buffer
			code := run(context.Background(), tc.args, &out, &diag)
			var r report
			if err := json.Unmarshal(out.Bytes(), &r); err != nil {
				t.Fatal(err, out.String())
			}
			if code != tc.code || r.Passed != (code == 0) || r.Stage != tc.stage || r.Version != 1 || r.Check != "stdio_initialization" || diag.Len() != 0 {
				t.Fatalf("code=%d report=%+v diag=%s", code, r, diag.String())
			}
			if code == 0 && (r.Server == nil || r.Server.Name != "fixture") {
				t.Fatal(r)
			}
		})
	}
	var out, diag bytes.Buffer
	if run(context.Background(), []string{"-report"}, &out, &diag) != 1 || out.Len() != 0 || diag.Len() == 0 {
		t.Fatal("missing command must be usage error")
	}
	out.Reset()
	diag.Reset()
	if run(context.Background(), []string{"--", "sh", "-c", script}, &out, &diag) != 0 {
		t.Fatal(diag.String())
	}
	var identity core.ServerIdentity
	if json.Unmarshal(out.Bytes(), &identity) != nil || identity.Name != "fixture" {
		t.Fatal(out.String())
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, errors.New("output unavailable") }
func TestReportWriteFailure(t *testing.T) {
	var diag bytes.Buffer
	if run(context.Background(), []string{"-report", "-timeout", "0s", "unused"}, brokenWriter{}, &diag) != 1 || diag.Len() == 0 {
		t.Fatal("missing writer error")
	}
}
