//go:build darwin

package doctor

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// HomebrewManager implements PackageManager for macOS using Homebrew
type HomebrewManager struct {
	brewPath string
}

// NewHomebrewManager creates a new Homebrew package manager
func NewHomebrewManager() *HomebrewManager {
	brewPath, _ := exec.LookPath("brew")
	return &HomebrewManager{
		brewPath: brewPath,
	}
}

// Name returns the package manager name
func (h *HomebrewManager) Name() string {
	return "homebrew"
}

// IsAvailable checks if Homebrew is installed
func (h *HomebrewManager) IsAvailable() bool {
	return h.brewPath != ""
}

// Install installs a package using Homebrew
func (h *HomebrewManager) Install(ctx context.Context, pkg string) error {
	if !h.IsAvailable() {
		return fmt.Errorf("homebrew is not installed; install it from https://brew.sh")
	}

	// Map package names to Homebrew formula names if needed
	formulaName := mapPackageToFormula(pkg)

	cmd := exec.CommandContext(ctx, h.brewPath, "install", formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install %s: %w\n%s", pkg, err, string(output))
	}

	return nil
}

// IsInstalled checks if a package is installed via Homebrew
func (h *HomebrewManager) IsInstalled(pkg string) bool {
	if !h.IsAvailable() {
		return false
	}

	formulaName := mapPackageToFormula(pkg)
	cmd := exec.Command(h.brewPath, "list", formulaName)
	err := cmd.Run()
	return err == nil
}

// mapPackageToFormula maps common package names to their Homebrew formula names
func mapPackageToFormula(pkg string) string {
	mapping := map[string]string{
		"go":         "go",
		"containerd": "containerd",
		"tor":        "tor",
		"ipfs":       "ipfs",
	}

	if formula, ok := mapping[strings.ToLower(pkg)]; ok {
		return formula
	}
	return pkg
}

// GetBrewPath returns the path to the brew executable
func (h *HomebrewManager) GetBrewPath() string {
	return h.brewPath
}

// Update runs brew update to refresh package lists
func (h *HomebrewManager) Update(ctx context.Context) error {
	if !h.IsAvailable() {
		return fmt.Errorf("homebrew is not installed")
	}

	cmd := exec.CommandContext(ctx, h.brewPath, "update")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update homebrew: %w\n%s", err, string(output))
	}

	return nil
}

// Upgrade upgrades a specific package
func (h *HomebrewManager) Upgrade(ctx context.Context, pkg string) error {
	if !h.IsAvailable() {
		return fmt.Errorf("homebrew is not installed")
	}

	formulaName := mapPackageToFormula(pkg)
	cmd := exec.CommandContext(ctx, h.brewPath, "upgrade", formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upgrade %s: %w\n%s", pkg, err, string(output))
	}

	return nil
}
