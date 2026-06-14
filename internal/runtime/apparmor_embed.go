package runtime

import _ "embed"

// embeddedAppArmorProfile is the canonical moltbunker-container AppArmor policy,
// embedded so the daemon binary is self-contained and can load the profile on
// first start without an operator pre-installing it under /etc/apparmor.d
// (HARDEN-01, R9). The bytes are used only by the Linux AppArmorLoader; the embed
// itself is pure Go and compiles on every platform (darwin reads but never loads
// them). The source of truth is configs/apparmor/moltbunker-container; this asset
// is a build-time copy kept in-package because go:embed cannot reach outside the
// package directory.
//
//go:embed embedded_profiles/moltbunker-container.aaprofile
var embeddedAppArmorProfile []byte

// EmbeddedAppArmorProfile returns a copy of the embedded profile bytes. Exposed
// for tests and tooling that want to assert on the shipped policy content without
// reading the filesystem.
func EmbeddedAppArmorProfile() []byte {
	out := make([]byte, len(embeddedAppArmorProfile))
	copy(out, embeddedAppArmorProfile)
	return out
}
