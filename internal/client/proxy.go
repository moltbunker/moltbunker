package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// ProxyStatusResponse contains proxy server status.
type ProxyStatusResponse struct {
	Running        bool   `json:"running"`
	SOCKS5Addr     string `json:"socks5_addr"`
	HTTPAddr       string `json:"http_addr"`
	UseTor         bool   `json:"use_tor"`
	ActiveSessions int    `json:"active_sessions"`
	MaxSessions    int    `json:"max_sessions"`
}

// ProxySessionInfo describes an active proxy session.
type ProxySessionInfo struct {
	ID        string    `json:"id"`
	Wallet    string    `json:"wallet"`
	Protocol  string    `json:"protocol"`
	Target    string    `json:"target"`
	BytesIn   int64     `json:"bytes_in"`
	BytesOut  int64     `json:"bytes_out"`
	StartedAt time.Time `json:"started_at"`
	UseTor    bool      `json:"use_tor"`
}

// ProxyUsageResponse contains bandwidth usage for a wallet.
type ProxyUsageResponse struct {
	Wallet       string `json:"wallet"`
	TotalIn      int64  `json:"total_bytes_in"`
	TotalOut     int64  `json:"total_bytes_out"`
	SessionCount int    `json:"session_count"`
}

// ProxyStatus returns the current proxy server status.
func (c *DaemonClient) ProxyStatus() (*ProxyStatusResponse, error) {
	resp, err := c.call("proxy_status", nil)
	if err != nil {
		return nil, err
	}
	var result ProxyStatusResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proxy status response: %w", err)
	}
	return &result, nil
}

// ProxySessions lists all active proxy sessions.
func (c *DaemonClient) ProxySessions() ([]ProxySessionInfo, error) {
	resp, err := c.call("proxy_sessions", nil)
	if err != nil {
		return nil, err
	}
	var result []ProxySessionInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proxy sessions response: %w", err)
	}
	return result, nil
}

// ProxyUsage returns bandwidth usage for the current wallet.
func (c *DaemonClient) ProxyUsage() (*ProxyUsageResponse, error) {
	resp, err := c.call("proxy_usage", nil)
	if err != nil {
		return nil, err
	}
	var result ProxyUsageResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse proxy usage response: %w", err)
	}
	return &result, nil
}

// ProxyCloseSession closes a proxy session by ID.
func (c *DaemonClient) ProxyCloseSession(sessionID string) error {
	_, err := c.call("proxy_close_session", map[string]string{"session_id": sessionID})
	return err
}
