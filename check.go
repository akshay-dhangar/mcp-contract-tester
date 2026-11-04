package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// RequiredContract deliberately describes presence only, not JSON Schema validity.
type RequiredContract struct {
	Required  []string       `json:"required"`
	Arguments map[string]any `json:"arguments"`
}
type RequiredReport struct {
	Check   string   `json:"check"`
	Valid   bool     `json:"valid"`
	Missing []string `json:"missing"`
}

func CheckRequiredJSON(r io.Reader) (RequiredReport, error) {
	data, err := io.ReadAll(io.LimitReader(r, (1<<20)+1))
	if err != nil {
		return RequiredReport{}, err
	}
	if len(data) > 1<<20 {
		return RequiredReport{}, fmt.Errorf("input exceeds 1 MiB")
	}
	var input RequiredContract
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if decoder.Decode(&input) != nil {
		return RequiredReport{}, fmt.Errorf("expected required array and arguments object")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return RequiredReport{}, fmt.Errorf("expected one JSON object")
	}
	if input.Required == nil || input.Arguments == nil {
		return RequiredReport{}, fmt.Errorf("required array and arguments object must be present and non-null")
	}
	missing := MissingRequired(input.Required, input.Arguments)
	return RequiredReport{Check: "required_presence_only", Valid: len(missing) == 0, Missing: missing}, nil
}
