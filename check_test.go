package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestJSONPresence(t *testing.T) {
	r, e := CheckRequiredJSON(strings.NewReader(`{"required":["z","a","z"],"arguments":{}}`))
	if e != nil || r.Valid || !reflect.DeepEqual(r.Missing, []string{"a", "z"}) {
		t.Fatalf("%+v %v", r, e)
	}
	r, e = CheckRequiredJSON(strings.NewReader(`{"required":["x",""],"arguments":{"x":null,"":false}}`))
	if e != nil || !r.Valid {
		t.Fatalf("%+v %v", r, e)
	}
}
func TestInvalidJSON(t *testing.T) {
	for _, s := range []string{`null`, `{}`, `{"required":[],"arguments":null}`, `{"required":null,"arguments":{}}`, `{"required":[1],"arguments":{}}`, `{"required":[],"arguments":[]}`, `{"required":[],"arguments":{},"secret":"value"}`, `{"required":[],"arguments":{}} {}`, strings.Repeat(" ", (1<<20)+1)} {
		if _, e := CheckRequiredJSON(strings.NewReader(s)); e == nil {
			t.Fatal("invalid input accepted")
		}
	}
}
