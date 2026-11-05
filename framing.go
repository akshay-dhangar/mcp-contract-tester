package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"unicode/utf8"
)

const MaxFrameBytes = 1 << 20

// FrameReader handles newline-delimited JSON-RPC envelopes, not MCP lifecycle.
// After any error, discard this reader; framing may no longer be synchronized.
type FrameReader struct{ reader *bufio.Reader }

func NewFrameReader(r io.Reader) *FrameReader { return &FrameReader{bufio.NewReader(r)} }
func (f *FrameReader) Read() (json.RawMessage, error) {
	var frame []byte
	for {
		part, err := f.reader.ReadSlice('\n')
		if len(frame)+len(part) > MaxFrameBytes+1 {
			return nil, fmt.Errorf("frame exceeds 1 MiB")
		}
		frame = append(frame, part...)
		if err == bufio.ErrBufferFull {
			continue
		}
		if err != nil {
			if err == io.EOF && len(frame) > 0 {
				return nil, fmt.Errorf("unterminated frame")
			}
			return nil, err
		}
		frame = frame[:len(frame)-1]
		if err = validateEnvelope(frame); err != nil {
			return nil, err
		}
		return json.RawMessage(frame), nil
	}
}
func WriteFrame(w io.Writer, frame json.RawMessage) error {
	if err := validateEnvelope(frame); err != nil {
		return err
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, frame); err != nil {
		return err
	}
	compact.WriteByte('\n')
	n, err := w.Write(compact.Bytes())
	if err == nil && n != compact.Len() {
		return io.ErrShortWrite
	}
	return err
}
func validateEnvelope(frame []byte) error {
	if len(frame) > MaxFrameBytes || !utf8.Valid(frame) || bytes.ContainsAny(frame, "\r\n") {
		return fmt.Errorf("invalid frame encoding, length or embedded newline")
	}
	var fields map[string]json.RawMessage
	if json.Unmarshal(frame, &fields) != nil || fields == nil {
		return fmt.Errorf("expected JSON-RPC object")
	}
	var version string
	if json.Unmarshal(fields["jsonrpc"], &version) != nil || version != "2.0" {
		return fmt.Errorf("expected JSON-RPC 2.0")
	}
	_, method := fields["method"]
	id, hasID := fields["id"]
	_, result := fields["result"]
	_, rpcError := fields["error"]
	if hasID {
		var v any
		d := json.NewDecoder(bytes.NewReader(id))
		d.UseNumber()
		if d.Decode(&v) != nil {
			return fmt.Errorf("invalid id")
		}
		switch v.(type) {
		case string, json.Number:
		default:
			return fmt.Errorf("id must be string or number")
		}
	}
	if method {
		var name string
		if json.Unmarshal(fields["method"], &name) != nil || name == "" || result || rpcError {
			return fmt.Errorf("invalid request/notification")
		}
		if p, ok := fields["params"]; ok {
			var v any
			if json.Unmarshal(p, &v) != nil {
				return fmt.Errorf("invalid params")
			}
			switch v.(type) {
			case map[string]any, []any:
			default:
				return fmt.Errorf("params must be structured")
			}
		}
	} else {
		if !hasID || result == rpcError {
			return fmt.Errorf("response requires id and exactly one result/error")
		}
		if rpcError {
			var e struct {
				Code    *int    `json:"code"`
				Message *string `json:"message"`
			}
			if json.Unmarshal(fields["error"], &e) != nil || e.Code == nil || e.Message == nil {
				return fmt.Errorf("invalid error response")
			}
		}
	}
	return nil
}
