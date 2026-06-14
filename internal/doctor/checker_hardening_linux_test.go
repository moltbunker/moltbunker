//go:build linux

package doctor

import (
	"context"
	"testing"

	"github.com/moltbunker/moltbunker/internal/runtime"
)

func TestUserNSChecker_OKWhenSupported(t *testing.T) {
	c := &UserNSChecker{compatFn: func() runtime.UserNSCompatResult {
		return runtime.UserNSCompatResult{Supported: true}
	}}
	if res := c.Check(context.Background()); res.Status != StatusOK {
		t.Fatalf("expected StatusOK when userns supported, got %s", res.Status)
	}
}

func TestUserNSChecker_WarnWhenUnsupported(t *testing.T) {
	c := &UserNSChecker{compatFn: func() runtime.UserNSCompatResult {
		return runtime.UserNSCompatResult{Supported: false, Reason: "max_user_namespaces=0"}
	}}
	res := c.Check(context.Background())
	if res.Status != StatusWarning {
		t.Fatalf("expected StatusWarning when userns unsupported, got %s", res.Status)
	}
	if res.Details == "" {
		t.Error("warning should carry remediation details")
	}
}

func TestKataPIDsChecker_SkippedWhenNoKata(t *testing.T) {
	c := &KataPIDsChecker{kataAvailableFn: func() bool { return false }}
	if res := c.Check(context.Background()); res.Status != StatusSkipped {
		t.Fatalf("expected StatusSkipped when Kata absent, got %s", res.Status)
	}
}

func TestKataPIDsChecker_WarnWhenNoLimitAtAll(t *testing.T) {
	// No effective OCI pids.limit and no Kata annotation -> warn.
	c := &KataPIDsChecker{kataAvailableFn: func() bool { return true }, cfg: nil, ociPIDLimit: 0}
	if res := c.Check(context.Background()); res.Status != StatusWarning {
		t.Fatalf("expected StatusWarning when no PID limit at all, got %s", res.Status)
	}

	c2 := &KataPIDsChecker{kataAvailableFn: func() bool { return true }, cfg: &runtime.KataConfig{DefaultPIDs: 0}, ociPIDLimit: 0}
	if res := c2.Check(context.Background()); res.Status != StatusWarning {
		t.Fatalf("expected StatusWarning when DefaultPIDs == 0 and no OCI limit, got %s", res.Status)
	}
}

// TestKataPIDsChecker_OKWhenOCILimitSet is the R17 correction: when the daemon's
// effective OCI pids.limit is set, the kata-agent enforces it in-guest, so the
// workload IS bounded even with no (inert) hypervisor annotation. The checker must
// NOT warn in that case.
func TestKataPIDsChecker_OKWhenOCILimitSet(t *testing.T) {
	c := &KataPIDsChecker{kataAvailableFn: func() bool { return true }, cfg: nil, ociPIDLimit: 100}
	res := c.Check(context.Background())
	if res.Status != StatusOK {
		t.Fatalf("expected StatusOK when OCI pids.limit is set, got %s", res.Status)
	}
}

func TestKataPIDsChecker_OKWhenAnnotationSet(t *testing.T) {
	// No OCI limit known, but a Kata annotation is configured: informational OK
	// (annotation is a forward-looking hint, inert without enable_annotations).
	c := &KataPIDsChecker{kataAvailableFn: func() bool { return true }, cfg: &runtime.KataConfig{DefaultPIDs: 1024}, ociPIDLimit: 0}
	if res := c.Check(context.Background()); res.Status != StatusOK {
		t.Fatalf("expected StatusOK when DefaultPIDs annotation set, got %s", res.Status)
	}
}

func TestSetKataConfig_ReplacesChecker(t *testing.T) {
	d := New(DoctorOptions{})
	d.SetKataConfig(&runtime.KataConfig{DefaultPIDs: 2048}, 100)

	var found *KataPIDsChecker
	for _, c := range d.checkers {
		if kc, ok := c.(*KataPIDsChecker); ok {
			found = kc
		}
	}
	if found == nil {
		t.Fatal("expected a KataPIDsChecker to be registered")
	}
	if found.cfg == nil || found.cfg.DefaultPIDs != 2048 {
		t.Errorf("SetKataConfig did not inject the config (got %+v)", found.cfg)
	}
	if found.ociPIDLimit != 100 {
		t.Errorf("SetKataConfig did not inject the OCI PID limit (got %d, want 100)", found.ociPIDLimit)
	}
}
