package tor

import (
	"fmt"
	"os"
	"path/filepath"
)

// TorConfig represents Tor configuration
type TorConfig struct {
	DataDir         string
	ControlPort     int
	SOCKS5Port      int
	TorrcPath       string
	ExitNodeCountry string // Country code for exit node selection
	StrictNodes     bool   // Only use specified exit nodes
	EnableSocks     bool   // Enable SOCKS5 proxy
	TorOnly         bool   // Only allow Tor connections
}

// DefaultTorConfig returns default Tor configuration
func DefaultTorConfig(dataDir string) *TorConfig {
	return &TorConfig{
		DataDir:     dataDir,
		ControlPort: 9051,
		SOCKS5Port:  9050,
		TorrcPath:   filepath.Join(dataDir, "torrc"),
	}
}

// WriteTorrc writes Tor configuration file
func (tc *TorConfig) WriteTorrc() error {
	// Create data directory
	if err := os.MkdirAll(tc.DataDir, 0700); err != nil {
		return fmt.Errorf("failed to create Tor data directory: %w", err)
	}

	// Build torrc content
	content := fmt.Sprintf(`DataDirectory %s
ControlPort %d
SOCKSPort %d
`, tc.DataDir, tc.ControlPort, tc.SOCKS5Port)

	if tc.ExitNodeCountry != "" {
		content += fmt.Sprintf("ExitNodes {%s}\n", tc.ExitNodeCountry)
		if tc.StrictNodes {
			content += "StrictNodes 1\n"
		}
	}

	// Write torrc file
	if err := os.WriteFile(tc.TorrcPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write torrc: %w", err)
	}

	return nil
}

// ControlAddress returns the control port address
func (tc *TorConfig) ControlAddress() string {
	return fmt.Sprintf("127.0.0.1:%d", tc.ControlPort)
}

// SOCKS5Address returns the SOCKS5 proxy address
func (tc *TorConfig) SOCKS5Address() string {
	return fmt.Sprintf("127.0.0.1:%d", tc.SOCKS5Port)
}
