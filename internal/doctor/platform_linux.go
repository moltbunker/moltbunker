//go:build linux

package doctor

func (d *Doctor) initPackageManager() {
	// Linux uses apt/yum directly — no auto-install package manager
}

func (d *Doctor) registerPlatformCheckers() {
	d.checkers = []Checker{
		NewNodeKeysChecker(),
		NewWalletChecker(),
		NewFileDescriptorChecker(),
		// Provider-only checks (filtered by RoleAware)
		NewKataRuntimeChecker(),
		// RUN-01: tooling for the expose-to-internet security gates. Each warns
		// (not errors) if its binary is absent so an operator does not enable a
		// gate that then silently no-ops.
		NewTrivyChecker(),                 // R4 image scanning
		NewNftChecker(),                   // R13/R14 nftables enforcement
		NewImageSignatureToolingChecker(), // R3 image signature verification
	}
}

// Provider-only checkers implement RoleAware to skip for requesters.

func (c *KataRuntimeChecker) Roles() []string { return []string{"provider", "hybrid"} }
