package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// ExecSession tracks an active WebSocket exec session
type ExecSession struct {
	SessionID     string    `json:"session_id"`
	ContainerID   string    `json:"container_id"`
	WalletAddress string    `json:"wallet_address"`
	CreatedAt     time.Time `json:"created_at"`
	LastActivity  time.Time `json:"last_activity"`
	BytesSent     int64     `json:"bytes_sent"`
	BytesReceived int64     `json:"bytes_received"`
}

// ExecChallenge is a single-use nonce for exec authentication
type ExecChallenge struct {
	Nonce       string
	ContainerID string
	Address     string
	Message     string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	Used        bool
}

// ExecSessionManager manages active exec WebSocket sessions
type ExecSessionManager struct {
	sessions   map[string]*ExecSession   // sessionID → session
	challenges map[string]*ExecChallenge // nonce → challenge

	mu     sync.RWMutex
	cancel context.CancelFunc

	// Limits
	maxSessionsPerWallet    int
	maxSessionsPerContainer int
	maxTotalSessions        int
	maxTotalChallenges      int
	idleTimeout             time.Duration
	challengeTimeout        time.Duration
}

// NewExecSessionManager creates a new exec session manager
func NewExecSessionManager() *ExecSessionManager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &ExecSessionManager{
		sessions:                make(map[string]*ExecSession),
		challenges:              make(map[string]*ExecChallenge),
		cancel:                  cancel,
		maxSessionsPerWallet:    5,
		maxSessionsPerContainer: 3,
		maxTotalSessions:        50,
		maxTotalChallenges:      200,
		idleTimeout:             30 * time.Minute,
		challengeTimeout:        30 * time.Second,
	}

	go m.reapLoop(ctx)
	return m
}

// Close stops the reaper goroutine
func (m *ExecSessionManager) Close() {
	m.cancel()
}

// CreateChallenge generates a single-use exec challenge for wallet signing
func (m *ExecSessionManager) CreateChallenge(containerID, walletAddress string) (*ExecChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Cap total challenges to prevent memory exhaustion from rapid requests
	if len(m.challenges) >= m.maxTotalChallenges {
		return nil, fmt.Errorf("too many pending challenges, try again shortly")
	}

	// Generate random nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)

	now := time.Now()
	challenge := &ExecChallenge{
		Nonce:       nonce,
		ContainerID: containerID,
		Address:     walletAddress,
		Message: fmt.Sprintf(
			"Authorize terminal access to container.\n\nContainer: %s\nWallet: %s\nNonce: %s\nTimestamp: %d",
			containerID, walletAddress, nonce, now.Unix(),
		),
		CreatedAt: now,
		ExpiresAt: now.Add(m.challengeTimeout),
	}

	m.challenges[nonce] = challenge

	logging.Debug("exec challenge created",
		"container_id", containerID,
		"wallet", walletAddress,
		"nonce", nonce[:16],
		logging.Component("exec_sessions"))

	return challenge, nil
}

// ValidateChallenge checks and consumes a challenge nonce. Returns the challenge if valid.
func (m *ExecSessionManager) ValidateChallenge(nonce string) (*ExecChallenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	challenge, exists := m.challenges[nonce]
	if !exists {
		return nil, fmt.Errorf("unknown challenge nonce")
	}

	if challenge.Used {
		return nil, fmt.Errorf("challenge already consumed")
	}

	if time.Now().After(challenge.ExpiresAt) {
		delete(m.challenges, nonce)
		return nil, fmt.Errorf("challenge expired")
	}

	// Mark as used (single-use nonce)
	challenge.Used = true
	delete(m.challenges, nonce)

	return challenge, nil
}

// AddSession registers a new active exec session
func (m *ExecSessionManager) AddSession(session *ExecSession) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check total limit
	if len(m.sessions) >= m.maxTotalSessions {
		return fmt.Errorf("max total exec sessions (%d) reached", m.maxTotalSessions)
	}

	// Count per-wallet and per-container
	walletCount := 0
	containerCount := 0
	for _, s := range m.sessions {
		if s.WalletAddress == session.WalletAddress {
			walletCount++
		}
		if s.ContainerID == session.ContainerID {
			containerCount++
		}
	}

	if walletCount >= m.maxSessionsPerWallet {
		return fmt.Errorf("max exec sessions per wallet (%d) reached", m.maxSessionsPerWallet)
	}
	if containerCount >= m.maxSessionsPerContainer {
		return fmt.Errorf("max exec sessions per container (%d) reached", m.maxSessionsPerContainer)
	}

	m.sessions[session.SessionID] = session
	return nil
}

// RemoveSession removes an exec session
func (m *ExecSessionManager) RemoveSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, sessionID)
}

// GetSession returns a session by ID
func (m *ExecSessionManager) GetSession(sessionID string) (*ExecSession, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.sessions[sessionID]
	return s, ok
}

// TouchSession updates the last activity timestamp
func (m *ExecSessionManager) TouchSession(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.LastActivity = time.Now()
	}
}

// AddBytes tracks bytes transferred for a session
func (m *ExecSessionManager) AddBytes(sessionID string, sent, received int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[sessionID]; ok {
		s.BytesSent += sent
		s.BytesReceived += received
		s.LastActivity = time.Now()
	}
}

// Count returns the number of active sessions
func (m *ExecSessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.sessions)
}

// reapLoop periodically removes idle sessions and expired challenges
func (m *ExecSessionManager) reapLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		m.mu.Lock()
		now := time.Now()

		// Reap idle sessions
		for id, s := range m.sessions {
			if now.Sub(s.LastActivity) > m.idleTimeout {
				logging.Info("reaping idle exec session",
					"session_id", id,
					"container_id", s.ContainerID,
					"idle_minutes", int(now.Sub(s.LastActivity).Minutes()),
					logging.Component("exec_sessions"))
				delete(m.sessions, id)
			}
		}

		// Reap expired challenges
		for nonce, c := range m.challenges {
			if now.After(c.ExpiresAt) {
				delete(m.challenges, nonce)
			}
		}

		m.mu.Unlock()
	}
}

// GenerateSessionID creates a unique session ID for exec
func GenerateSessionID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "exec_" + hex.EncodeToString(b), nil
}
