package commands

import (
	"runtime"
	"runtime/debug"
)

// Global CLI flags
var (
	// SocketPath is the path to the daemon socket
	SocketPath string
)

// Version information (set at build time)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// GetVersion returns the version string
func GetVersion() string {
	if Version != "dev" {
		return Version
	}
	// Try to get version from build info
	if info, ok := debug.ReadBuildInfo(); ok {
		if info.Main.Version != "" && info.Main.Version != "(devel)" {
			return info.Main.Version
		}
	}
	return "dev"
}

// GetCommit returns the git commit
func GetCommit() string {
	if Commit != "unknown" {
		return Commit
	}
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				if len(setting.Value) > 8 {
					return setting.Value[:8]
				}
				return setting.Value
			}
		}
	}
	return "unknown"
}

// GetGoVersion returns the Go version
func GetGoVersion() string {
	return runtime.Version()
}
