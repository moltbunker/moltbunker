package types

// This file holds edge-provider (L7 edge layer, Approach A) types. Per the
// types zone rule, a new topic gets its own file — these constants are NOT
// appended to economics.go or types.go.

// StakingTierEdge is the edge-provider role tier. It is distinct from the
// compute-provider staking tiers (starter/bronze/silver/gold/platinum) because
// edge providers stake against a separate on-chain registry (BunkerEdgeRegistry)
// with its own minimum and slashing semantics, rather than the BunkerStaking
// tier enum.
const StakingTierEdge StakingTier = "edge"

// IsEdgeTier reports whether the given staking tier is the edge-provider role.
func IsEdgeTier(t StakingTier) bool {
	return t == StakingTierEdge
}

// EdgeProviderCapability is a bitmask describing the L7 edge features a provider
// advertises. EDGE-02's tunnel gate logic references these bits without locking
// the on-chain registry to a specific bitmask encoding (the contract stores
// opaque metadata; the daemon interprets capabilities off-chain).
type EdgeProviderCapability uint64

const (
	// EdgeCapWAF indicates the edge node runs an application-layer WAF.
	EdgeCapWAF EdgeProviderCapability = 1 << 0

	// EdgeCapDDoSMitigation indicates the edge node performs L3/L4/L7 DDoS mitigation.
	EdgeCapDDoSMitigation EdgeProviderCapability = 1 << 1

	// EdgeCapTLSTermination indicates the edge node terminates TLS for tenants.
	EdgeCapTLSTermination EdgeProviderCapability = 1 << 2

	// EdgeCapACME indicates the edge node issues/renews certificates via ACME.
	EdgeCapACME EdgeProviderCapability = 1 << 3
)

// Has reports whether the capability set includes the given capability bit.
func (c EdgeProviderCapability) Has(cap EdgeProviderCapability) bool {
	return c&cap == cap
}
