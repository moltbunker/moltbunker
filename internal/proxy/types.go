package proxy

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Config configures the proxy server.
type Config struct {
	SOCKS5Addr  string // SOCKS5 listen address (e.g., ":1080")
	HTTPAddr    string // HTTP proxy listen address (e.g., ":8118")
	UseTor      bool   // Route through Tor by default
	MaxSessions int    // Max concurrent proxy sessions
}

// DefaultConfig returns sensible proxy defaults.
func DefaultConfig() Config {
	return Config{
		SOCKS5Addr:  ":1080",
		HTTPAddr:    ":8118",
		UseTor:      false,
		MaxSessions: 1000,
	}
}

// Session tracks an active proxy connection.
type Session struct {
	ID        string    `json:"id"`
	Wallet    string    `json:"wallet"`
	Protocol  string    `json:"protocol"` // "socks5", "http_connect", "http_forward"
	Target    string    `json:"target"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
	StartedAt time.Time `json:"started_at"`
	UseTor    bool      `json:"use_tor"`
}

// UsageReport summarizes bandwidth usage for a wallet.
type UsageReport struct {
	Wallet       string `json:"wallet"`
	TotalIn      int64  `json:"total_bytes_in"`
	TotalOut     int64  `json:"total_bytes_out"`
	SessionCount int    `json:"session_count"`
}

// SessionTracker manages active proxy sessions with concurrency control.
type SessionTracker struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	max      int
}

// NewSessionTracker creates a tracker with a max session limit.
func NewSessionTracker(max int) *SessionTracker {
	return &SessionTracker{
		sessions: make(map[string]*Session),
		max:      max,
	}
}

// Add registers a new session. Returns false if limit reached.
func (t *SessionTracker) Add(s *Session) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.sessions) >= t.max {
		return false
	}
	t.sessions[s.ID] = s
	return true
}

// Remove removes a session by ID.
func (t *SessionTracker) Remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.sessions, id)
}

// Get retrieves a session by ID.
func (t *SessionTracker) Get(id string) (*Session, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.sessions[id]
	return s, ok
}

// List returns all active sessions.
func (t *SessionTracker) List() []*Session {
	t.mu.RLock()
	defer t.mu.RUnlock()
	result := make([]*Session, 0, len(t.sessions))
	for _, s := range t.sessions {
		result = append(result, s)
	}
	return result
}

// Count returns the number of active sessions.
func (t *SessionTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.sessions)
}

// BandwidthMeter wraps an io.Reader or io.Writer to count bytes transferred.
type BandwidthMeter struct {
	bytesRead    atomic.Int64
	bytesWritten atomic.Int64
}

// NewBandwidthMeter creates a new bandwidth meter.
func NewBandwidthMeter() *BandwidthMeter {
	return &BandwidthMeter{}
}

// WrapReader wraps a net.Conn to track reads.
func (m *BandwidthMeter) WrapReader(conn net.Conn) *MeteredConn {
	return &MeteredConn{Conn: conn, meter: m}
}

// BytesRead returns total bytes read.
func (m *BandwidthMeter) BytesRead() int64 {
	return m.bytesRead.Load()
}

// BytesWritten returns total bytes written.
func (m *BandwidthMeter) BytesWritten() int64 {
	return m.bytesWritten.Load()
}

// MeteredConn wraps a net.Conn with byte counting.
type MeteredConn struct {
	net.Conn
	meter *BandwidthMeter
}

// Read counts bytes read.
func (c *MeteredConn) Read(b []byte) (int, error) {
	n, err := c.Conn.Read(b)
	c.meter.bytesRead.Add(int64(n))
	return n, err
}

// Write counts bytes written.
func (c *MeteredConn) Write(b []byte) (int, error) {
	n, err := c.Conn.Write(b)
	c.meter.bytesWritten.Add(int64(n))
	return n, err
}
