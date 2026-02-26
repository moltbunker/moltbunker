package ingress

import (
	"testing"
	"time"
)

func TestResolveByPrefix_MinimumLength(t *testing.T) {
	r := NewResolver(nil, nil)

	// Register a service with deployment ID "dep-a1b2c3d4e5f6abcd"
	r.Register(&ServiceEntry{
		DeploymentID:   "dep-a1b2c3d4e5f6abcd",
		ProviderNodeID: "node123",
		ProviderAddr:   "10.0.0.1:9002",
		ContainerPort:  8080,
		HostPort:       32000,
	})

	tests := []struct {
		name      string
		prefix    string
		wantMatch bool
	}{
		{"empty prefix rejected", "", false},
		{"1-char prefix rejected", "a", false},
		{"2-char prefix rejected", "a1", false},
		{"3-char prefix rejected", "a1b", false},
		{"7-char prefix rejected", "a1b2c3d", false},
		{"8-char prefix accepted", "a1b2c3d4", true},
		{"full ID prefix accepted", "a1b2c3d4e5f6abcd", true},
		{"8-char non-matching rejected", "ffffffff", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.resolveByPrefix(tt.prefix)
			if tt.wantMatch && got == nil {
				t.Errorf("resolveByPrefix(%q) = nil, want match", tt.prefix)
			}
			if !tt.wantMatch && got != nil {
				t.Errorf("resolveByPrefix(%q) = %v, want nil", tt.prefix, got.DeploymentID)
			}
		})
	}
}

func TestResolveByPrefix_ExpiredEntry(t *testing.T) {
	r := NewResolver(nil, nil)

	r.mu.Lock()
	r.services["dep-a1b2c3d4e5f6"] = &ServiceEntry{
		DeploymentID: "dep-a1b2c3d4e5f6",
		LastSeen:     time.Now().Add(-10 * time.Minute), // expired
	}
	r.mu.Unlock()

	got := r.resolveByPrefix("a1b2c3d4")
	if got != nil {
		t.Error("resolveByPrefix returned expired entry")
	}
}
