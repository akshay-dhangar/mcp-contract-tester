package core

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestProbe(t *testing.T) {
	script := `read -r req || exit 1
 case "$req" in *'"method":"initialize"'*) ;; *) exit 1;; esac
 echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{},"serverInfo":{"name":"fixture","version":"1"}}}'
 read -r ack || exit 1
 case "$ack" in *notifications/initialized*) ;; *) exit 1;; esac
 if read -r extra; then exit 1; fi`
	id, err := Probe(context.Background(), 3*time.Second, "sh", "-c", script)
	if err != nil || id.Name != "fixture" {
		t.Fatalf("%+v %v", id, err)
	}
}
func TestProbeFailure(t *testing.T) {
	_, err := Probe(context.Background(), 3*time.Second, "sh", "-c", `read -r req; echo '{"jsonrpc":"2.0","id":2,"result":{}}'`)
	if err == nil || err.Error() != "initialize: initialize response ID mismatch" {
		t.Fatal(err)
	}
	_, err = Probe(context.Background(), 100*time.Millisecond, "sh", "-c", "read -r req; read -r blocked")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestInitializeFailures(t *testing.T) {
	cases := []struct{ name, reply, want string }{
		{"rpc", `{"jsonrpc":"2.0","id":1,"error":{"code":-32602,"message":"SECRET","data":"SECRET"}}`, "RPC error -32602"},
		{"version", `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"unknown"}}`, "unsupported protocol version"},
		{"missing identity", `{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{}}}`, "requires capabilities"},
		{"null result", `{"jsonrpc":"2.0","id":1,"result":null}`, "unsupported protocol version"},
		{"wrong result type", `{"jsonrpc":"2.0","id":1,"result":[]}`, "invalid initialize reply"},
		{"malformed", `not-json`, "read initialize reply"},
		{"notification first", `{"jsonrpc":"2.0","method":"notifications/message"}`, "ID mismatch"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out strings.Builder
			_, err := initialize(&out, NewFrameReader(strings.NewReader(tc.reply+"\n")))
			if err == nil || !strings.Contains(err.Error(), tc.want) || strings.Contains(err.Error(), "SECRET") {
				t.Fatalf("got %v; want %s", err, tc.want)
			}
			if strings.Contains(out.String(), "notifications/initialized") {
				t.Fatal("acknowledged rejected reply")
			}
		})
	}
}

func TestProbeLifecycleFailures(t *testing.T) {
	valid := `echo '{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2025-11-25","capabilities":{},"serverInfo":{"name":"fixture","version":"1"}}}'`
	cases := []struct {
		name, script, stage string
		timeout             time.Duration
		deadline            bool
	}{
		{"early EOF", "exit 0", "initialize", 3 * time.Second, false},
		{"nonzero exit", "read -r req; " + valid + "; read -r ack; exit 7", "shutdown", 3 * time.Second, false},
		{"shutdown hang", "read -r req; " + valid + "; read -r ack; while :; do :; done", "shutdown", time.Second, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Probe(context.Background(), tc.timeout, "sh", "-c", tc.script)
			var pe *ProbeError
			if !errors.As(err, &pe) || pe.Stage != tc.stage {
				t.Fatalf("got %v; want stage %s", err, tc.stage)
			}
			if tc.deadline && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatal(err)
			}
		})
	}
	_, err := Probe(context.Background(), time.Second, filepath.Join(t.TempDir(), "missing-server"))
	var pe *ProbeError
	if !errors.As(err, &pe) || pe.Stage != "startup" {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = Probe(ctx, time.Second, "sh", "-c", "exit 0")
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	for _, timeout := range []time.Duration{0, -time.Second, time.Minute + time.Second} {
		_, err = Probe(context.Background(), timeout, "sh", "-c", "exit 0")
		if !errors.As(err, &pe) || pe.Stage != "configuration" {
			t.Fatal(err)
		}
	}
}
