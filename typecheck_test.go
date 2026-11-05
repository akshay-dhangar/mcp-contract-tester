package core

import (
	"strings"
	"testing"
)

func TestTypes(t *testing.T) {
	for _, tc := range []struct {
		schema, value string
		valid         bool
	}{{`"integer"`, `1.0`, true}, {`"integer"`, `1e2`, true}, {`"integer"`, `1.01`, false}, {`"integer"`, `9007199254740992.1`, false}, {`"number"`, `12`, true}, {`"string"`, `12`, false}, {`["string","null"]`, `null`, true}, {`"array"`, `[]`, true}, {`"object"`, `{}`, true}, {`"boolean"`, `false`, true}, {`"null"`, `"null"`, false}} {
		r, e := CheckTypeJSON(strings.NewReader(`{"schema":{"type":` + tc.schema + `},"value":` + tc.value + `}`))
		if e != nil || r.Valid != tc.valid {
			t.Fatalf("%+v => %+v %v", tc, r, e)
		}
	}
}
func TestRejectUnsupportedSchema(t *testing.T) {
	for _, s := range []string{`{}`, `null`, `{"schema":{"type":"string"}}`, `{"schema":{"type":"string","minLength":1},"value":"x"}`, `{"schema":{"type":[]},"value":1}`, `{"schema":{"type":["null","null"]},"value":null}`, `{"schema":{"type":"bogus"},"value":1}`, `{"schema":{"type":"number"},"value":1e9999999}`, `{"schema":{"type":"number"},"value":1} {}`} {
		if _, e := CheckTypeJSON(strings.NewReader(s)); e == nil {
			t.Fatalf("accepted %s", s)
		}
	}
}
