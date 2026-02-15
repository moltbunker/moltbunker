package p2p

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// sendLengthPrefixed sends data with a 4-byte length prefix
func (r *Router) sendLengthPrefixed(conn *tls.Conn, data []byte) error {
	if conn == nil {
		return fmt.Errorf("connection is nil")
	}

	// Set write deadline
	conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// Write length prefix (4 bytes, big endian)
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}

	// Write data
	if _, err := conn.Write(data); err != nil {
		return err
	}

	return nil
}

// WriteLengthPrefixed writes data with a 4-byte length prefix to any writer.
// This is the public counterpart of Router.sendLengthPrefixed.
func WriteLengthPrefixed(conn io.Writer, data []byte) error {
	lengthBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))

	if _, err := conn.Write(lengthBuf); err != nil {
		return err
	}
	if _, err := conn.Write(data); err != nil {
		return err
	}
	return nil
}

// ReadLengthPrefixed reads data with a 4-byte length prefix
func ReadLengthPrefixed(conn io.Reader) ([]byte, error) {
	// Read length prefix
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)

	// Sanity check - max message size 16MB
	if length > 16*1024*1024 {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}

	// Read data
	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}
