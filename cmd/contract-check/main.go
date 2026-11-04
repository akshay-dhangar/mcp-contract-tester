package main

import (
	"encoding/json"
	"fmt"
	core "github.com/akshay-dhangar/mcp-contract-tester"
	"io"
	"os"
)

func run(in io.Reader, out, diagnostics io.Writer) int {
	report, err := core.CheckRequiredJSON(in)
	if err != nil {
		fmt.Fprintln(diagnostics, "contract-check:", err)
		return 1
	}
	if err = json.NewEncoder(out).Encode(report); err != nil {
		fmt.Fprintln(diagnostics, "contract-check: output failed")
		return 1
	}
	if !report.Valid {
		return 2
	}
	return 0
}
func main() { os.Exit(run(os.Stdin, os.Stdout, os.Stderr)) }
