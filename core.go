package core

import (
	"fmt"
	"sort"
)

// MissingRequired is a small contract check, not a complete JSON Schema validator.
func MissingRequired(required []string, args map[string]any) []string {
	missing := []string{}
	seen := map[string]bool{}
	for _, key := range required {
		if _, exists := args[key]; !exists && !seen[key] {
			missing = append(missing, key)
			seen[key] = true
		}
	}
	sort.Strings(missing)
	return missing
}
func ValidateRequired(required []string, args map[string]any) error {
	if missing := MissingRequired(required, args); len(missing) > 0 {
		return fmt.Errorf("missing required arguments: %v", missing)
	}
	return nil
}
