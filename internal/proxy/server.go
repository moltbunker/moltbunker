package proxy

import (
	"context"
	"fmt"
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// Server manages the proxy server lifecycle (SOCKS5 + HTTP).
type Server struct {
	cfg     Config
	dialer  Dialer
	auth    Authenticator
	tracker *SessionTracker
	acl     *ACL

	socks5 *SOCKS5Server
	http   *HTTPProxyServer

	mu      sync.Mutex
	running bool
}

// NewServer creates a new proxy server with the given configuration.
func NewServer(cfg Config, dialer Dialer, auth Authenticator) *Server {
	return &Server{
		cfg:     cfg,
		dialer:  dialer,
		auth:    auth,
		tracker: NewSessionTracker(cfg.MaxSessions),
	}
}

// SetACL sets the domain access control list.
func (s *Server) SetACL(acl *ACL) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.acl = acl
}

// Start starts the proxy server (SOCKS5 and/or HTTP listeners).
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return fmt.Errorf("proxy server already running")
	}
	s.running = true
	s.mu.Unlock()

	errCh := make(chan error, 2)

	// Start SOCKS5 server
	if s.cfg.SOCKS5Addr != "" {
		s.socks5 = NewSOCKS5Server(s.dialer, s.auth, s.tracker, s.acl)
		go func() {
			if err := s.socks5.ListenAndServe(s.cfg.SOCKS5Addr); err != nil {
				errCh <- fmt.Errorf("socks5: %w", err)
			}
		}()
		logging.Info("proxy SOCKS5 server started", "addr", s.cfg.SOCKS5Addr, logging.Component("proxy"))
	}

	// Start HTTP proxy server
	if s.cfg.HTTPAddr != "" {
		s.http = NewHTTPProxyServer(s.dialer, s.auth, s.tracker, s.acl)
		go func() {
			if err := s.http.ListenAndServe(s.cfg.HTTPAddr); err != nil {
				errCh <- fmt.Errorf("http proxy: %w", err)
			}
		}()
		logging.Info("proxy HTTP server started", "addr", s.cfg.HTTPAddr, logging.Component("proxy"))
	}

	// Wait for context cancellation or startup error
	select {
	case err := <-errCh:
		s.Stop()
		return err
	case <-ctx.Done():
		return s.Stop()
	}
}

// Stop gracefully shuts down all proxy servers.
func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}
	s.running = false

	var errs []error
	if s.socks5 != nil {
		if err := s.socks5.Close(); err != nil {
			errs = append(errs, err)
		}
		s.socks5 = nil
	}
	if s.http != nil {
		if err := s.http.Close(); err != nil {
			errs = append(errs, err)
		}
		s.http = nil
	}

	if len(errs) > 0 {
		return fmt.Errorf("proxy shutdown errors: %v", errs)
	}

	logging.Info("proxy server stopped", logging.Component("proxy"))
	return nil
}

// Tracker returns the session tracker.
func (s *Server) Tracker() *SessionTracker {
	return s.tracker
}

// IsRunning returns whether the server is running.
func (s *Server) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// Usage returns a usage report for the given wallet, aggregated from active sessions.
func (s *Server) Usage(wallet string) *UsageReport {
	report := &UsageReport{Wallet: wallet}
	for _, session := range s.tracker.List() {
		if session.Wallet == wallet {
			report.TotalIn += session.BytesIn
			report.TotalOut += session.BytesOut
			report.SessionCount++
		}
	}
	return report
}
