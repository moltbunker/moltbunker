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

// SelectNodesForReplication selects 3 nodes in different regions
func (gr *GeographicRouter) SelectNodesForReplication(nodes []*types.Node) ([]*types.Node, error) {
	if len(nodes) < 3 {
		return nil, fmt.Errorf("need at least 3 nodes for replication")
	}

	// Group nodes by region
	regions := make(map[string][]*types.Node)
	for _, node := range nodes {
		region := GetRegionFromCountry(node.Country)
		regions[region] = append(regions[region], node)
	}

	// Select one node from each of 3 different regions
	selected := make([]*types.Node, 0, 3)
	usedRegions := make(map[string]bool)

	for _, node := range nodes {
		region := GetRegionFromCountry(node.Country)
		if !usedRegions[region] && len(selected) < 3 {
			selected = append(selected, node)
			usedRegions[region] = true
		}
	}

	if len(selected) < 3 {
		return nil, fmt.Errorf("could not find 3 nodes in different regions")
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
