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
		// Optional services
		NewTorChecker(),
	}
}
