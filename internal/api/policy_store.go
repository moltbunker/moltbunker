package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// Policy validation limits
const (
	maxAllowedRegions    = 100
	maxBlacklistedImages = 500
	maxAllowedTiers      = 10
)

// NetworkPolicies stores admin-configurable network policies
type NetworkPolicies struct {
	NodePolicies NodePolicies `json:"node_policies"`
	NetworkRules NetworkRules `json:"network_rules"`

	Slashing   *types.SlashingConfig   `json:"slashing,omitempty"`
	Pricing    *types.PricingConfig    `json:"pricing,omitempty"`
	Fees       *types.ProtocolFees     `json:"fees,omitempty"`
	Reputation *types.ReputationConfig `json:"reputation,omitempty"`

	UpdatedAt time.Time `json:"updated_at"`
	UpdatedBy string    `json:"updated_by"`
}

// Validate checks policy values for sanity
func (p *NetworkPolicies) Validate() error {
	if p.NetworkRules.MaxContainersPerNode < 0 {
		return fmt.Errorf("max_containers_per_node cannot be negative")
	}
	if p.NetworkRules.MaxDeploymentHours < 0 {
		return fmt.Errorf("max_deployment_hours cannot be negative")
	}
	if len(p.NetworkRules.AllowedRegions) > maxAllowedRegions {
		return fmt.Errorf("too many allowed regions (max %d)", maxAllowedRegions)
	}
	if len(p.NetworkRules.BlacklistedImages) > maxBlacklistedImages {
		return fmt.Errorf("too many blacklisted images (max %d)", maxBlacklistedImages)
	}
	if len(p.NodePolicies.AllowedTiers) > maxAllowedTiers {
		return fmt.Errorf("too many allowed tiers (max %d)", maxAllowedTiers)
	}
	if p.NodePolicies.MinReputation < 0 {
		return fmt.Errorf("min_reputation cannot be negative")
	}

	if p.Fees != nil {
		if p.Fees.TotalFeePercent < 0 || p.Fees.TotalFeePercent > 100 {
			return fmt.Errorf("total_fee_percent must be 0-100")
		}
		if p.Fees.BurnPercent < 0 || p.Fees.BurnPercent > 100 {
			return fmt.Errorf("burn_percent must be 0-100")
		}
		if p.Fees.TreasuryPercent < 0 || p.Fees.TreasuryPercent > 100 {
			return fmt.Errorf("treasury_percent must be 0-100")
		}
	}

	if p.Slashing != nil {
		if p.Slashing.DowntimePerHourPercent < 0 || p.Slashing.DowntimeMaxPercent < 0 {
			return fmt.Errorf("slashing percentages cannot be negative")
		}
		if p.Slashing.MaliciousActivityPercent > 100 {
			return fmt.Errorf("slashing percentage cannot exceed 100")
		}
	}

	return nil
}

// NodePolicies defines requirements for provider nodes
type NodePolicies struct {
	MinStakingAmount string   `json:"min_staking_amount,omitempty"`
	MinReputation    int      `json:"min_reputation,omitempty"`
	RequireSEVSNP    bool     `json:"require_sev_snp"`
	RequireECC       bool     `json:"require_ecc"`
	RequireNVMe      bool     `json:"require_nvme"`
	RequireTPM       bool     `json:"require_tpm"`
	AllowedTiers     []string `json:"allowed_tiers,omitempty"`
}

// NetworkRules defines network-wide operational rules
type NetworkRules struct {
	MaxContainersPerNode int      `json:"max_containers_per_node,omitempty"`
	AllowedRegions       []string `json:"allowed_regions,omitempty"`
	BlacklistedImages    []string `json:"blacklisted_images,omitempty"`
	MaxDeploymentHours   int      `json:"max_deployment_hours,omitempty"`
	RequireEncryption    bool     `json:"require_encryption"`
}

// PolicyStore persists admin-configurable policies to a JSON file
type PolicyStore struct {
	mu       sync.RWMutex
	policies *NetworkPolicies
	filePath string
}

// NewPolicyStore creates a store, loading existing policies from disk
func NewPolicyStore(filePath string) *PolicyStore {
	// Ensure parent directory exists with secure permissions
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		logging.Warn("failed to create admin policies directory",
			"dir", dir,
			"error", err.Error(),
			logging.Component("api"))
	}

	s := &PolicyStore{
		policies: &NetworkPolicies{},
		filePath: filePath,
	}
	if err := s.load(); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to load admin policies",
			"error", err.Error(),
			logging.Component("api"))
	}
	return s
}

// load reads policies from disk
func (s *PolicyStore) load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}

	var policies NetworkPolicies
	if err := json.Unmarshal(data, &policies); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies = &policies
	return nil
}

// saveLocked writes policies to disk.
// MUST be called while s.mu is held.
func (s *PolicyStore) saveLocked() error {
	data, err := json.MarshalIndent(s.policies, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0600)
}

// Get returns the current policies (deep copy of top-level struct)
func (s *PolicyStore) Get() *NetworkPolicies {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Marshal/unmarshal for deep copy to prevent shared slice mutation
	data, err := json.Marshal(s.policies)
	if err != nil {
		cp := *s.policies
		return &cp
	}
	var cp NetworkPolicies
	if err := json.Unmarshal(data, &cp); err != nil {
		shallow := *s.policies
		return &shallow
	}
	return &cp
}

// Update replaces policies and persists to disk
func (s *PolicyStore) Update(updated *NetworkPolicies) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	updated.UpdatedAt = time.Now()
	s.policies = updated
	return s.saveLocked()
}
