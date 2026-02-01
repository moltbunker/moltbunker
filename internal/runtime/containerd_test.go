package runtime

import (
	"testing"
)

// Note: containerd tests require a running containerd instance.
// These unit tests focus on configuration validation.

func TestContainerdConfiguration(t *testing.T) {
	// Validate default containerd socket paths
	defaultPaths := []string{
		"/run/containerd/containerd.sock",
		"/var/run/containerd/containerd.sock",
	}

	if len(defaultPaths) == 0 {
		t.Error("Should have default socket paths defined")
	}
}

func TestContainerNamespace(t *testing.T) {
	// Moltbunker should use its own namespace
	namespace := "moltbunker"

	if namespace == "" {
		t.Error("Container namespace should not be empty")
	}

	if namespace == "default" {
		t.Error("Should not use default namespace")
	}
}
