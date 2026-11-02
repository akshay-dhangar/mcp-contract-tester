package core

import (
	"reflect"
	"testing"
)

func TestRequired(t *testing.T) {
	got := MissingRequired([]string{"z", "a", "z"}, map[string]any{})
	if !reflect.DeepEqual(got, []string{"a", "z"}) {
		t.Fatal(got)
	}
	if e := ValidateRequired([]string{"query"}, map[string]any{"query": nil}); e != nil {
		t.Fatal("presence is distinct from null validation")
	}
	if e := ValidateRequired([]string{"query"}, nil); e == nil {
		t.Fatal("missing accepted")
	}
}
