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

// ProbeError identifies the lifecycle stage while preserving errors.Is/As.
type ProbeError struct {
	Stage string
	Err   error
}

func (e *ProbeError) Error() string { return e.Stage + ": " + e.Err.Error() }
func (e *ProbeError) Unwrap() error { return e.Err }

func probeError(stage string, err error) error { return &ProbeError{Stage: stage, Err: err} }

// Probe starts a trusted executable and completes one initialization exchange.
// The deadline covers startup, handshake and shutdown.
func Probe(parent context.Context, timeout time.Duration, executable string, args ...string) (ServerIdentity, error) {
	if timeout <= 0 || timeout > time.Minute {
		return ServerIdentity{}, probeError("configuration", fmt.Errorf("timeout must be positive and at most one minute"))
	}
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return ServerIdentity{}, probeError("startup", err)
	}
	defer stdin.Close()
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ServerIdentity{}, probeError("startup", err)
	}
	defer stdout.Close()
	if err = cmd.Start(); err != nil {
		return ServerIdentity{}, probeError("startup", err)
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
		stage := "shutdown"
		if exchangeErr != nil {
			stage = "initialize"
		}
		return ServerIdentity{}, probeError(stage, ctx.Err())
	}
	if exchangeErr != nil {
		return ServerIdentity{}, probeError("initialize", exchangeErr)
	}
	if waitErr != nil {
		return ServerIdentity{}, probeError("shutdown", fmt.Errorf("server exit: %w", waitErr))
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
	if r.ProtocolVersion != "2025-11-25" {
		return ServerIdentity{}, fmt.Errorf("unsupported protocol version; expected 2025-11-25")
	}
	if r.Capabilities == nil || r.ServerInfo.Name == "" || r.ServerInfo.Version == "" {
		return ServerIdentity{}, fmt.Errorf("initialize result requires capabilities and nonempty server name/version")
	}
	if err = WriteFrame(out, []byte(`{"jsonrpc":"2.0","method":"notifications/initialized"}`)); err != nil {
		return ServerIdentity{}, err
	}
	return r.ServerInfo, nil
}
