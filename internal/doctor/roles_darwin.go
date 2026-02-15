//go:build darwin

package doctor

// Provider-only checkers implement RoleAware to skip for requesters.

func (c *ColimaChecker) Roles() []string          { return []string{"provider", "hybrid"} }
func (c *ContainerdChecker) Roles() []string       { return []string{"provider", "hybrid"} }
func (c *IPFSChecker) Roles() []string             { return []string{"provider", "hybrid"} }
func (c *SocketPermissionChecker) Roles() []string { return []string{"provider", "hybrid"} }
