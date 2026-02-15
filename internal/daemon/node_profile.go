package daemon

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/moltbunker/moltbunker/internal/config"
	"github.com/moltbunker/moltbunker/internal/logging"
	"github.com/moltbunker/moltbunker/internal/runtime"
	"github.com/moltbunker/moltbunker/pkg/types"
)

// NodeProfile represents a known node in the network
type NodeProfile struct {
	NodeID        string              `json:"node_id"`
	Address       string              `json:"address,omitempty"`
	WalletAddress string              `json:"wallet_address,omitempty"`
	Region        string              `json:"region"`
	Country       string              `json:"country,omitempty"`
	Location      *types.NodeLocation `json:"location,omitempty"`
	Online        bool                `json:"online"`
	LastSeen      time.Time           `json:"last_seen"`

	// Capacity — self: from config DeclaredCPU/Memory/Storage; peers: from gossip (future, 0 for now)
	Capacity CapacityProfile `json:"capacity"`

	// Economics
	Tier            string `json:"tier"`
	Role            string `json:"role"`
	ReputationScore int    `json:"reputation_score"`
	StakingAmount   uint64 `json:"staking_amount"`

	// Runtime
	ActiveContainers int                       `json:"active_containers"`
	EncryptedCount   int                       `json:"encrypted_containers"`
	Version          string                    `json:"version,omitempty"`
	ProviderTier     types.ProviderTier         `json:"provider_tier,omitempty"`
	RuntimeName      string                    `json:"runtime_name,omitempty"`
	KataSupported    bool                      `json:"kata_supported,omitempty"`
	SEVSNPActive     bool                      `json:"sev_snp_active,omitempty"`

	// Admin-assigned metadata (merged from admin store)
	Badges  []string `json:"badges,omitempty"`
	Blocked bool     `json:"blocked,omitempty"`
}

// CapacityProfile describes a node's declared resource capacity
type CapacityProfile struct {
	CPUCores      int              `json:"cpu_cores"`
	MemoryGB      int              `json:"memory_gb"`
	StorageGB     int              `json:"storage_gb"`
	BandwidthMbps int              `json:"bandwidth_mbps"`
	GPUCount      int              `json:"gpu_count,omitempty"`
	GPUModel      string           `json:"gpu_model,omitempty"`
	Hardware      *HardwareProfile `json:"hardware,omitempty"`
}

// AggregatedCapacity sums capacity across all known online nodes
type AggregatedCapacity struct {
	CPUTotal       int     `json:"cpu_total"`
	MemoryTotalGB  int     `json:"memory_total_gb"`
	StorageTotalGB int     `json:"storage_total_gb"`

	CPUUsed       int     `json:"cpu_used"`
	MemoryUsedGB  float64 `json:"memory_used_gb"`
	StorageUsedGB float64 `json:"storage_used_gb"`

	OnlineNodes int `json:"online_nodes"`
	TotalNodes  int `json:"total_nodes"`
}

// NodeProfileManager manages node profiles on disk (~/.moltbunker/nodes/)
type NodeProfileManager struct {
	dir    string
	config *config.Config
	node   *Node
	cm     *ContainerManager
	mu     sync.RWMutex
	self   *NodeProfile
	peers  map[string]*NodeProfile // nodeID prefix → profile
}

// NewNodeProfileManager creates a manager and loads existing profiles from disk
func NewNodeProfileManager(dir string, cfg *config.Config, node *Node, cm *ContainerManager) (*NodeProfileManager, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	pm := &NodeProfileManager{
		dir:    dir,
		config: cfg,
		node:   node,
		cm:     cm,
		peers:  make(map[string]*NodeProfile),
	}

	pm.loadFromDisk()
	return pm, nil
}

// loadFromDisk loads existing profile JSON files
func (pm *NodeProfileManager) loadFromDisk() {
	// Load self profile
	selfPath := filepath.Join(pm.dir, "self.json")
	if data, err := os.ReadFile(selfPath); err == nil {
		var profile NodeProfile
		if err := json.Unmarshal(data, &profile); err == nil {
			pm.self = &profile
		}
	}

	// Load peer profiles
	entries, err := os.ReadDir(pm.dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "self.json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(pm.dir, entry.Name()))
		if err != nil {
			continue
		}
		var profile NodeProfile
		if err := json.Unmarshal(data, &profile); err == nil {
			prefix := entry.Name()[:len(entry.Name())-5] // strip .json
			pm.peers[prefix] = &profile
		}
	}
}

// removeFromDisk deletes a profile file from disk.
func (pm *NodeProfileManager) removeFromDisk(filename string) {
	path := filepath.Join(pm.dir, filename)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		logging.Warn("failed to remove stale node profile", "path", path, "error", err.Error())
	}
}

// saveToDisk writes a profile to disk
func (pm *NodeProfileManager) saveToDisk(filename string, profile *NodeProfile) {
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		logging.Warn("failed to marshal node profile", "file", filename, "error", err.Error())
		return
	}
	path := filepath.Join(pm.dir, filename)
	if err := os.WriteFile(path, data, 0600); err != nil {
		logging.Warn("failed to write node profile", "path", path, "error", err.Error())
	}
}

// RefreshSelf rebuilds the self profile from config + runtime state
func (pm *NodeProfileManager) RefreshSelf() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Query on-chain reputation score; fall back to persisted/default
	reputation := 1000
	if pm.self != nil && pm.self.ReputationScore > 0 {
		reputation = pm.self.ReputationScore
	}
	if ps := pm.node.PaymentService(); ps != nil {
		walletAddr := pm.node.WalletAddress()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		score, err := ps.GetReputationScore(ctx, walletAddr)
		cancel()
		if err == nil && score != nil {
			reputation = int(score.Int64())
		}
	}

	// Count deployments
	activeContainers := 0
	encryptedCount := 0
	if pm.cm != nil {
		deployments := pm.cm.ListDeployments()
		activeContainers = len(deployments)
		for _, d := range deployments {
			if d.Encrypted {
				encryptedCount++
			}
		}
	}

	// Map config HardwareProfile to daemon HardwareProfile
	cfgHW := pm.config.Node.Provider.Hardware
	hw := &HardwareProfile{
		CPUModel:         cfgHW.CPUModel,
		CPUArch:          cfgHW.CPUArch,
		CPUThreads:       cfgHW.CPUThreads,
		CPUCores:         cfgHW.CPUCores,
		CPUSockets:       cfgHW.CPUSockets,
		MemoryGB:         cfgHW.MemoryGB,
		MemoryType:       cfgHW.MemoryType,
		MemoryECC:        cfgHW.MemoryECC,
		StorageGB:        cfgHW.StorageGB,
		StorageType:      cfgHW.StorageType,
		StorageModel:     cfgHW.StorageModel,
		BandwidthMbps:    cfgHW.BandwidthMbps,
		NetworkInterface: cfgHW.NetworkInterface,
		SEVSNPSupported:  cfgHW.SEVSNPSupported,
		SEVSNPLevel:      cfgHW.SEVSNPLevel,
		TPMVersion:       cfgHW.TPMVersion,
		OS:               cfgHW.OS,
		OSVersion:        cfgHW.OSVersion,
		Kernel:           cfgHW.Kernel,
		Hostname:         cfgHW.Hostname,
	}

	loc := pm.node.nodeInfo.Location
	walletAddr := pm.node.WalletAddress()
	profile := &NodeProfile{
		NodeID:        pm.node.nodeInfo.ID.String(),
		WalletAddress: walletAddr.Hex(),
		Region:        pm.node.nodeInfo.Region,
		Country:       pm.node.nodeInfo.Country,
		Location:      &loc,
		Online:        pm.node.IsRunning(),
		LastSeen:      time.Now(),
		Capacity: CapacityProfile{
			CPUCores:      pm.config.Node.Provider.DeclaredCPU,
			MemoryGB:      pm.config.Node.Provider.DeclaredMemoryGB,
			StorageGB:     pm.config.Node.Provider.DeclaredStorageGB,
			BandwidthMbps: pm.config.Node.Provider.DeclaredBandwidth,
			Hardware:      hw,
		},
		Tier:             string(pm.config.Node.Provider.TargetTier),
		Role:             string(pm.config.Node.Role),
		ReputationScore:  reputation,
		ActiveContainers: activeContainers,
		EncryptedCount:   encryptedCount,
		Version:          "0.1.0",
	}

	// Detect runtime capabilities
	rtCaps := runtime.DetectRuntime(pm.config.Runtime.RuntimeName)
	profile.ProviderTier = rtCaps.ProviderTier
	profile.RuntimeName = rtCaps.RuntimeName
	profile.KataSupported = rtCaps.KataAvailable
	profile.SEVSNPActive = rtCaps.SEVSNPActive

	// GPU from config
	if pm.config.Node.Provider.GPUEnabled {
		profile.Capacity.GPUCount = pm.config.Node.Provider.GPUCount
		profile.Capacity.GPUModel = pm.config.Node.Provider.GPUModel
	}

	pm.self = profile
	pm.saveToDisk("self.json", profile)
}

// RefreshPeers updates peer profiles from the P2P router's peer list
func (pm *NodeProfileManager) RefreshPeers() {
	if pm.node == nil || pm.node.router == nil {
		return
	}

	peers := pm.node.router.GetPeers()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Track which peers are currently online
	currentPeers := make(map[string]bool)

	// Get stake verifier for wallet lookups
	sv := pm.node.router.StakeVerifier()

	for _, peer := range peers {
		idStr := peer.ID.String()
		prefix := idStr[:8]
		currentPeers[prefix] = true

		tier := deriveTierFromStake(peer.Capabilities.StakingAmount)

		var walletHex string
		if sv != nil {
			if wallet, ok := sv.GetWallet(peer.ID); ok {
				walletHex = wallet.Hex()
			}
		}

		peerLoc := peer.Location
		profile := &NodeProfile{
			NodeID:        idStr,
			Address:       peer.Address,
			WalletAddress: walletHex,
			Region:        peer.Region,
			Country:       peer.Country,
			Location:      &peerLoc,
			Online:        true,
			LastSeen:      peer.LastSeen,
			Tier:          string(tier),
			StakingAmount: peer.Capabilities.StakingAmount,
		}

		pm.peers[prefix] = profile
		pm.saveToDisk(prefix+".json", profile)
	}

	// Mark missing peers as offline and prune stale ones
	now := time.Now()
	const stalePeerTTL = 7 * 24 * time.Hour // 7 days
	for prefix, profile := range pm.peers {
		if !currentPeers[prefix] {
			profile.Online = false
			// Remove peers that have been offline longer than the TTL
			if now.Sub(profile.LastSeen) > stalePeerTTL {
				delete(pm.peers, prefix)
				pm.removeFromDisk(prefix + ".json")
				continue
			}
			pm.saveToDisk(prefix+".json", profile)
		}
	}
}

// deriveTierFromStake maps a staking amount to a tier name
func deriveTierFromStake(amount uint64) types.StakingTier {
	tiers := types.DefaultStakingTiers()
	// Check from highest to lowest
	for _, tier := range []types.StakingTier{
		types.StakingTierPlatinum,
		types.StakingTierGold,
		types.StakingTierSilver,
		types.StakingTierBronze,
	} {
		cfg := tiers[tier]
		if cfg.MinStake != nil && cfg.MinStake.Uint64() > 0 && amount >= cfg.MinStake.Uint64() {
			return tier
		}
		// Parse if needed
		if cfg.MinStake == nil {
			_ = cfg.ParseMinStake()
			if cfg.MinStake != nil && cfg.MinStake.Uint64() > 0 && amount >= cfg.MinStake.Uint64() {
				return tier
			}
		}
	}
	return types.StakingTierStarter
}

// GetSelf returns the cached self profile
func (pm *NodeProfileManager) GetSelf() *NodeProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.self
}

// GetAll returns self + all peer profiles
func (pm *NodeProfileManager) GetAll() []NodeProfile {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	all := make([]NodeProfile, 0, 1+len(pm.peers))
	if pm.self != nil {
		all = append(all, *pm.self)
	}
	for _, p := range pm.peers {
		all = append(all, *p)
	}
	return all
}

// GetAggregatedCapacity sums capacity across all online nodes
func (pm *NodeProfileManager) GetAggregatedCapacity() *AggregatedCapacity {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	agg := &AggregatedCapacity{}

	allNodes := make([]*NodeProfile, 0, 1+len(pm.peers))
	if pm.self != nil {
		allNodes = append(allNodes, pm.self)
	}
	for _, p := range pm.peers {
		allNodes = append(allNodes, p)
	}

	agg.TotalNodes = len(allNodes)

	for _, n := range allNodes {
		if n.Online {
			agg.OnlineNodes++
			agg.CPUTotal += n.Capacity.CPUCores
			agg.MemoryTotalGB += n.Capacity.MemoryGB
			agg.StorageTotalGB += n.Capacity.StorageGB
		}
	}

	// Used resources: only from self (container allocations)
	if pm.self != nil && pm.cm != nil {
		deployments := pm.cm.ListDeployments()
		for _, d := range deployments {
			if d.Resources.CPUQuota > 0 && d.Resources.CPUPeriod > 0 {
				agg.CPUUsed += int(d.Resources.CPUQuota / int64(d.Resources.CPUPeriod))
			}
			if d.Resources.MemoryLimit > 0 {
				agg.MemoryUsedGB += float64(d.Resources.MemoryLimit) / (1024 * 1024 * 1024)
			}
			if d.Resources.DiskLimit > 0 {
				agg.StorageUsedGB += float64(d.Resources.DiskLimit) / (1024 * 1024 * 1024)
			}
		}
	}

	return agg
}
