package proxy

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// SOCKS5 protocol constants (RFC 1928).
const (
	socks5Version = 0x05

	// Auth methods
	authNone           = 0x00
	authUsernamePasswd = 0x02
	authNoAcceptable   = 0xFF

	// Commands
	cmdConnect = 0x01

	// Address types
	addrIPv4   = 0x01
	addrDomain = 0x03
	addrIPv6   = 0x04

	// Reply codes
	repSuccess             = 0x00
	repGeneralFailure      = 0x01
	repConnectionNotAllowed = 0x02
	repNetworkUnreachable  = 0x03
	repHostUnreachable     = 0x04
	repConnectionRefused   = 0x05
	repCommandNotSupported = 0x07
	repAddrTypeNotSupported = 0x08
)

// Dialer abstracts how the proxy connects to target hosts.
type Dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

// DirectDialer connects directly to the target.
type DirectDialer struct{}

// DialContext connects directly using the standard net dialer.
func (d *DirectDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return (&net.Dialer{Timeout: 30 * time.Second}).DialContext(ctx, network, address)
}

// Authenticator validates proxy credentials and returns a wallet address.
type Authenticator interface {
	// Authenticate validates username/password and returns the wallet address.
	// Returns empty string if authentication fails.
	Authenticate(username, password string) string
}

// AllowAllAuth accepts any connection with a default wallet.
type AllowAllAuth struct {
	DefaultWallet string
}

// Authenticate always returns the default wallet.
func (a *AllowAllAuth) Authenticate(_, _ string) string {
	return a.DefaultWallet
}

// SOCKS5Server implements a SOCKS5 proxy server (RFC 1928).
type SOCKS5Server struct {
	listener net.Listener
	dialer   Dialer
	auth     Authenticator
	tracker  *SessionTracker
	acl      *ACL

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// NewSOCKS5Server creates a new SOCKS5 proxy server.
func NewSOCKS5Server(dialer Dialer, auth Authenticator, tracker *SessionTracker, acl *ACL) *SOCKS5Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &SOCKS5Server{
		dialer:  dialer,
		auth:    auth,
		tracker: tracker,
		acl:     acl,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// ListenAndServe starts the SOCKS5 server on the given address.
func (s *SOCKS5Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("socks5 listen: %w", err)
	}
	s.listener = ln
	logging.Info("SOCKS5 proxy listening", "addr", addr, logging.Component("proxy"))

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
				logging.Debug("socks5 accept error", "error", err.Error(), logging.Component("proxy"))
				continue
			}
		}

		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConnection(conn)
		}()
	}
}

// Close gracefully shuts down the server.
func (s *SOCKS5Server) Close() error {
	s.cancel()
	if s.listener != nil {
		s.listener.Close()
	}
	s.wg.Wait()
	return nil
}

// handleConnection processes a single SOCKS5 client connection.
func (s *SOCKS5Server) handleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	if err := clientConn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		logging.Debug("socks5 set greeting deadline failed", "error", err.Error(), logging.Component("proxy"))
		return
	}

	// Phase 1: Greeting — client sends supported auth methods
	wallet, err := s.handleGreeting(clientConn)
	if err != nil {
		logging.Debug("socks5 greeting failed", "error", err.Error(), logging.Component("proxy"))
		return
	}

	// Phase 2: Connect request
	if err := clientConn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		logging.Debug("socks5 set connect deadline failed", "error", err.Error(), logging.Component("proxy"))
		return
	}
	target, err := s.handleConnectRequest(clientConn)
	if err != nil {
		logging.Debug("socks5 connect request failed", "error", err.Error(), logging.Component("proxy"))
		return
	}

	// ACL check
	host, _, _ := net.SplitHostPort(target)
	if s.acl != nil && !s.acl.IsAllowed(host) {
		s.sendReply(clientConn, repConnectionNotAllowed, nil)
		return
	}

	// Create session
	sessionID, err := generateSessionID()
	if err != nil {
		logging.Warn("failed to generate socks5 session ID",
			"error", err.Error(),
			logging.Component("proxy"))
		s.sendReply(clientConn, repGeneralFailure, nil)
		return
	}
	session := &Session{
		ID:        sessionID,
		Wallet:    wallet,
		Protocol:  "socks5",
		Target:    target,
		StartedAt: time.Now(),
	}
	if !s.tracker.Add(session) {
		s.sendReply(clientConn, repGeneralFailure, nil)
		return
	}
	defer s.tracker.Remove(sessionID)

	// Connect to target
	targetConn, err := s.dialer.DialContext(s.ctx, "tcp", target)
	if err != nil {
		s.sendConnectError(clientConn, err)
		return
	}
	defer targetConn.Close()

	// Send success reply with bound address
	localAddr := targetConn.LocalAddr().(*net.TCPAddr)
	s.sendReply(clientConn, repSuccess, localAddr)

	// Clear deadline for relay phase
	if err := clientConn.SetDeadline(time.Time{}); err != nil {
		logging.Debug("socks5 clear deadline failed", "error", err.Error(), logging.Component("proxy"))
	}

	// Relay data bidirectionally with bandwidth metering
	meter := NewBandwidthMeter()
	relay(s.ctx, clientConn, targetConn, meter)

	// Update session with final byte counts
	session.BytesIn = meter.BytesRead()
	session.BytesOut = meter.BytesWritten()
}

// handleGreeting processes the SOCKS5 greeting and authenticates the client.
func (s *SOCKS5Server) handleGreeting(conn net.Conn) (string, error) {
	// Read version + nmethods
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", fmt.Errorf("read greeting header: %w", err)
	}
	if header[0] != socks5Version {
		return "", fmt.Errorf("unsupported SOCKS version: %d", header[0])
	}

	// Read method list
	nMethods := int(header[1])
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return "", fmt.Errorf("read methods: %w", err)
	}

	// Check if username/password auth is supported
	hasUserPass := false
	hasNoAuth := false
	for _, m := range methods {
		if m == authUsernamePasswd {
			hasUserPass = true
		}
		if m == authNone {
			hasNoAuth = true
		}
	}

	if hasUserPass {
		// Select username/password auth
		if _, err := conn.Write([]byte{socks5Version, authUsernamePasswd}); err != nil {
			return "", fmt.Errorf("write auth method selection: %w", err)
		}
		return s.handleUserPassAuth(conn)
	}

	if hasNoAuth {
		// No auth — use default wallet from authenticator
		if _, err := conn.Write([]byte{socks5Version, authNone}); err != nil {
			return "", fmt.Errorf("write auth method selection: %w", err)
		}
		wallet := s.auth.Authenticate("", "")
		return wallet, nil
	}

	// No acceptable method
	if _, err := conn.Write([]byte{socks5Version, authNoAcceptable}); err != nil {
		return "", fmt.Errorf("write auth method selection: %w", err)
	}
	return "", fmt.Errorf("no acceptable auth method")
}

// handleUserPassAuth processes RFC 1929 username/password auth.
func (s *SOCKS5Server) handleUserPassAuth(conn net.Conn) (string, error) {
	// Read auth version
	version := make([]byte, 1)
	if _, err := io.ReadFull(conn, version); err != nil {
		return "", fmt.Errorf("read auth version: %w", err)
	}

	// Read username length + username
	ulenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, ulenBuf); err != nil {
		return "", fmt.Errorf("read username length: %w", err)
	}
	ulen := int(ulenBuf[0])
	username := make([]byte, ulen)
	if _, err := io.ReadFull(conn, username); err != nil {
		return "", fmt.Errorf("read username: %w", err)
	}

	// Read password length + password
	plenBuf := make([]byte, 1)
	if _, err := io.ReadFull(conn, plenBuf); err != nil {
		return "", fmt.Errorf("read password length: %w", err)
	}
	plen := int(plenBuf[0])
	password := make([]byte, plen)
	if _, err := io.ReadFull(conn, password); err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	wallet := s.auth.Authenticate(string(username), string(password))
	if wallet == "" {
		// Auth failed
		if _, err := conn.Write([]byte{0x01, 0x01}); err != nil {
			return "", fmt.Errorf("write auth failure: %w", err)
		}
		return "", fmt.Errorf("authentication failed for user %q", string(username))
	}

	// Auth success
	if _, err := conn.Write([]byte{0x01, 0x00}); err != nil {
		return "", fmt.Errorf("write auth success: %w", err)
	}
	return wallet, nil
}

// handleConnectRequest processes the SOCKS5 CONNECT command.
func (s *SOCKS5Server) handleConnectRequest(conn net.Conn) (string, error) {
	// Read VER CMD RSV ATYP
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return "", fmt.Errorf("read request header: %w", err)
	}
	if header[0] != socks5Version {
		return "", fmt.Errorf("invalid version in request: %d", header[0])
	}
	if header[1] != cmdConnect {
		s.sendReply(conn, repCommandNotSupported, nil)
		return "", fmt.Errorf("unsupported command: %d", header[1])
	}

	// Read destination address based on address type
	var host string
	switch header[3] {
	case addrIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("read ipv4 addr: %w", err)
		}
		host = net.IP(addr).String()

	case addrDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return "", fmt.Errorf("read domain length: %w", err)
		}
		domain := make([]byte, int(lenBuf[0]))
		if _, err := io.ReadFull(conn, domain); err != nil {
			return "", fmt.Errorf("read domain: %w", err)
		}
		host = string(domain)

	case addrIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return "", fmt.Errorf("read ipv6 addr: %w", err)
		}
		host = net.IP(addr).String()

	default:
		s.sendReply(conn, repAddrTypeNotSupported, nil)
		return "", fmt.Errorf("unsupported address type: %d", header[3])
	}

	// Read port (2 bytes, big-endian)
	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return "", fmt.Errorf("read port: %w", err)
	}
	port := binary.BigEndian.Uint16(portBuf)

	return fmt.Sprintf("%s:%d", host, port), nil
}

// sendReply sends a SOCKS5 reply to the client.
func (s *SOCKS5Server) sendReply(conn net.Conn, rep byte, bindAddr *net.TCPAddr) {
	reply := []byte{socks5Version, rep, 0x00}

	if bindAddr != nil && bindAddr.IP.To4() != nil {
		reply = append(reply, addrIPv4)
		reply = append(reply, bindAddr.IP.To4()...)
	} else {
		// Default: 0.0.0.0:0
		reply = append(reply, addrIPv4, 0, 0, 0, 0)
	}

	port := make([]byte, 2)
	if bindAddr != nil {
		binary.BigEndian.PutUint16(port, uint16(bindAddr.Port))
	}
	reply = append(reply, port...)
	if _, err := conn.Write(reply); err != nil {
		logging.Debug("socks5 reply write failed",
			"error", err.Error(),
			logging.Component("proxy"))
	}
}

// sendConnectError maps a dial error to an appropriate SOCKS5 reply code.
func (s *SOCKS5Server) sendConnectError(conn net.Conn, err error) {
	errStr := err.Error()
	var rep byte
	switch {
	case strings.Contains(errStr, "refused"):
		rep = repConnectionRefused
	case strings.Contains(errStr, "unreachable"):
		rep = repNetworkUnreachable
	case strings.Contains(errStr, "no such host"):
		rep = repHostUnreachable
	default:
		rep = repGeneralFailure
	}
	s.sendReply(conn, rep, nil)
}

// relay copies data bidirectionally between two connections.
func relay(ctx context.Context, client, target net.Conn, meter *BandwidthMeter) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Wrap both connections with metering
	meteredClient := meter.WrapReader(client)

	var wg sync.WaitGroup
	wg.Add(2)

	// client → target (reads from client counted as BytesIn)
	go func() {
		defer wg.Done()
		// io.Copy errors are expected when the relay is torn down; ignore.
		_, _ = io.Copy(target, meteredClient)
		cancel()
	}()

	// target → client (writes to client counted as BytesWritten)
	meteredTarget := &MeteredConn{Conn: target, meter: meter}
	go func() {
		defer wg.Done()
		// io.Copy errors are expected when the relay is torn down; ignore.
		_, _ = io.Copy(client, meteredTarget)
		cancel()
	}()

	// Wait for context cancellation, then close both sides
	<-ctx.Done()
	// Force-close both sides; errors here just mean the conn is already shut.
	_ = client.SetDeadline(time.Now())
	_ = target.SetDeadline(time.Now())
	wg.Wait()
}

// generateSessionID creates a random session ID.
func generateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}
