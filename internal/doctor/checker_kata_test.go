//go:build linux

package doctor

import (
	"context"
	"testing"
)

func TestKataRuntimeChecker_Interface(t *testing.T) {
	checker := NewKataRuntimeChecker()

	if checker.Name() != "Kata Containers" {
		t.Errorf("Name() = %q, want %q", checker.Name(), "Kata Containers")
	}
	if checker.Category() != CategoryRuntime {
		t.Errorf("Category() = %q, want %q", checker.Category(), CategoryRuntime)
	}
	if checker.CanFix() {
		t.Error("CanFix() should return false")
	}
}

func TestKataRuntimeChecker_Check(t *testing.T) {
	ctx := context.Background()
	checker := NewKataRuntimeChecker()

	result := checker.Check(ctx)

	// Must always return a valid result
	if result.Name != "Kata Containers" {
		t.Errorf("result.Name = %q, want %q", result.Name, "Kata Containers")
	}
	if result.Category != CategoryRuntime {
		t.Errorf("result.Category = %q, want %q", result.Category, CategoryRuntime)
	}

	// Status must be one of the valid values
	switch result.Status {
	case StatusOK, StatusWarning, StatusError:
		// valid
	default:
		t.Errorf("unexpected status: %q", result.Status)
	}

	// Message must not be empty
	if result.Message == "" {
		t.Error("result.Message should not be empty")
	}
}

func TestKataRuntimeChecker_RoleAware(t *testing.T) {
	checker := NewKataRuntimeChecker()

	roles := checker.Roles()
	if len(roles) != 2 {
		t.Fatalf("Roles() returned %d roles, want 2", len(roles))
	}

	wantRoles := map[string]bool{"provider": true, "hybrid": true}
	for _, role := range roles {
		if !wantRoles[role] {
			t.Errorf("unexpected role %q", role)
		}
	}
}

func TestKataRuntimeChecker_FixReturnsError(t *testing.T) {
	checker := NewKataRuntimeChecker()
	err := checker.Fix(context.Background(), nil)
	if err == nil {
		t.Error("Fix() should return an error (not auto-fixable)")
	}
}
