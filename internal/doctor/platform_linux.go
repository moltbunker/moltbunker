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
	}
}

// Provider-only checkers implement RoleAware to skip for requesters.

func (c *KataRuntimeChecker) Roles() []string { return []string{"provider", "hybrid"} }
