package p2p

import (
	"sync"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// PeerDiversityEnforcer ensures the peer table isn't dominated by a single
// region or subnet, defending against eclipse attacks.
type PeerDiversityEnforcer struct {
	mu             sync.RWMutex
	regionCounts   map[string]int // region → peer count
	subnetCounts   map[string]int // /16 prefix → peer count
	totalPeers     int
	maxRegionShare float64 // max fraction of peers from one region (default: 0.5)
	maxSubnetShare float64 // max fraction of peers from one /16 (default: 0.3)
	minRegions     int     // minimum distinct regions desired (default: 2)
}

// NewPeerDiversityEnforcer creates a new enforcer with default thresholds.
func NewPeerDiversityEnforcer() *PeerDiversityEnforcer {
	return &PeerDiversityEnforcer{
		regionCounts:   make(map[string]int),
		subnetCounts:   make(map[string]int),
		maxRegionShare: 0.5,
		maxSubnetShare: 0.3,
		minRegions:     2,
	}
}

// NewPeerDiversityEnforcerWithConfig creates an enforcer with custom thresholds.
func NewPeerDiversityEnforcerWithConfig(maxRegionShare, maxSubnetShare float64, minRegions int) *PeerDiversityEnforcer {
	pde := NewPeerDiversityEnforcer()
	pde.maxRegionShare = maxRegionShare
	pde.maxSubnetShare = maxSubnetShare
	pde.minRegions = minRegions
	return pde
}

// AllowPeer checks whether adding a peer from the given region and /16 subnet
// would violate diversity constraints. Returns true if the peer should be accepted.
func (pde *PeerDiversityEnforcer) AllowPeer(region, subnet string) bool {
	pde.mu.RLock()
	defer pde.mu.RUnlock()

	// Always allow if we have very few peers (bootstrapping)
	if pde.totalPeers < 4 {
		return true
	}

	newTotal := pde.totalPeers + 1

	// Check region share
	if region != "" && region != "Unknown" {
		regionCount := pde.regionCounts[region] + 1
		if float64(regionCount)/float64(newTotal) > pde.maxRegionShare {
			logging.Debug("peer diversity: region share exceeded",
				"region", region,
				"count", regionCount,
				"total", newTotal,
				"max_share", pde.maxRegionShare,
				logging.Component("diversity"))
			return false
		}
	}

	// Check /16 subnet share
	if subnet != "" {
		subnetCount := pde.subnetCounts[subnet] + 1
		if float64(subnetCount)/float64(newTotal) > pde.maxSubnetShare {
			logging.Debug("peer diversity: subnet share exceeded",
				"subnet", subnet,
				"count", subnetCount,
				"total", newTotal,
				"max_share", pde.maxSubnetShare,
				logging.Component("diversity"))
			return false
		}
	}

	return true
}

// RecordPeer records a new peer for diversity tracking.
func (pde *PeerDiversityEnforcer) RecordPeer(region, subnet string) {
	pde.mu.Lock()
	defer pde.mu.Unlock()

	pde.totalPeers++
	if region != "" {
		pde.regionCounts[region]++
	}
	if subnet != "" {
		pde.subnetCounts[subnet]++
	}
}

// RemovePeer removes a peer from diversity tracking.
func (pde *PeerDiversityEnforcer) RemovePeer(region, subnet string) {
	pde.mu.Lock()
	defer pde.mu.Unlock()

	pde.totalPeers--
	if pde.totalPeers < 0 {
		pde.totalPeers = 0
	}

	if region != "" {
		pde.regionCounts[region]--
		if pde.regionCounts[region] <= 0 {
			delete(pde.regionCounts, region)
		}
	}
	if subnet != "" {
		pde.subnetCounts[subnet]--
		if pde.subnetCounts[subnet] <= 0 {
			delete(pde.subnetCounts, subnet)
		}
	}
}

// RegionCount returns the number of distinct regions tracked.
func (pde *PeerDiversityEnforcer) RegionCount() int {
	pde.mu.RLock()
	defer pde.mu.RUnlock()
	return len(pde.regionCounts)
}

// NeedsDiversity returns true if we have fewer regions than desired.
func (pde *PeerDiversityEnforcer) NeedsDiversity() bool {
	pde.mu.RLock()
	defer pde.mu.RUnlock()
	return len(pde.regionCounts) < pde.minRegions && pde.totalPeers > 0
}

// TotalPeers returns the total number of tracked peers.
func (pde *PeerDiversityEnforcer) TotalPeers() int {
	pde.mu.RLock()
	defer pde.mu.RUnlock()
	return pde.totalPeers
}
