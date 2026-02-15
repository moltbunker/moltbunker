package p2p

import (
	"context"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/payment"
	"github.com/moltbunker/moltbunker/pkg/types"
)

const (
	// DefaultStakeCacheTTL is how long positive stake info is cached.
	DefaultStakeCacheTTL = 5 * time.Minute

	// DefaultNegativeStakeCacheTTL is how long "no stake" results are cached.
	DefaultNegativeStakeCacheTTL = 2 * time.Minute

	// DefaultErrorStakeCacheTTL is how long transient errors are cached.
	DefaultErrorStakeCacheTTL = 30 * time.Second
)

// CachedStakeInfo holds cached on-chain stake lookup results.
type CachedStakeInfo struct {
	HasMinStake bool
	Tier        types.StakingTier
	Amount      *big.Int
	FetchedAt   time.Time
	IsError     bool // true if this entry was cached due to a lookup error
}

// StakeVerifier maps NodeID → wallet → on-chain stake with caching.
// It decouples the announce handshake from the actual on-chain verification,
// allowing peers to connect and be verified asynchronously.
type StakeVerifier struct {
	mu             sync.RWMutex
	peerWallets    map[types.NodeID]common.Address
	stakeCache     map[common.Address]*CachedStakeInfo
	paymentService *payment.PaymentService
	cacheTTL       time.Duration
	negativeTTL    time.Duration
	errorTTL       time.Duration
	nowFunc        func() time.Time
}

// NewStakeVerifier creates a new stake verifier.
// paymentService may be nil initially and set later via SetPaymentService.
func NewStakeVerifier(ps *payment.PaymentService) *StakeVerifier {
	return &StakeVerifier{
		peerWallets:    make(map[types.NodeID]common.Address),
		stakeCache:     make(map[common.Address]*CachedStakeInfo),
		paymentService: ps,
		cacheTTL:       DefaultStakeCacheTTL,
		negativeTTL:    DefaultNegativeStakeCacheTTL,
		errorTTL:       DefaultErrorStakeCacheTTL,
		nowFunc:        time.Now,
	}
}

// NewStakeVerifierWithConfig creates a stake verifier with custom TTLs.
func NewStakeVerifierWithConfig(ps *payment.PaymentService, cacheTTL, negativeTTL, errorTTL time.Duration) *StakeVerifier {
	sv := NewStakeVerifier(ps)
	sv.cacheTTL = cacheTTL
	sv.negativeTTL = negativeTTL
	sv.errorTTL = errorTTL
	return sv
}

// SetPaymentService sets or updates the payment service.
// This solves initialization ordering: StakeVerifier can be created before
// PaymentService is ready and wired up later.
func (sv *StakeVerifier) SetPaymentService(ps *payment.PaymentService) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.paymentService = ps
}

// RegisterPeerWallet records the NodeID → wallet mapping after a verified announce.
func (sv *StakeVerifier) RegisterPeerWallet(nodeID types.NodeID, wallet common.Address) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	sv.peerWallets[nodeID] = wallet

	logging.Debug("registered peer wallet",
		logging.NodeID(nodeID.String()[:16]),
		"wallet", wallet.Hex()[:10],
		logging.Component("stake_verifier"))
}

// HasAnnounced returns true if the peer has completed the announce handshake.
func (sv *StakeVerifier) HasAnnounced(nodeID types.NodeID) bool {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	_, exists := sv.peerWallets[nodeID]
	return exists
}

// GetWallet returns the wallet address for a peer, if known.
func (sv *StakeVerifier) GetWallet(nodeID types.NodeID) (common.Address, bool) {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	wallet, exists := sv.peerWallets[nodeID]
	return wallet, exists
}

// IsStaked returns true if the peer has minimum stake on-chain.
// If PaymentService is nil or the peer hasn't announced, returns false.
func (sv *StakeVerifier) IsStaked(ctx context.Context, nodeID types.NodeID) bool {
	sv.mu.RLock()
	wallet, exists := sv.peerWallets[nodeID]
	ps := sv.paymentService
	sv.mu.RUnlock()

	if !exists {
		return false
	}

	if ps == nil {
		// PaymentService not yet available — can't verify
		return false
	}

	info := sv.getCachedOrFetch(ctx, wallet, ps)
	return info != nil && info.HasMinStake
}

// GetTier returns the staking tier for a peer.
// Returns "unstaked" if the peer hasn't announced, "unknown" if PaymentService
// is unavailable, or the actual tier from on-chain data.
func (sv *StakeVerifier) GetTier(ctx context.Context, nodeID types.NodeID) string {
	sv.mu.RLock()
	wallet, exists := sv.peerWallets[nodeID]
	ps := sv.paymentService
	sv.mu.RUnlock()

	if !exists {
		return "unstaked"
	}

	if ps == nil {
		return "unknown"
	}

	info := sv.getCachedOrFetch(ctx, wallet, ps)
	if info == nil || info.IsError {
		return "unknown"
	}
	if !info.HasMinStake {
		return "unstaked"
	}
	return string(info.Tier)
}

// RemovePeer removes a peer's wallet mapping and any cached stake info.
func (sv *StakeVerifier) RemovePeer(nodeID types.NodeID) {
	sv.mu.Lock()
	defer sv.mu.Unlock()

	wallet, exists := sv.peerWallets[nodeID]
	if exists {
		delete(sv.peerWallets, nodeID)
		// Only remove cache if no other peer uses this wallet
		walletInUse := false
		for _, w := range sv.peerWallets {
			if w == wallet {
				walletInUse = true
				break
			}
		}
		if !walletInUse {
			delete(sv.stakeCache, wallet)
		}
	}
}

// PeerCount returns the number of announced peers.
func (sv *StakeVerifier) PeerCount() int {
	sv.mu.RLock()
	defer sv.mu.RUnlock()
	return len(sv.peerWallets)
}

// getCachedOrFetch returns cached stake info or fetches from chain.
func (sv *StakeVerifier) getCachedOrFetch(ctx context.Context, wallet common.Address, ps *payment.PaymentService) *CachedStakeInfo {
	now := sv.nowFunc()

	// Check cache first
	sv.mu.RLock()
	cached, exists := sv.stakeCache[wallet]
	sv.mu.RUnlock()

	if exists {
		ttl := sv.cacheTTL
		if !cached.HasMinStake {
			ttl = sv.negativeTTL
		}
		if cached.IsError {
			ttl = sv.errorTTL
		}
		if now.Sub(cached.FetchedAt) < ttl {
			return cached
		}
	}

	// Fetch from chain
	info := sv.fetchStakeInfo(ctx, wallet, ps)

	// Cache the result
	sv.mu.Lock()
	sv.stakeCache[wallet] = info
	sv.mu.Unlock()

	return info
}

// fetchStakeInfo queries on-chain stake for a wallet.
func (sv *StakeVerifier) fetchStakeInfo(ctx context.Context, wallet common.Address, ps *payment.PaymentService) *CachedStakeInfo {
	now := sv.nowFunc()

	hasMinStake, err := ps.HasMinimumStake(ctx, wallet)
	if err != nil {
		logging.Warn("failed to check stake",
			"wallet", wallet.Hex()[:10],
			"error", err.Error(),
			logging.Component("stake_verifier"))
		return &CachedStakeInfo{
			FetchedAt: now,
			IsError:   true,
		}
	}

	tier, err := ps.GetTier(ctx, wallet)
	if err != nil {
		logging.Warn("failed to get tier",
			"wallet", wallet.Hex()[:10],
			"error", err.Error(),
			logging.Component("stake_verifier"))
		return &CachedStakeInfo{
			HasMinStake: hasMinStake,
			FetchedAt:   now,
			IsError:     true,
		}
	}

	var amount *big.Int
	stakeInfo, err := ps.GetStakeInfo(ctx, wallet)
	if err == nil && stakeInfo != nil {
		amount = stakeInfo.StakedAmount
	}

	return &CachedStakeInfo{
		HasMinStake: hasMinStake,
		Tier:        tier,
		Amount:      amount,
		FetchedAt:   now,
	}
}

// InvalidateWallet removes the cached stake info for a wallet,
// forcing a fresh on-chain fetch on the next access.
func (sv *StakeVerifier) InvalidateWallet(wallet common.Address) {
	sv.mu.Lock()
	defer sv.mu.Unlock()
	delete(sv.stakeCache, wallet)

	logging.Debug("invalidated stake cache for wallet",
		"wallet", wallet.Hex()[:10],
		logging.Component("stake_verifier"))
}

// VerifyNodeIDOnChain checks if the given NodeID is registered on-chain to the given wallet.
// Returns true if the on-chain mapping matches, or true (optimistic) if the chain is unavailable.
func (sv *StakeVerifier) VerifyNodeIDOnChain(ctx context.Context, nodeID types.NodeID, wallet common.Address) bool {
	sv.mu.RLock()
	ps := sv.paymentService
	sv.mu.RUnlock()

	if ps == nil {
		// Chain unavailable — optimistic pass
		return true
	}

	// Convert types.NodeID (hex string of SHA256) to [32]byte
	var nodeIDBytes [32]byte
	nodeIDStr := nodeID.String()
	if len(nodeIDStr) >= 64 {
		for i := 0; i < 32; i++ {
			b, err := hexByteParse(nodeIDStr[i*2], nodeIDStr[i*2+1])
			if err != nil {
				return true // parse error — optimistic pass
			}
			nodeIDBytes[i] = b
		}
	}

	registeredAddr, err := ps.NodeIDToProvider(ctx, nodeIDBytes)
	if err != nil {
		logging.Warn("failed to verify NodeID on-chain, allowing optimistically",
			logging.NodeID(nodeID.String()[:16]),
			"error", err.Error(),
			logging.Component("stake_verifier"))
		return true
	}

	// Zero address means not registered — that's OK for new nodes
	if registeredAddr == (common.Address{}) {
		return true
	}

	return registeredAddr == wallet
}

// StartEventListener launches a goroutine that listens for stake and slash events
// and invalidates the corresponding wallet's cache entry.
func (sv *StakeVerifier) StartEventListener(ctx context.Context, stakeEvents <-chan *payment.StakeEvent, slashEvents <-chan *payment.SlashEvent) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-stakeEvents:
				if !ok {
					return
				}
				sv.InvalidateWallet(ev.Provider)
				logging.Debug("invalidated cache on stake event",
					"provider", ev.Provider.Hex()[:10],
					logging.Component("stake_verifier"))
			case ev, ok := <-slashEvents:
				if !ok {
					return
				}
				sv.InvalidateWallet(ev.Provider)
				logging.Debug("invalidated cache on slash event",
					"provider", ev.Provider.Hex()[:10],
					logging.Component("stake_verifier"))
			}
		}
	}()
}

// hexByteParse parses two hex characters into a byte.
func hexByteParse(hi, lo byte) (byte, error) {
	h, ok1 := hexVal(hi)
	l, ok2 := hexVal(lo)
	if !ok1 || !ok2 {
		return 0, fmt.Errorf("invalid hex byte: %c%c", hi, lo)
	}
	return (h << 4) | l, nil
}

func hexVal(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

// NewMockStakeVerifier creates a stake verifier pre-populated with test data.
// Useful for unit tests that don't want to wire up a PaymentService.
func NewMockStakeVerifier(peers map[types.NodeID]common.Address, stakes map[common.Address]*CachedStakeInfo) *StakeVerifier {
	sv := NewStakeVerifier(nil)
	for nodeID, wallet := range peers {
		sv.peerWallets[nodeID] = wallet
	}
	for wallet, info := range stakes {
		sv.stakeCache[wallet] = info
	}
	return sv
}
