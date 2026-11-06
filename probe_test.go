package core

import (
	"context"
	"errors"
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
	if err == nil || err.Error() != "initialize response ID mismatch" {
		t.Fatal(err)
	}
	_, err = Probe(context.Background(), 100*time.Millisecond, "sh", "-c", "read -r req; read -r blocked")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}
