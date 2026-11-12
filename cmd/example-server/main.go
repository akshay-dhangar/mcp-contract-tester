// example-server is a deterministic handshake fixture, not a production MCP server.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	core "github.com/akshay-dhangar/mcp-contract-tester"
	"io"
	"os"
)

func serve(in io.Reader, out io.Writer, mode string) error {
	if mode != "success" && mode != "wrong-id" && mode != "wrong-version" && mode != "rpc-error" {
		return fmt.Errorf("unknown fixture mode")
	}
	reader := core.NewFrameReader(in)
	frame, err := reader.Read()
	if err != nil {
		return err
	}
	var request struct {
		ID     json.RawMessage
		Method string
	}
	if err = json.Unmarshal(frame, &request); err != nil {
		return err
	}
	if request.Method != "initialize" || string(request.ID) != "1" {
		return fmt.Errorf("expected probe initialize request with ID 1")
	}
	id := 1
	version := "2025-11-25"
	if mode == "wrong-id" {
		id = 2
	}
	if mode == "wrong-version" {
		version = "unsupported-fixture-version"
	}
	reply := map[string]any{"jsonrpc": "2.0", "id": id, "result": map[string]any{"protocolVersion": version, "capabilities": map[string]any{}, "serverInfo": core.ServerIdentity{Name: "deterministic-fixture", Version: "1"}}}
	if mode == "rpc-error" {
		delete(reply, "result")
		reply["error"] = map[string]any{"code": -32602, "message": "fixture rejection"}
	}
	data, err := json.Marshal(reply)
	if err != nil {
		return err
	}
	if err = core.WriteFrame(out, data); err != nil {
		return err
	}
	if mode != "success" {
		return nil
	}
	frame, err = reader.Read()
	if err != nil {
		return err
	}
	var ack struct {
		Method string
		ID     json.RawMessage
	}
	if err = json.Unmarshal(frame, &ack); err != nil {
		return err
	}
	if ack.Method != "notifications/initialized" || ack.ID != nil {
		return fmt.Errorf("expected initialized notification")
	}
	_, err = reader.Read()
	if err != io.EOF {
		return fmt.Errorf("expected stdin EOF after initialized notification")
	}
	return nil
}
func main() {
	mode := flag.String("mode", "success", "success, wrong-id, wrong-version, rpc-error")
	flag.Parse()
	if err := serve(os.Stdin, os.Stdout, *mode); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
