//go:build !linux

package runtime

import "context"

// AppArmorProfileName is the canonical profile name (kept defined on all platforms
// so callers can reference it without a build-tag split).
const AppArmorProfileName = "moltbunker-container"

// AppArmorLoader is a no-op on non-Linux platforms (AppArmor is Linux-only).
// Keeping the type and method set identical to the Linux build lets the doctor
// checker and daemon wiring compile without per-OS conditionals at the call site.
type AppArmorLoader struct{}

// EnsureProfile is a no-op on non-Linux platforms and always succeeds.
func (l *AppArmorLoader) EnsureProfile(_ context.Context, _, _ string) error { return nil }

// IsProfileLoaded always reports false on non-Linux platforms.
func (l *AppArmorLoader) IsProfileLoaded(_ string) bool { return false }
