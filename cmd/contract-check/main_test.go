package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestExitCodes(t *testing.T) {
	for _, tc := range []struct {
		in   string
		code int
	}{{`{"required":[],"arguments":{}}`, 0}, {`{"required":["x"],"arguments":{}}`, 2}, {`{}`, 1}} {
		var out, diag bytes.Buffer
		if got := run(strings.NewReader(tc.in), &out, &diag); got != tc.code {
			t.Fatalf("exit %d want %d", got, tc.code)
		}
		if tc.code == 1 && out.Len() != 0 {
			t.Fatal("invalid input produced report")
		}
	}
}
