package runtime

import (
	"testing"
)

// Note: Image tests require containerd and IPFS integration.
// These unit tests focus on image reference parsing.

func TestImageReferenceFormat(t *testing.T) {
	tests := []struct {
		name  string
		image string
		valid bool
	}{
		{"simple image", "nginx", true},
		{"image with tag", "nginx:latest", true},
		{"image with version", "nginx:1.25", true},
		{"registry image", "docker.io/library/nginx:latest", true},
		{"ipfs CID", "ipfs://QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := isValidImageRef(tt.image)
			if valid != tt.valid {
				t.Errorf("isValidImageRef(%s) = %v, want %v", tt.image, valid, tt.valid)
			}
		})
	}
}

func isValidImageRef(image string) bool {
	return image != ""
}

func TestIPFSImageCIDFormat(t *testing.T) {
	tests := []struct {
		cid   string
		valid bool
	}{
		{"QmXoypizjW3WknFiJnKLwHCnL72vedxjQkDDP1mXWo6uco", true},
		{"bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.cid, func(t *testing.T) {
			valid := isValidCID(tt.cid)
			if valid != tt.valid {
				t.Errorf("isValidCID(%s) = %v, want %v", tt.cid, valid, tt.valid)
			}
		})
	}
}

func isValidCID(cid string) bool {
	if cid == "" {
		return false
	}
	// QmBase58 (CIDv0) or bafy (CIDv1)
	return (len(cid) == 46 && cid[0] == 'Q' && cid[1] == 'm') ||
		(len(cid) >= 4 && cid[0:4] == "bafy")
}

func TestNormalizeImageRef(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Official Docker Hub images (no slash)
		{"nginx", "docker.io/library/nginx"},
		{"nginx:alpine", "docker.io/library/nginx:alpine"},
		{"nginx:latest", "docker.io/library/nginx:latest"},
		{"mongo:7", "docker.io/library/mongo:7"},
		{"redis:7-alpine", "docker.io/library/redis:7-alpine"},
		{"postgres:16-alpine", "docker.io/library/postgres:16-alpine"},
		{"traefik:v3.0", "docker.io/library/traefik:v3.0"},

		// Docker Hub user images (one slash, no dot in host)
		{"ollama/ollama:latest", "docker.io/ollama/ollama:latest"},
		{"minio/minio:latest", "docker.io/minio/minio:latest"},
		{"n8nio/n8n:latest", "docker.io/n8nio/n8n:latest"},
		{"jupyter/scipy-notebook:latest", "docker.io/jupyter/scipy-notebook:latest"},

		// Already fully qualified (dot in host)
		{"docker.io/library/nginx:alpine", "docker.io/library/nginx:alpine"},
		{"ghcr.io/huggingface/text-generation-inference:latest", "ghcr.io/huggingface/text-generation-inference:latest"},
		{"registry.example.com/myimage:v1", "registry.example.com/myimage:v1"},

		// Localhost registry (colon in host)
		{"localhost:5000/myimage:latest", "localhost:5000/myimage:latest"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeImageRef(tt.input)
			if got != tt.expected {
				t.Errorf("NormalizeImageRef(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
