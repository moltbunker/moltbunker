//go:build linux

package main

import (
	"encoding/binary"
	"fmt"
	"io"
)

// Frame type constants for the exec agent wire protocol.
// These match the WebSocket frame types in internal/api/exec_handler.go.
const (
	FrameData    byte = 0x01 // Encrypted terminal I/O
	FrameResize  byte = 0x02 // Terminal dimensions change
	FramePing    byte = 0x03 // Keepalive
	FramePong    byte = 0x04 // Keepalive response
	FrameClose   byte = 0x05 // Graceful close
	FrameError   byte = 0x06 // Error message
	FrameKeyInit byte = 0x07 // Session key initialization (carries session_nonce)
	FrameKeyAck  byte = 0x08 // Session key acknowledgment
)

// maxFrameSize is the maximum allowed frame payload (1MB).
const maxFrameSize = 1 << 20

// Frame represents a single protocol frame on the wire.
type Frame struct {
	Type    byte
	Payload []byte
}

// writeFrame writes a length-prefixed frame: [4-byte big-endian length][1-byte type][payload].
func writeFrame(w io.Writer, f *Frame) error {
	totalLen := uint32(1 + len(f.Payload))
	if err := binary.Write(w, binary.BigEndian, totalLen); err != nil {
		return fmt.Errorf("write frame length: %w", err)
	}
	if _, err := w.Write([]byte{f.Type}); err != nil {
		return fmt.Errorf("write frame type: %w", err)
	}
	if len(f.Payload) > 0 {
		if _, err := w.Write(f.Payload); err != nil {
			return fmt.Errorf("write frame payload: %w", err)
		}
	}
	return nil
}

// readFrame reads a length-prefixed frame from a reader.
func readFrame(r io.Reader) (*Frame, error) {
	var totalLen uint32
	if err := binary.Read(r, binary.BigEndian, &totalLen); err != nil {
		return nil, fmt.Errorf("read frame length: %w", err)
	}
	if totalLen < 1 || totalLen > maxFrameSize {
		return nil, fmt.Errorf("invalid frame length: %d", totalLen)
	}

	buf := make([]byte, totalLen)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read frame data: %w", err)
	}

	return &Frame{
		Type:    buf[0],
		Payload: buf[1:],
	}, nil
}
