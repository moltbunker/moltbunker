package jsruntime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Message types for the stdio JSON-RPC protocol between Go and Deno workers.
const (
	MsgTypeInvoke     = "invoke"
	MsgTypeHostCall   = "host_call"
	MsgTypeHostResult = "host_result"
	MsgTypeResponse   = "response"
	MsgTypePing       = "ping"
	MsgTypePong       = "pong"
	MsgTypeShutdown   = "shutdown"
)

// Message is the generic envelope for stdio JSON-RPC.
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`

	// For host_call / host_result messages:
	ID   int             `json:"id,omitempty"`
	Fn   string          `json:"fn,omitempty"`
	Args json.RawMessage `json:"args,omitempty"`

	// For response messages:
	Status  int               `json:"status,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Error   string            `json:"error,omitempty"`
}

// InvokeData is the payload for "invoke" messages sent to the Deno worker.
type InvokeData struct {
	ScriptPath string        `json:"script_path"`
	Request    InvokeRequest `json:"request"`
}

// InvokeRequest is the HTTP-like request sent to the JS handler.
type InvokeRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"` // base64-encoded
}

// WriteMessage writes a JSON-line message to a writer.
func WriteMessage(w io.Writer, msg *Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	data = append(data, '\n')
	_, err = w.Write(data)
	return err
}

// ReadMessage reads a single JSON-line message from a buffered reader.
func ReadMessage(r *bufio.Reader) (*Message, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("read message: %w", err)
	}

	var msg Message
	if err := json.Unmarshal(line, &msg); err != nil {
		return nil, fmt.Errorf("unmarshal message: %w", err)
	}

	return &msg, nil
}
