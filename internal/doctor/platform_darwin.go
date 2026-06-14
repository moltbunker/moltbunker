//go:build darwin

package doctor

func (d *Doctor) initPackageManager() {
	d.packageManager = NewHomebrewManager()
}

func (d *Doctor) registerPlatformCheckers() {
	d.checkers = []Checker{
		// Universal checks (all roles)
		NewConfigFileChecker(),
		NewNodeKeysChecker(),
		NewWalletChecker(),
		NewGoVersionChecker(),
		NewDiskSpaceChecker(),
		NewMemoryChecker(),
		NewFileDescriptorChecker(),
		// Provider-only checks (filtered by RoleAware)
		NewColimaChecker(),
		NewContainerdChecker(),
		NewIPFSChecker(),
		NewSocketPermissionChecker(),
		// RUN-01: expose-to-internet security-gate tooling. trivy/cosign can be
		// present on darwin dev machines; nft is Skipped on non-Linux but stays
		// visible in `moltbunker doctor` output.
		NewTrivyChecker(),                 // R4 image scanning
		NewNftChecker(),                   // R13/R14 (Skipped on darwin)
		NewImageSignatureToolingChecker(), // R3 image signature verification
		// HARDEN-01: runtime isolation hardening checks (Skipped on darwin, but
		// kept visible in `moltbunker doctor` output for parity).
		NewAppArmorChecker(),    // R9 (Skipped on darwin)
		NewUserNSChecker(),      // R12 (Skipped on darwin)
		NewKataPIDsChecker(nil), // R17 (Skipped on darwin)
		// Optional services
		NewTorChecker(),
	}
}
