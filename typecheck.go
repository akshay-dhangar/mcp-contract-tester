package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"strconv"
	"strings"
)

type TypeReport struct {
	Check    string   `json:"check"`
	Valid    bool     `json:"valid"`
	Expected []string `json:"expected"`
	Actual   string   `json:"actual"`
}

// CheckTypeJSON supports only the type keyword. Other schema keywords fail closed.
func CheckTypeJSON(r io.Reader) (TypeReport, error) {
	data, err := io.ReadAll(io.LimitReader(r, (1<<20)+1))
	if err != nil {
		return TypeReport{}, err
	}
	if len(data) > 1<<20 {
		return TypeReport{}, fmt.Errorf("input exceeds 1 MiB")
	}
	var input struct {
		Schema map[string]json.RawMessage `json:"schema"`
		Value  json.RawMessage            `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(data))
	d.DisallowUnknownFields()
	if d.Decode(&input) != nil {
		return TypeReport{}, fmt.Errorf("expected schema and value")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return TypeReport{}, fmt.Errorf("expected one JSON object")
	}
	if len(input.Schema) != 1 || input.Schema["type"] == nil || len(input.Value) == 0 {
		return TypeReport{}, fmt.Errorf("schema must contain only type; value must be present")
	}
	var types []string
	var single string
	if json.Unmarshal(input.Schema["type"], &single) == nil && single != "" {
		types = []string{single}
	} else if json.Unmarshal(input.Schema["type"], &types) != nil || len(types) == 0 {
		return TypeReport{}, fmt.Errorf("type must be a name or nonempty array of names")
	}
	seen := map[string]bool{}
	for _, name := range types {
		switch name {
		case "string", "number", "integer", "boolean", "null", "object", "array":
		default:
			return TypeReport{}, fmt.Errorf("unsupported type name")
		}
		if seen[name] {
			return TypeReport{}, fmt.Errorf("duplicate type name")
		}
		seen[name] = true
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(input.Value))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil {
		return TypeReport{}, fmt.Errorf("invalid value")
	}
	actual := "null"
	switch v := value.(type) {
	case string:
		actual = "string"
	case bool:
		actual = "boolean"
	case []any:
		actual = "array"
	case map[string]any:
		actual = "object"
	case json.Number:
		s := v.String()
		if len(s) > 256 {
			return TypeReport{}, fmt.Errorf("number exceeds supported precision limit")
		}
		if pos := strings.IndexAny(s, "eE"); pos >= 0 {
			exp, e := strconv.Atoi(s[pos+1:])
			if e != nil || exp < -1000 || exp > 1000 {
				return TypeReport{}, fmt.Errorf("number exponent exceeds supported limit")
			}
		}
		n, ok := new(big.Rat).SetString(s)
		if !ok {
			return TypeReport{}, fmt.Errorf("unsupported numeric representation")
		}
		actual = "number"
		if n.IsInt() {
			actual = "integer"
		}
	}
	return TypeReport{Check: "type_keyword_only", Valid: seen[actual] || (actual == "integer" && seen["number"]), Expected: types, Actual: actual}, nil
}
