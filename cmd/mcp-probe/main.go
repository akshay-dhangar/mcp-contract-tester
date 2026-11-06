package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	core "github.com/akshay-dhangar/mcp-contract-tester"
	"os"
	"time"
)

func main() {
	timeout := flag.Duration("timeout", 5*time.Second, "total handshake and shutdown deadline")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: mcp-probe [-timeout 5s] -- executable [args...]")
		os.Exit(1)
	}
	identity, err := core.Probe(context.Background(), *timeout, flag.Arg(0), flag.Args()[1:]...)
	if err == nil {
		err = json.NewEncoder(os.Stdout).Encode(identity)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
