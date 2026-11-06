package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"time"
)

type ServerIdentity struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Probe starts a trusted executable and completes one initialization exchange.
// The deadline covers startup, handshake and shutdown.
func Probe(parent context.Context, timeout time.Duration, executable string, args ...string) (ServerIdentity, error) {
	if timeout <= 0 || timeout > time.Minute {
		return ServerIdentity{}, fmt.Errorf("timeout must be positive and at most one minute")
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ServerIdentity{}, err
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ServerIdentity{}, err
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return ServerIdentity{}, err
	}
	// Closing both ends also interrupts blocked I/O when descendants inherit a pipe.
	stop := context.AfterFunc(ctx, func() { stdin.Close(); stdout.Close() })
	defer stop()
	identity, exchangeErr := initialize(stdin, NewFrameReader(stdout))
	stdin.Close()
	if exchangeErr != nil {
		_ = cmd.Process.Kill()
	}
	waitErr := cmd.Wait()
	if ctx.Err() != nil {
		return ServerIdentity{}, ctx.Err()
	}
	if exchangeErr != nil {
		return ServerIdentity{}, exchangeErr
	}
	if waitErr != nil {
		return ServerIdentity{}, fmt.Errorf("server exit: %w", waitErr)
	}
	return identity, nil
}

func initialize(out io.Writer, in *FrameReader) (ServerIdentity, error) {
	const request = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"mcp-contract-tester","version":"0.1"}}}`
	if err := WriteFrame(out, []byte(request)); err != nil {
		return ServerIdentity{}, err
	}
	frame, err := in.Read()
	if err != nil {
		return ServerIdentity{}, fmt.Errorf("read initialize reply: %w", err)
	}
	var reply struct {
		ID     json.RawMessage
		Result struct {
			ProtocolVersion string
			Capabilities    map[string]json.RawMessage
			ServerInfo      ServerIdentity
		}
		Error *struct{ Code int }
	}
	if err = json.Unmarshal(frame, &reply); err != nil {
		return ServerIdentity{}, fmt.Errorf("invalid initialize reply")
	}
	if string(reply.ID) != "1" {
		return ServerIdentity{}, fmt.Errorf("initialize response ID mismatch")
	}
	if reply.Error != nil {
		return ServerIdentity{}, fmt.Errorf("initialize RPC error %d", reply.Error.Code)
	}
	r := reply.Result
	if r.ProtocolVersion != "2025-11-25" || r.Capabilities == nil || r.ServerInfo.Name == "" || r.ServerInfo.Version == "" {
		return ServerIdentity{}, fmt.Errorf("incompatible or incomplete initialize result")
	}
	if err = WriteFrame(out, []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)); err != nil {
		return ServerIdentity{}, err
	}
	return r.ServerInfo, nil
}
