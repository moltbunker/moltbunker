package runtime

import (
	"context"
	"testing"

	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// applyKataAnnotations builds the kata annotation SpecOpt for the given config and
// applies it to a fresh spec, returning the resulting annotations map (nil if the
// opt itself was nil).
func applyKataAnnotations(t *testing.T, cfg *KataConfig) map[string]string {
	t.Helper()
	cc := &ContainerdClient{kataConfig: cfg}
	opt := cc.kataAnnotations()
	if opt == nil {
		return nil
	}
	s := &specs.Spec{}
	if err := opt(context.Background(), nil, nil, s); err != nil {
		t.Fatalf("apply kata annotations: %v", err)
	}
	return s.Annotations
}

func TestKataAnnotations_DefaultPIDs(t *testing.T) {
	ann := applyKataAnnotations(t, &KataConfig{DefaultPIDs: 1024})
	if ann == nil {
		t.Fatal("expected annotations for non-zero DefaultPIDs")
	}
	got := ann["io.katacontainers.config.hypervisor.default_pids"]
	if got != "1024" {
		t.Errorf("default_pids annotation = %q, want %q", got, "1024")
	}
}

func TestKataAnnotations_DefaultPIDsZeroOmitted(t *testing.T) {
	// With every field zero, kataAnnotations returns a nil opt (nothing to emit).
	if ann := applyKataAnnotations(t, &KataConfig{}); ann != nil {
		if _, ok := ann["io.katacontainers.config.hypervisor.default_pids"]; ok {
			t.Error("default_pids should be omitted when DefaultPIDs == 0")
		}
	}
}

func TestKataAnnotations_DefaultPIDsAlongsideOthers(t *testing.T) {
	ann := applyKataAnnotations(t, &KataConfig{VMMemoryMB: 512, VMCPUs: 2, DefaultPIDs: 2048})
	if ann["io.katacontainers.config.hypervisor.default_memory"] != "512" {
		t.Error("expected default_memory annotation")
	}
	if ann["io.katacontainers.config.hypervisor.default_vcpus"] != "2" {
		t.Error("expected default_vcpus annotation")
	}
	if ann["io.katacontainers.config.hypervisor.default_pids"] != "2048" {
		t.Error("expected default_pids annotation alongside the others")
	}
}
