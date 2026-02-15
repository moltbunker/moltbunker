package doctor

import (
	"context"
)

// Category represents a category of checks
type Category string

const (
	CategoryRuntime     Category = "runtime"
	CategoryServices    Category = "services"
	CategorySystem      Category = "system"
	CategoryConfig      Category = "config"
	CategoryPermissions Category = "permissions"
)

// Status represents the result status of a check
type Status string

const (
	StatusOK      Status = "ok"
	StatusWarning Status = "warning"
	StatusError   Status = "error"
	StatusSkipped Status = "skipped"
)

// CheckResult represents the result of a single check
type CheckResult struct {
	Name        string   `json:"name"`
	Category    Category `json:"category"`
	Status      Status   `json:"status"`
	Message     string   `json:"message"`
	Details     string   `json:"details,omitempty"`
	Fixable     bool     `json:"fixable"`
	FixCommand  string   `json:"fix_command,omitempty"`
	FixPackage  string   `json:"fix_package,omitempty"`
}

// Checker is the interface that all checkers must implement
type Checker interface {
	// Name returns the display name of the checker
	Name() string
	// Category returns the category this checker belongs to
	Category() Category
	// Check performs the check and returns the result
	Check(ctx context.Context) CheckResult
	// CanFix returns true if the checker can auto-fix issues
	CanFix() bool
	// Fix attempts to fix the issue (only called if CanFix returns true)
	Fix(ctx context.Context, pm PackageManager) error
}

// PackageManager is the interface for installing packages
type PackageManager interface {
	// Name returns the package manager name (e.g., "homebrew")
	Name() string
	// IsAvailable checks if the package manager is available
	IsAvailable() bool
	// Install installs a package
	Install(ctx context.Context, pkg string) error
	// IsInstalled checks if a package is installed
	IsInstalled(pkg string) bool
}

// RoleAware is an optional interface for checkers that only apply to specific roles.
// Checkers that don't implement this run for all roles.
type RoleAware interface {
	// Roles returns the roles this checker applies to (e.g., ["provider", "hybrid"])
	Roles() []string
}

// DoctorOptions configures the doctor run
type DoctorOptions struct {
	// Fix enables auto-fix mode
	Fix bool
	// DryRun shows what would be done without making changes
	DryRun bool
	// JSON outputs results as JSON
	JSON bool
	// Category filters checks to a specific category
	Category Category
	// Role filters checks to a specific node role
	Role string
	// Verbose enables verbose output
	Verbose bool
}

// DoctorReport is the complete report from running all checks
type DoctorReport struct {
	Checks  []CheckResult `json:"checks"`
	Summary Summary       `json:"summary"`
}

// Summary provides an overview of the check results
type Summary struct {
	Total   int `json:"total"`
	Passed  int `json:"passed"`
	Failed  int `json:"failed"`
	Warned  int `json:"warned"`
	Skipped int `json:"skipped"`
	Fixable int `json:"fixable"`
}

// IsHealthy returns true if all checks passed or warned
func (s Summary) IsHealthy() bool {
	return s.Failed == 0
}
