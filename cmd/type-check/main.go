package main

import (
	"encoding/json"
	"fmt"
	core "github.com/akshay-dhangar/mcp-contract-tester"
	"os"
)

func main() {
	r, e := core.CheckTypeJSON(os.Stdin)
	if e != nil {
		fmt.Fprintln(os.Stderr, "type-check:", e)
		os.Exit(1)
	}
	if e = json.NewEncoder(os.Stdout).Encode(r); e != nil {
		os.Exit(1)
	}
	if !r.Valid {
		os.Exit(2)
	}
}
