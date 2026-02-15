package types

// ProviderTier classifies a provider node's isolation capabilities.
type ProviderTier string

const (
	// ProviderTierConfidential is a Tier 1 provider with VM isolation and hardware TEE (SEV-SNP).
	ProviderTierConfidential ProviderTier = "confidential"

	// ProviderTierStandard is a Tier 2 provider with VM isolation (Kata) or runc (Linux).
	ProviderTierStandard ProviderTier = "standard"

	// ProviderTierDev is a development/macOS provider using runc via Colima.
	ProviderTierDev ProviderTier = "dev"
)

// TierRank returns a numeric rank for the provider tier.
// Higher rank = stronger isolation: confidential(3) > standard(2) > dev(1) > unknown(0).
func (t ProviderTier) TierRank() int {
	switch t {
	case ProviderTierConfidential:
		return 3
	case ProviderTierStandard:
		return 2
	case ProviderTierDev:
		return 1
	default:
		return 0
	}
}

// MeetsTierRequirement returns true if this tier is at least as strong as the required tier.
// If required is empty, any tier is accepted (backward compatible).
func (t ProviderTier) MeetsTierRequirement(required ProviderTier) bool {
	if required == "" {
		return true
	}
	return t.TierRank() >= required.TierRank()
}

// RuntimeCapabilities describes the runtime isolation features of a provider node.
type RuntimeCapabilities struct {
	// RuntimeName is the OCI runtime shim used (e.g. "io.containerd.runc.v2", "io.containerd.kata.v2").
	RuntimeName string `json:"runtime_name"`

	// ProviderTier is the classified tier based on hardware and runtime.
	ProviderTier ProviderTier `json:"provider_tier"`

	// KataAvailable indicates whether Kata Containers runtime is installed.
	KataAvailable bool `json:"kata_available"`

	// SEVSNPActive indicates active AMD SEV-SNP confidential computing.
	SEVSNPActive bool `json:"sev_snp_active"`

	// VMIsolation indicates whether containers run in dedicated VMs (Kata).
	VMIsolation bool `json:"vm_isolation"`
}
