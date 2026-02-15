package client

import (
	"fmt"
)

// ClientMode indicates how the SmartClient is communicating.
type ClientMode string

const (
	ModeDaemon ClientMode = "daemon" // Unix socket to local daemon
	ModeAPI    ClientMode = "api"    // HTTP API with wallet auth
	ModeNone   ClientMode = "none"   // Not connected
)

// SmartClient wraps DaemonClient and APIClient with automatic fallback.
// It tries the daemon socket first; if unavailable, falls back to HTTP API.
type SmartClient struct {
	daemon *DaemonClient
	api    *APIClient
	mode   ClientMode
}

// NewSmartClient creates a SmartClient that tries daemon first, then API.
// Either or both of daemon/api may be nil.
func NewSmartClient(daemon *DaemonClient, api *APIClient) *SmartClient {
	return &SmartClient{
		daemon: daemon,
		api:    api,
		mode:   ModeNone,
	}
}

// Connect establishes the best available connection.
// Returns nil if any transport is available.
func (sc *SmartClient) Connect() error {
	// Try daemon first
	if sc.daemon != nil {
		if err := sc.daemon.Connect(); err == nil {
			sc.mode = ModeDaemon
			return nil
		}
	}

	// Fall back to API
	if sc.api != nil {
		sc.mode = ModeAPI
		return nil
	}

	sc.mode = ModeNone
	return fmt.Errorf("no daemon running and no API configured.\nRun 'moltbunker init' to set up, or 'moltbunker start' to start the daemon")
}

// Close closes the active connection.
func (sc *SmartClient) Close() error {
	if sc.mode == ModeDaemon && sc.daemon != nil {
		return sc.daemon.Close()
	}
	return nil
}

// Mode returns the current connection mode.
func (sc *SmartClient) Mode() ClientMode {
	return sc.mode
}

// RequireDaemon returns an error if the daemon is not connected.
// Use this for provider-only operations that cannot work via HTTP API.
func (sc *SmartClient) RequireDaemon() error {
	if sc.mode != ModeDaemon {
		return fmt.Errorf("this operation requires the daemon to be running.\nStart it with: moltbunker start")
	}
	return nil
}

// Status retrieves node/network status.
func (sc *SmartClient) Status() (*StatusResponse, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.Status()
	}
	if sc.mode == ModeAPI {
		return sc.api.Status()
	}
	return nil, errNotConnected()
}

// Deploy deploys a container.
func (sc *SmartClient) Deploy(req *DeployRequest) (*DeployResponse, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.Deploy(req)
	}
	if sc.mode == ModeAPI {
		return sc.api.Deploy(req)
	}
	return nil, errNotConnected()
}

// List lists containers.
func (sc *SmartClient) List() ([]ContainerInfo, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.List()
	}
	if sc.mode == ModeAPI {
		return sc.api.List()
	}
	return nil, errNotConnected()
}

// GetLogs retrieves container logs.
func (sc *SmartClient) GetLogs(containerID string, follow bool, tail int) (string, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.GetLogs(containerID, follow, tail)
	}
	if sc.mode == ModeAPI {
		return sc.api.GetLogs(containerID, follow, tail)
	}
	return "", errNotConnected()
}

// Stop stops a container.
func (sc *SmartClient) Stop(containerID string) error {
	if sc.mode == ModeDaemon {
		return sc.daemon.Stop(containerID)
	}
	if sc.mode == ModeAPI {
		return sc.api.Stop(containerID)
	}
	return errNotConnected()
}

// Delete deletes a container.
func (sc *SmartClient) Delete(containerID string) error {
	if sc.mode == ModeDaemon {
		return sc.daemon.Delete(containerID)
	}
	if sc.mode == ModeAPI {
		return sc.api.Delete(containerID)
	}
	return errNotConnected()
}

// Health retrieves health information.
func (sc *SmartClient) Health(containerID string) (*HealthResponse, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.Health(containerID)
	}
	if sc.mode == ModeAPI {
		return sc.api.Health(containerID)
	}
	return nil, errNotConnected()
}

// Balance retrieves BUNKER token balance (API-only).
func (sc *SmartClient) Balance() (*BalanceResponse, error) {
	if sc.mode == ModeAPI {
		return sc.api.Balance()
	}
	// Daemon doesn't have a balance endpoint via socket — use API if available
	if sc.api != nil {
		return sc.api.Balance()
	}
	return nil, fmt.Errorf("balance requires API access.\nConfigure with: moltbunker config set api.endpoint <url>")
}

// GetContainerDetail retrieves detailed container info.
func (sc *SmartClient) GetContainerDetail(containerID string) (*ContainerDetail, error) {
	if sc.mode == ModeDaemon {
		return sc.daemon.GetContainerDetail(containerID)
	}
	if sc.mode == ModeAPI {
		return sc.api.GetContainerDetail(containerID)
	}
	return nil, errNotConnected()
}

// --- Daemon-only operations ---

// GetPeers retrieves the list of connected peers (daemon-only).
func (sc *SmartClient) GetPeers() ([]PeerInfo, error) {
	if err := sc.RequireDaemon(); err != nil {
		return nil, err
	}
	return sc.daemon.GetPeers()
}

// ConfigGet retrieves a configuration value (daemon-only).
func (sc *SmartClient) ConfigGet(key string) (interface{}, error) {
	if err := sc.RequireDaemon(); err != nil {
		return nil, err
	}
	return sc.daemon.ConfigGet(key)
}

// ConfigSet sets a configuration value (daemon-only).
func (sc *SmartClient) ConfigSet(key string, value interface{}) error {
	if err := sc.RequireDaemon(); err != nil {
		return err
	}
	return sc.daemon.ConfigSet(key, value)
}

// TorStart starts the Tor service (daemon-only).
func (sc *SmartClient) TorStart() (map[string]interface{}, error) {
	if err := sc.RequireDaemon(); err != nil {
		return nil, err
	}
	return sc.daemon.TorStart()
}

// TorStop stops the Tor service (daemon-only).
func (sc *SmartClient) TorStop() error {
	if err := sc.RequireDaemon(); err != nil {
		return err
	}
	return sc.daemon.TorStop()
}

// TorStatus retrieves Tor service status (daemon-only).
func (sc *SmartClient) TorStatus() (*TorStatusResponse, error) {
	if err := sc.RequireDaemon(); err != nil {
		return nil, err
	}
	return sc.daemon.TorStatus()
}

// TorRotate rotates the Tor circuit (daemon-only).
func (sc *SmartClient) TorRotate() error {
	if err := sc.RequireDaemon(); err != nil {
		return err
	}
	return sc.daemon.TorRotate()
}

// DaemonClient returns the underlying daemon client, if connected via daemon.
func (sc *SmartClient) DaemonClient() *DaemonClient {
	return sc.daemon
}

func errNotConnected() error {
	return fmt.Errorf("not connected — run 'moltbunker init' to set up")
}
