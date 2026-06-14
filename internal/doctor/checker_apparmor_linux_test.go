//go:build linux

package doctor

import (
	"context"
	"errors"
	"os"
	"testing"
)

// fakeAALoader implements appArmorLoader for tests.
type fakeAALoader struct {
	loaded     bool
	ensureErr  error
	ensureCall bool
}

func (f *fakeAALoader) IsProfileLoaded(string) bool { return f.loaded }
func (f *fakeAALoader) EnsureProfile(context.Context, string, string) error {
	f.ensureCall = true
	return f.ensureErr
}

func okStat(string) (os.FileInfo, error)  { return nil, nil }
func errStat(string) (os.FileInfo, error) { return nil, errors.New("not found") }
func okLook(string) (string, error)       { return "/usr/sbin/apparmor_parser", nil }
func errLook(string) (string, error)      { return "", errors.New("not found") }

func TestAppArmorChecker_SkippedWhenNoLSM(t *testing.T) {
	c := &AppArmorChecker{statFn: errStat, lookPathFn: okLook, loader: &fakeAALoader{}}
	res := c.Check(context.Background())
	if res.Status != StatusSkipped {
		t.Fatalf("expected StatusSkipped when AppArmor LSM absent, got %s", res.Status)
	}
}

func TestAppArmorChecker_OKWhenLoaded(t *testing.T) {
	c := &AppArmorChecker{statFn: okStat, lookPathFn: okLook, loader: &fakeAALoader{loaded: true}}
	res := c.Check(context.Background())
	if res.Status != StatusOK {
		t.Fatalf("expected StatusOK when profile loaded, got %s", res.Status)
	}
}

func TestAppArmorChecker_ErrorWhenParserMissing(t *testing.T) {
	c := &AppArmorChecker{statFn: okStat, lookPathFn: errLook, loader: &fakeAALoader{loaded: false}}
	res := c.Check(context.Background())
	if res.Status != StatusError {
		t.Fatalf("expected StatusError when parser missing, got %s", res.Status)
	}
	if res.Fixable {
		t.Error("checker should not be auto-fixable when apparmor_parser is missing")
	}
}

func TestAppArmorChecker_WarnWhenNotLoaded(t *testing.T) {
	c := &AppArmorChecker{statFn: okStat, lookPathFn: okLook, loader: &fakeAALoader{loaded: false}}
	res := c.Check(context.Background())
	if res.Status != StatusWarning {
		t.Fatalf("expected StatusWarning when parser present but profile not loaded, got %s", res.Status)
	}
	if !res.Fixable {
		t.Error("checker should be fixable when the parser is present")
	}
}

func TestAppArmorChecker_FixCallsLoader(t *testing.T) {
	fl := &fakeAALoader{loaded: false}
	c := &AppArmorChecker{statFn: okStat, lookPathFn: okLook, loader: fl}
	if err := c.Fix(context.Background(), nil); err != nil {
		t.Fatalf("Fix returned error: %v", err)
	}
	if !fl.ensureCall {
		t.Error("Fix should call EnsureProfile on the loader")
	}
}
