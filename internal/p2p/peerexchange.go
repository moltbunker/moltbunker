package p2p

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/util"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// DefaultPeerExchangeInterval is the default interval between peer exchange rounds.
	DefaultPeerExchangeInterval = 5 * time.Minute

	// DefaultMaxSharedPeers is the maximum number of peers shared per exchange.
	DefaultMaxSharedPeers = 20

	// MessageTypePeerExchangeRequest is the message type for requesting peers.
	MessageTypePeerExchangeRequest types.MessageType = "peer_exchange_request"

	// MessageTypePeerExchangeResponse is the message type for responding with peers.
	MessageTypePeerExchangeResponse types.MessageType = "peer_exchange_response"
)

// PeerExchangeConfig holds configuration for the peer exchange protocol.
type PeerExchangeConfig struct {
	// Interval is the duration between periodic peer exchange rounds.
	// Defaults to DefaultPeerExchangeInterval if zero.
	Interval time.Duration

	// MaxSharedPeers is the maximum number of peers to share per exchange.
	// Defaults to DefaultMaxSharedPeers if zero.
	MaxSharedPeers int
}

// PeerExchangeAddr is a serializable representation of a peer address
// used in peer exchange messages.
type PeerExchangeAddr struct {
	NodeID  string   `json:"node_id"`
	Addrs   []string `json:"addrs"`
	Region  string   `json:"region,omitempty"`
	Country string   `json:"country,omitempty"`
}

// PeerExchangeMessage is the payload for peer exchange request/response messages.
type PeerExchangeMessage struct {
	Peers     []PeerExchangeAddr `json:"peers,omitempty"`
	RequestID string             `json:"request_id"`
}

// PeerExchangeProtocol implements periodic peer exchange with connected peers.
// It shares high-quality peers from the address book and integrates discovered
// peers back into the local address book and routing table.
type PeerExchangeProtocol struct {
	router      *Router
	addressBook *AddressBook
	config      *PeerExchangeConfig

	// pendingRequests tracks outstanding peer exchange requests.
	pendingRequests   map[string]chan []PeerExchangeAddr
	pendingRequestsMu sync.Mutex

	// nowFunc allows injecting time for testing.
	nowFunc func() time.Time
}

// NewPeerExchangeProtocol creates a new peer exchange protocol instance.
func NewPeerExchangeProtocol(router *Router, addressBook *AddressBook, config *PeerExchangeConfig) *PeerExchangeProtocol {
	if config == nil {
		config = &PeerExchangeConfig{}
	}
	if config.Interval <= 0 {
		config.Interval = DefaultPeerExchangeInterval
	}
	if config.MaxSharedPeers <= 0 {
		config.MaxSharedPeers = DefaultMaxSharedPeers
	}

	pex := &PeerExchangeProtocol{
		router:          router,
		addressBook:     addressBook,
		config:          config,
		pendingRequests: make(map[string]chan []PeerExchangeAddr),
		nowFunc:         time.Now,
	}

	// Register message handlers on the router
	if router != nil {
		router.RegisterHandler(MessageTypePeerExchangeRequest, pex.handleRequest)
		router.RegisterHandler(MessageTypePeerExchangeResponse, pex.handleResponse)
	}

	return pex
}

// Start starts the periodic peer exchange loop. It blocks until the
// context is canceled.
func (pex *PeerExchangeProtocol) Start(ctx context.Context) {
	ticker := time.NewTicker(pex.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pex.exchangeRound(ctx)
		}
	}
}

// StartBackground starts the peer exchange loop in a background goroutine.
func (pex *PeerExchangeProtocol) StartBackground(ctx context.Context) {
	util.SafeGoWithName("peer-exchange", func() {
		pex.Start(ctx)
	})
}

// exchangeRound performs one round of peer exchange with all connected peers.
func (pex *PeerExchangeProtocol) exchangeRound(ctx context.Context) {
	peers := pex.router.GetPeers()
	if len(peers) == 0 {
		return
	}

	logging.Debug("starting peer exchange round",
		"connected_peers", len(peers),
		logging.Component("peerexchange"))

	for _, peer := range peers {
		select {
		case <-ctx.Done():
			return
		default:
		}

		exchanged, err := pex.ExchangePeers(ctx, peer.ID)
		if err != nil {
			logging.Debug("peer exchange failed",
				logging.NodeID(peer.ID.String()[:16]),
				"error", err,
				logging.Component("peerexchange"))
			continue
		}

		pex.integrateDiscoveredPeers(exchanged)
	}
}

// ExchangePeers sends a peer exchange request to a specific peer and waits
// for the response. Returns the peers shared by the remote peer.
func (pex *PeerExchangeProtocol) ExchangePeers(ctx context.Context, peerID types.NodeID) ([]PeerExchangeAddr, error) {
	requestID := fmt.Sprintf("pex-%d-%s", pex.nowFunc().UnixNano(), peerID.String()[:16])

	// Create response channel
	respCh := make(chan []PeerExchangeAddr, 1)
	pex.pendingRequestsMu.Lock()
	pex.pendingRequests[requestID] = respCh
	pex.pendingRequestsMu.Unlock()

	defer func() {
		pex.pendingRequestsMu.Lock()
		delete(pex.pendingRequests, requestID)
		pex.pendingRequestsMu.Unlock()
	}()

	// Build request
	reqMsg := PeerExchangeMessage{
		RequestID: requestID,
	}

	payload, err := json.Marshal(reqMsg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal peer exchange request: %w", err)
	}

	msg := &types.Message{
		Type:      MessageTypePeerExchangeRequest,
		Payload:   payload,
		Timestamp: pex.nowFunc(),
	}

	if err := pex.router.SendMessage(ctx, peerID, msg); err != nil {
		return nil, fmt.Errorf("failed to send peer exchange request: %w", err)
	}

	// Wait for response with timeout
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case peers := <-respCh:
		return peers, nil
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("peer exchange request timed out")
	}
}

// HandlePeerExchange processes an incoming peer exchange request and returns
// the top peers by quality score. This is the synchronous handler used
// for direct invocation (e.g., in tests).
func (pex *PeerExchangeProtocol) HandlePeerExchange(msg *PeerExchangeMessage) []PeerExchangeAddr {
	return pex.getTopPeers()
}

// handleRequest is the message handler for incoming peer exchange requests.
func (pex *PeerExchangeProtocol) handleRequest(ctx context.Context, msg *types.Message, from *types.Node) error {
	var reqMsg PeerExchangeMessage
	if err := json.Unmarshal(msg.Payload, &reqMsg); err != nil {
		return fmt.Errorf("failed to unmarshal peer exchange request: %w", err)
	}

	// Get our best peers to share
	topPeers := pex.getTopPeers()

	// Build response
	respMsg := PeerExchangeMessage{
		Peers:     topPeers,
		RequestID: reqMsg.RequestID,
	}

	payload, err := json.Marshal(respMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal peer exchange response: %w", err)
	}

	response := &types.Message{
		Type:      MessageTypePeerExchangeResponse,
		Payload:   payload,
		Timestamp: pex.nowFunc(),
	}

	if from != nil {
		if err := pex.router.SendMessage(ctx, from.ID, response); err != nil {
			return fmt.Errorf("failed to send peer exchange response: %w", err)
		}
	}

	return nil
}

// handleResponse is the message handler for incoming peer exchange responses.
func (pex *PeerExchangeProtocol) handleResponse(_ context.Context, msg *types.Message, _ *types.Node) error {
	var respMsg PeerExchangeMessage
	if err := json.Unmarshal(msg.Payload, &respMsg); err != nil {
		return fmt.Errorf("failed to unmarshal peer exchange response: %w", err)
	}

	pex.pendingRequestsMu.Lock()
	ch, exists := pex.pendingRequests[respMsg.RequestID]
	pex.pendingRequestsMu.Unlock()

	if exists {
		select {
		case ch <- respMsg.Peers:
		default:
			// Channel already has a value or is closed
		}
	}

	return nil
}

// getTopPeers returns the top peers by quality score from the address book.
func (pex *PeerExchangeProtocol) getTopPeers() []PeerExchangeAddr {
	if pex.addressBook == nil {
		return nil
	}

	entries := pex.addressBook.GetBestPeers(pex.config.MaxSharedPeers)

	result := make([]PeerExchangeAddr, 0, len(entries))
	for _, entry := range entries {
		if len(entry.Addrs) == 0 {
			continue
		}

		peerAddr := PeerExchangeAddr{
			NodeID: entry.PeerID.String(),
			Addrs:  entry.Addrs,
		}
		result = append(result, peerAddr)
	}

	return result
}

// integrateDiscoveredPeers adds newly discovered peers to the router's
// peer tracking. This does not immediately connect; peers are added
// for future connection attempts.
func (pex *PeerExchangeProtocol) integrateDiscoveredPeers(peers []PeerExchangeAddr) {
	if len(peers) == 0 {
		return
	}

	added := 0
	for _, pa := range peers {
		if len(pa.Addrs) == 0 {
			continue
		}

		// Decode the hex NodeID directly instead of hashing the string.
		// Invalid or malformed NodeIDs are skipped to prevent routing
		// table pollution from malicious peer exchange messages.
		nodeIDBytes, err := hex.DecodeString(pa.NodeID)
		if err != nil || len(nodeIDBytes) != 32 {
			continue
		}
		var nodeID types.NodeID
		copy(nodeID[:], nodeIDBytes)

		node := &types.Node{
			ID:       nodeID,
			Address:  pa.Addrs[0],
			Region:   pa.Region,
			Country:  pa.Country,
			LastSeen: pex.nowFunc(),
		}

		if pex.router.AddPeer(node) {
			added++
		}
	}

	if added > 0 {
		logging.Debug("integrated peers from exchange",
			"added", added,
			"received", len(peers),
			logging.Component("peerexchange"))
	}
}

// SortPeersByQuality sorts peers by a quality metric derived from the
// address book success rate and recency. This is exported for use by
// other components that need peer ranking.
func SortPeersByQuality(entries []*AddressEntry) []*AddressEntry {
	sorted := make([]*AddressEntry, len(entries))
	copy(sorted, entries)

	sort.Slice(sorted, func(i, j int) bool {
		scoreI := peerQualityScore(sorted[i])
		scoreJ := peerQualityScore(sorted[j])
		return scoreI > scoreJ
	})

	return sorted
}

// peerQualityScore computes a quality score for a peer entry.
// Higher is better. Combines success rate (0-1) with recency bonus.
func peerQualityScore(entry *AddressEntry) float64 {
	// Base score from success rate (0.0 - 1.0)
	score := entry.successRate()

	// Recency bonus: peers seen in the last hour get a boost
	age := time.Since(entry.LastSeen)
	switch {
	case age < 1*time.Hour:
		score += 0.3
	case age < 6*time.Hour:
		score += 0.2
	case age < 24*time.Hour:
		score += 0.1
	}

	// Bonus for peers with successful connections
	if entry.ConnSuccess > 0 {
		score += 0.1
	}

	return score
}
