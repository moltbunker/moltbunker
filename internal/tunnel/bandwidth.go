package tunnel

import (
	"net"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// BandwidthLimiter enforces per-session bandwidth and request rate limits.
type BandwidthLimiter struct {
	bwLimit int64         // bytes per second
	rps     *rate.Limiter // per-subdomain RPS
}

// NewBandwidthLimiter creates a limiter from tunnel limits.
func NewBandwidthLimiter(limits *TunnelLimits) *BandwidthLimiter {
	return &BandwidthLimiter{
		bwLimit: limits.MaxBandwidth,
		rps:     rate.NewLimiter(rate.Limit(limits.MaxRPS), limits.MaxRPS),
	}
}

// AllowRequest checks if a new request is allowed under the RPS limit.
func (b *BandwidthLimiter) AllowRequest() bool {
	return b.rps.Allow()
}

// WrapConn wraps a net.Conn with bandwidth limiting.
func (b *BandwidthLimiter) WrapConn(conn net.Conn) net.Conn {
	return &rateLimitedConn{
		Conn:        conn,
		bytesPerSec: b.bwLimit,
		bucket:      b.bwLimit, // start with full bucket
		lastFill:    time.Now(),
	}
}

// rateLimitedConn is a net.Conn wrapper that enforces bandwidth limits
// using a token bucket algorithm.
type rateLimitedConn struct {
	net.Conn
	mu          sync.Mutex
	bytesPerSec int64
	bucket      int64
	lastFill    time.Time
}

func (c *rateLimitedConn) Read(b []byte) (int, error) {
	c.refill()
	n, err := c.Conn.Read(b)
	if n > 0 {
		c.mu.Lock()
		c.bucket -= int64(n)
		c.mu.Unlock()
	}
	return n, err
}

func (c *rateLimitedConn) Write(b []byte) (int, error) {
	c.refill()
	// If over budget, sleep proportionally
	c.mu.Lock()
	deficit := -c.bucket
	c.mu.Unlock()

	if deficit > 0 {
		sleepDur := time.Duration(float64(deficit) / float64(c.bytesPerSec) * float64(time.Second))
		if sleepDur > 0 && sleepDur < 5*time.Second {
			time.Sleep(sleepDur)
			c.refill()
		}
	}

	n, err := c.Conn.Write(b)
	if n > 0 {
		c.mu.Lock()
		c.bucket -= int64(n)
		c.mu.Unlock()
	}
	return n, err
}

func (c *rateLimitedConn) refill() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastFill)
	c.lastFill = now

	fill := int64(elapsed.Seconds() * float64(c.bytesPerSec))
	c.bucket += fill
	// Cap at 1 second worth of burst
	if c.bucket > c.bytesPerSec {
		c.bucket = c.bytesPerSec
	}
}
