package p2p

import (
	"testing"

	"github.com/moltbunker/moltbunker/pkg/types"
)

func createTestNode(id string, country string) *types.Node {
	var nodeID types.NodeID
	idBytes := []byte(id)
	if len(idBytes) > 32 {
		copy(nodeID[:], idBytes[:32])
	} else {
		copy(nodeID[:], idBytes)
	}
	return &types.Node{
		ID:      nodeID,
		Country: country,
		Region:  GetRegionFromCountry(country),
	}
}

func TestGeographicRouter_SelectNodesForReplication(t *testing.T) {
	geolocator := NewGeoLocator()
	gr := NewGeographicRouter(geolocator)

	nodes := []*types.Node{
		createTestNode("node1", "US"),
		createTestNode("node2", "GB"),
		createTestNode("node3", "CN"),
		createTestNode("node4", "DE"),
	}

	selected, err := gr.SelectNodesForReplication(nodes)
	if err != nil {
		t.Fatalf("Failed to select nodes: %v", err)
	}

	if len(selected) != 3 {
		t.Errorf("Should select 3 nodes, got %d", len(selected))
	}

	// Verify they're in different regions
	regions := make(map[string]bool)
	for _, node := range selected {
		region := GetRegionFromCountry(node.Country)
		if regions[region] {
			t.Errorf("Duplicate region selected: %s", region)
		}
		regions[region] = true
	}
}

func TestGeographicRouter_SelectNodesForReplication_InsufficientNodes(t *testing.T) {
	geolocator := NewGeoLocator()
	gr := NewGeographicRouter(geolocator)

	nodes := []*types.Node{
		createTestNode("node1", "US"),
		createTestNode("node2", "GB"),
	}

	_, err := gr.SelectNodesForReplication(nodes)
	if err == nil {
		t.Error("Should fail with insufficient nodes")
	}
}

func TestGeographicRouter_SelectNodesForReplication_SameRegion(t *testing.T) {
	geolocator := NewGeoLocator()
	gr := NewGeographicRouter(geolocator)

	nodes := []*types.Node{
		createTestNode("node1", "US"),
		createTestNode("node2", "CA"),
		createTestNode("node3", "MX"),
	}

	_, err := gr.SelectNodesForReplication(nodes)
	if err == nil {
		t.Error("Should fail when all nodes are in same region")
	}
}

func TestGeographicRouter_FindReplacementNode(t *testing.T) {
	geolocator := NewGeoLocator()
	gr := NewGeographicRouter(geolocator)

	nodes := []*types.Node{
		createTestNode("node1", "US"),
		createTestNode("node2", "GB"),
		createTestNode("node3", "CN"),
		createTestNode("node4", "CA"), // Another Americas node for replacement
	}

	var excludeID types.NodeID
	copy(excludeID[:], []byte("node1"))

	replacement, err := gr.FindReplacementNode("Americas", nodes, []types.NodeID{excludeID})
	if err != nil {
		t.Fatalf("Failed to find replacement: %v", err)
	}

	if replacement == nil {
		t.Error("Should find a replacement node")
	}

	if GetRegionFromCountry(replacement.Country) != "Americas" {
		t.Errorf("Replacement should be in Americas, got %s", replacement.Country)
	}
}

func TestGeographicRouter_FindReplacementNode_NotFound(t *testing.T) {
	geolocator := NewGeoLocator()
	gr := NewGeographicRouter(geolocator)

	nodes := []*types.Node{
		createTestNode("node1", "US"),
	}

	_, err := gr.FindReplacementNode("Europe", nodes, []types.NodeID{})
	if err == nil {
		t.Error("Should fail when no node found in target region")
	}
}
