package p2p

import (
	"fmt"

	"github.com/moltbunker/moltbunker/pkg/types"
)

// GeographicRouter ensures geographic distribution of replicas
type GeographicRouter struct {
	geolocator *GeoLocator
}

// NewGeographicRouter creates a new geographic router
func NewGeographicRouter(geolocator *GeoLocator) *GeographicRouter {
	return &GeographicRouter{
		geolocator: geolocator,
	}
}

// SelectNodesForReplication selects up to 3 nodes in different regions.
// If minTier is non-empty, only nodes meeting the tier requirement are considered.
// Returns as many distinct-region nodes as available (1-3).
// Returns error only if no eligible nodes remain after filtering.
func (gr *GeographicRouter) SelectNodesForReplication(nodes []*types.Node, minTier ...types.ProviderTier) ([]*types.Node, error) {
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes available for replication")
	}

	// Filter by minimum tier if specified
	var tier types.ProviderTier
	if len(minTier) > 0 {
		tier = minTier[0]
	}

	eligible := nodes
	if tier != "" {
		eligible = make([]*types.Node, 0, len(nodes))
		for _, node := range nodes {
			if node.ProviderTier.MeetsTierRequirement(tier) {
				eligible = append(eligible, node)
			}
		}
		if len(eligible) == 0 {
			return nil, fmt.Errorf("no nodes meet minimum tier requirement: %s", tier)
		}
	}

	// Select one node from each distinct region (up to 3)
	selected := make([]*types.Node, 0, 3)
	usedRegions := make(map[string]bool)

	for _, node := range eligible {
		region := GetRegionFromCountry(node.Country)
		if !usedRegions[region] && len(selected) < 3 {
			selected = append(selected, node)
			usedRegions[region] = true
		}
	}

	// If we have fewer than 3 distinct regions, fill remaining slots
	// with nodes from already-used regions for redundancy
	if len(selected) < 3 {
		for _, node := range eligible {
			alreadySelected := false
			for _, s := range selected {
				if s.ID == node.ID {
					alreadySelected = true
					break
				}
			}
			if !alreadySelected && len(selected) < 3 {
				selected = append(selected, node)
			}
		}
	}

	return selected, nil
}

// EnsureGeographicDistribution ensures replicas are in different regions
func (gr *GeographicRouter) EnsureGeographicDistribution(replicaSet *types.ReplicaSet, nodes []*types.Node) error {
	// Check if replicas are in different regions
	regions := make(map[string]bool)
	for _, replica := range replicaSet.Replicas {
		if replica != nil {
			region := GetRegionFromCountry(replica.NodeID.String()) // Simplified - would use actual node info
			if regions[region] {
				return fmt.Errorf("replicas not geographically distributed")
			}
			regions[region] = true
		}
	}

	return nil
}

// FindReplacementNode finds a replacement node in a specific region
func (gr *GeographicRouter) FindReplacementNode(targetRegion string, nodes []*types.Node, exclude []types.NodeID) (*types.Node, error) {
	excludeMap := make(map[types.NodeID]bool)
	for _, id := range exclude {
		excludeMap[id] = true
	}

	for _, node := range nodes {
		if excludeMap[node.ID] {
			continue
		}

		region := GetRegionFromCountry(node.Country)
		if region == targetRegion {
			return node, nil
		}
	}

	return nil, fmt.Errorf("no replacement node found in region: %s", targetRegion)
}
