package core

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	messages := []string{`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`, `{"jsonrpc":"2.0","method":"notifications/initialized"}`, `{"jsonrpc":"2.0","id":1,"result":{}}`, `{"jsonrpc":"2.0","id":"x","error":{"code":-32601,"message":"unknown"}}`}
	var out bytes.Buffer
	for _, m := range messages {
		if err := WriteFrame(&out, []byte(m)); err != nil {
			t.Fatal(err)
		}
	}
	reader := NewFrameReader(&out)
	for _, want := range messages {
		got, err := reader.Read()
		if err != nil || string(got) != want {
			t.Fatalf("%s %v", got, err)
		}
	}
	if _, err := reader.Read(); err != io.EOF {
		t.Fatal(err)
	}
}
func TestRejectBadFrames(t *testing.T) {
	for _, m := range []string{`[]`, `null`, `{}`, `{"jsonrpc":"1.0","method":"x"}`, `{"jsonrpc":"2.0","id":null,"result":{}}`, `{"jsonrpc":"2.0","id":1,"result":{},"error":{}}`, `{"jsonrpc":"2.0","id":1,"error":{}}`, `{"jsonrpc":"2.0","method":"x","params":null}`, "{\n}", string([]byte{255}), strings.Repeat("x", MaxFrameBytes+1)} {
		if _, err := NewFrameReader(strings.NewReader(m + "\n")).Read(); err == nil {
			t.Fatal("bad frame accepted")
		}
	}
	if _, err := NewFrameReader(strings.NewReader(`{"jsonrpc":"2.0","method":"x"}`)).Read(); err == nil {
		t.Fatal("unterminated frame accepted")
	}
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestShortWrite(t *testing.T) {
	if err := WriteFrame(shortWriter{}, []byte(`{"jsonrpc":"2.0","method":"x"}`)); err != io.ErrShortWrite {
		t.Fatal(err)
	}
}
