package p2p

import (
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// PeerEvent represents a scored behavioral event.
type PeerEvent string

const (
	PeerEventValidMessage    PeerEvent = "valid_message"
	PeerEventInvalidMessage  PeerEvent = "invalid_message"
	PeerEventFailedDeploy    PeerEvent = "failed_deploy"
	PeerEventGossipSpam      PeerEvent = "gossip_spam"
	PeerEventRateLimitHit    PeerEvent = "rate_limit_hit"
	PeerEventMalformedPayload PeerEvent = "malformed_payload"
	PeerEventGoodUptime      PeerEvent = "good_uptime"
)

// eventDeltas maps events to their score impact.
var eventDeltas = map[PeerEvent]float64{
	PeerEventValidMessage:    +0.001,
	PeerEventInvalidMessage:  -0.05,
	PeerEventFailedDeploy:    -0.10,
	PeerEventGossipSpam:      -0.03,
	PeerEventRateLimitHit:    -0.02,
	PeerEventMalformedPayload: -0.05,
	PeerEventGoodUptime:      +0.01,
}

// PeerScorerConfig holds scoring thresholds and decay settings.
type PeerScorerConfig struct {
	WarnThreshold     float64       // Score below this → increase logging (default: 0.3)
	ThrottleThreshold float64       // Score below this → 50% rate limit cut (default: 0.2)
	BanThreshold      float64       // Score below this → ban (default: 0.1)
	DecayInterval     time.Duration // How often scores decay toward 0.5 (default: 1h)
	DecayRate         float64       // How much to decay per interval (default: 0.01)
}

// DefaultPeerScorerConfig returns reasonable defaults.
func DefaultPeerScorerConfig() *PeerScorerConfig {
	return &PeerScorerConfig{
		WarnThreshold:     0.3,
		ThrottleThreshold: 0.2,
		BanThreshold:      0.1,
		DecayInterval:     1 * time.Hour,
		DecayRate:         0.01,
	}
}

// PeerScore tracks behavioral metrics for a single peer.
type PeerScore struct {
	Score           float64
	InvalidMessages int64
	FailedDeploys   int64
	GossipSpam      int64
	RateLimitHits   int64
	LastUpdate      time.Time
}

// PeerScorer tracks behavioral patterns per peer. Score starts at 0.5 (neutral),
// ranges from 0.0 (worst) to 1.0 (best).
type PeerScorer struct {
	mu      sync.RWMutex
	scores  map[types.NodeID]*PeerScore
	config  *PeerScorerConfig
	banList *BanList
	stakeVerifier *StakeVerifier
	nowFunc func() time.Time
}

// NewPeerScorer creates a new peer scorer.
func NewPeerScorer(config *PeerScorerConfig) *PeerScorer {
	if config == nil {
		config = DefaultPeerScorerConfig()
	}
	return &PeerScorer{
		scores:  make(map[types.NodeID]*PeerScore),
		config:  config,
		nowFunc: time.Now,
	}
}

// SetBanList sets the ban list for automatic banning.
func (ps *PeerScorer) SetBanList(bl *BanList) {
	ps.banList = bl
}

// SetStakeVerifier sets the stake verifier for tier-aware banning.
func (ps *PeerScorer) SetStakeVerifier(sv *StakeVerifier) {
	ps.stakeVerifier = sv
}

// RecordEvent records a behavioral event for a peer and adjusts their score.
func (ps *PeerScorer) RecordEvent(nodeID types.NodeID, event PeerEvent) {
	delta, ok := eventDeltas[event]
	if !ok {
		return
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	score, exists := ps.scores[nodeID]
	if !exists {
		score = &PeerScore{
			Score:      0.5, // Start neutral
			LastUpdate: ps.nowFunc(),
		}
		ps.scores[nodeID] = score
	}

	// Apply delta and clamp to [0.0, 1.0]
	score.Score += delta
	if score.Score > 1.0 {
		score.Score = 1.0
	}
	if score.Score < 0.0 {
		score.Score = 0.0
	}
	score.LastUpdate = ps.nowFunc()

	// Update counters
	switch event {
	case PeerEventInvalidMessage:
		score.InvalidMessages++
	case PeerEventFailedDeploy:
		score.FailedDeploys++
	case PeerEventGossipSpam:
		score.GossipSpam++
	case PeerEventRateLimitHit:
		score.RateLimitHits++
	}

	// Check thresholds
	ps.checkThresholds(nodeID, score)
}

// checkThresholds applies scoring actions based on score level.
// Must be called with ps.mu held.
func (ps *PeerScorer) checkThresholds(nodeID types.NodeID, score *PeerScore) {
	if score.Score < ps.config.BanThreshold {
		// BAN
		if ps.banList != nil {
			banDuration := 1 * time.Hour // default for staked
			// Unstaked peers get permanent ban
			if ps.stakeVerifier != nil && !ps.stakeVerifier.HasAnnounced(nodeID) {
				banDuration = 0 // permanent
			}
			ps.banList.Ban(nodeID, "peer score below ban threshold", banDuration)
		}
		logging.Warn("peer score below ban threshold",
			logging.NodeID(nodeID.String()[:16]),
			"score", score.Score,
			logging.Component("peer_scorer"))
	} else if score.Score < ps.config.ThrottleThreshold {
		logging.Warn("peer score below throttle threshold",
			logging.NodeID(nodeID.String()[:16]),
			"score", score.Score,
			logging.Component("peer_scorer"))
	} else if score.Score < ps.config.WarnThreshold {
		logging.Debug("peer score below warn threshold",
			logging.NodeID(nodeID.String()[:16]),
			"score", score.Score,
			logging.Component("peer_scorer"))
	}
}

// GetScore returns the current score for a peer, or 0.5 if unknown.
func (ps *PeerScorer) GetScore(nodeID types.NodeID) float64 {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[nodeID]; exists {
		return score.Score
	}
	return 0.5
}

// GetPeerScore returns the full score data for a peer, or nil if unknown.
func (ps *PeerScorer) GetPeerScore(nodeID types.NodeID) *PeerScore {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if score, exists := ps.scores[nodeID]; exists {
		// Return a copy
		copy := *score
		return &copy
	}
	return nil
}

// DecayScores decays all scores toward 0.5 (neutral).
// Call periodically (e.g. every DecayInterval).
func (ps *PeerScorer) DecayScores() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, score := range ps.scores {
		if score.Score > 0.5 {
			score.Score -= ps.config.DecayRate
			if score.Score < 0.5 {
				score.Score = 0.5
			}
		} else if score.Score < 0.5 {
			score.Score += ps.config.DecayRate
			if score.Score > 0.5 {
				score.Score = 0.5
			}
		}
	}
}

// RemovePeer removes a peer's score data.
func (ps *PeerScorer) RemovePeer(nodeID types.NodeID) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	delete(ps.scores, nodeID)
}

// PeerCount returns the number of scored peers.
func (ps *PeerScorer) PeerCount() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()
	return len(ps.scores)
}
