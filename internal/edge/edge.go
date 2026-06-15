// Package edge implements the daemon-side edge-provider role: the stake-gated
// "edge" node tier (Approach A from the 2026-05-18 decision) that terminates
// TLS + runs an L7 WAF + handles DDoS for container hosts that tunnel to it.
//
// This package is interface-first and deliberately decoupled from the on-chain
// edge-stake contract (BunkerEdgeRegistry, SC-EDGE-01). The daemon can gate the
// edge role with either:
//
//   - ConfigEdgeTierChecker — a static allowlist of NodeIDs from config. This is
//     the default and lets the edge role work WITHOUT any contract deployed.
//   - OnChainEdgeTierChecker — backed by the EdgeRegistryReader seam from
//     SC-EDGE-01 (payment.EdgeRegistryReader), which checks the wallet's
//     on-chain edge-provider registration.
//
// A third implementation can be dropped in later (e.g. a dedicated edge-staking
// contract) without touching any call site, because everything consumes the
// EdgeTierChecker interface.
package edge

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// EdgeTierChecker decides whether a node is authorized to act as an edge
// provider. It is the single gating seam consumed by the reverse tunnel server
// (internal/tunnel) so the gate can be swapped between a config allowlist and
// an on-chain check without changing the call site.
type EdgeTierChecker interface {
	// IsEdgeAuthorized reports whether the node identified by nodeID / wallet is
	// permitted to register as an edge provider. A non-nil error means the check
	// could not be completed (e.g. RPC failure) and callers MUST treat that as
	// "not authorized" (fail-closed) for the edge role.
	IsEdgeAuthorized(ctx context.Context, nodeID types.NodeID, wallet common.Address) (bool, error)
}

// Mode selects which EdgeTierChecker implementation NewEdgeTierChecker builds.
type Mode string

const (
	// ModeConfig gates the edge role on a static allowlist of NodeIDs. Default.
	ModeConfig Mode = "config"
	// ModeOnChain gates the edge role on the on-chain BunkerEdgeRegistry via the
	// payment.EdgeRegistryReader seam.
	ModeOnChain Mode = "onchain"
)

// Config is the daemon-side edge-role configuration, mapped from
// config.EdgeRoleConfig. Kept in this package (rather than importing
// internal/config) so internal/config does not depend on internal/edge.
type Config struct {
	// Mode is "config" (allowlist) or "onchain" (contract). Empty defaults to
	// "config".
	Mode Mode
	// AllowedNodeIDs is the allowlist consumed by ConfigEdgeTierChecker. Each
	// entry is a NodeID hex string (the SHA256 of the node's SPKI, as returned
	// by types.NodeID.String()).
	AllowedNodeIDs []string
	// MinTier is the minimum staking tier accepted by tier-based checks. It is
	// informational for the config checker and reserved for future tier-aware
	// on-chain policy; the current OnChainEdgeTierChecker uses the registry's
	// active/frozen flags rather than a tier comparison.
	MinTier string
}

// EdgeRegistryReader is the read-only on-chain seam this package consumes for
// the on-chain gate. It is satisfied by payment.EdgeRegistryReader (the
// production contract reader and its mock both implement it). Re-declaring the
// method set here keeps internal/edge from importing internal/payment, avoiding
// an import cycle and keeping the dependency direction one-way.
type EdgeRegistryReader interface {
	IsActiveEdgeProvider(ctx context.Context, addr common.Address) (bool, error)
}

// ConfigEdgeTierChecker authorizes the edge role from a static NodeID
// allowlist. It is the default gate and requires no contract. Safe for
// concurrent use (the allowlist is immutable after construction).
type ConfigEdgeTierChecker struct {
	allowed map[string]struct{}
}

// NewConfigEdgeTierChecker builds an allowlist gate from the given NodeID hex
// strings. An empty allowlist authorizes nothing.
func NewConfigEdgeTierChecker(allowedNodeIDs []string) *ConfigEdgeTierChecker {
	allowed := make(map[string]struct{}, len(allowedNodeIDs))
	for _, id := range allowedNodeIDs {
		if id == "" {
			continue
		}
		allowed[id] = struct{}{}
	}
	return &ConfigEdgeTierChecker{allowed: allowed}
}

// IsEdgeAuthorized implements EdgeTierChecker against the static allowlist.
func (c *ConfigEdgeTierChecker) IsEdgeAuthorized(_ context.Context, nodeID types.NodeID, _ common.Address) (bool, error) {
	_, ok := c.allowed[nodeID.String()]
	return ok, nil
}

// OnChainEdgeTierChecker authorizes the edge role from the on-chain
// BunkerEdgeRegistry via the EdgeRegistryReader seam. It works on Base Sepolia
// today against a deployed registry and falls back cleanly (returns the
// registry error, which the caller treats as not-authorized) when the RPC is
// unavailable.
type OnChainEdgeTierChecker struct {
	reader EdgeRegistryReader
}

// NewOnChainEdgeTierChecker builds an on-chain gate backed by the given
// registry reader (payment.EdgeRegistryReader satisfies it).
func NewOnChainEdgeTierChecker(reader EdgeRegistryReader) *OnChainEdgeTierChecker {
	return &OnChainEdgeTierChecker{reader: reader}
}

// IsEdgeAuthorized implements EdgeTierChecker against the on-chain registry.
// A nil reader fails closed (returns false, nil) so a mis-wired daemon never
// silently authorizes every node.
func (c *OnChainEdgeTierChecker) IsEdgeAuthorized(ctx context.Context, _ types.NodeID, wallet common.Address) (bool, error) {
	if c.reader == nil {
		return false, nil
	}
	return c.reader.IsActiveEdgeProvider(ctx, wallet)
}

// NewEdgeTierChecker selects the right EdgeTierChecker for the config Mode.
//
//   - ModeOnChain requires a non-nil reader; if reader is nil it falls back to
//     the config allowlist (fail-safe: never returns nil).
//   - Any other Mode (including the empty default) returns the config allowlist
//     checker.
func NewEdgeTierChecker(cfg Config, reader EdgeRegistryReader) EdgeTierChecker {
	if cfg.Mode == ModeOnChain && reader != nil {
		return NewOnChainEdgeTierChecker(reader)
	}
	return NewConfigEdgeTierChecker(cfg.AllowedNodeIDs)
}

// MockEdgeTierChecker is a fixed-answer EdgeTierChecker for tests.
type MockEdgeTierChecker struct {
	Authorized bool
	Err        error
}

// IsEdgeAuthorized returns the configured answer.
func (m *MockEdgeTierChecker) IsEdgeAuthorized(_ context.Context, _ types.NodeID, _ common.Address) (bool, error) {
	return m.Authorized, m.Err
}

// Compile-time assertions that all implementations satisfy the interface.
var (
	_ EdgeTierChecker = (*ConfigEdgeTierChecker)(nil)
	_ EdgeTierChecker = (*OnChainEdgeTierChecker)(nil)
	_ EdgeTierChecker = (*MockEdgeTierChecker)(nil)
)
