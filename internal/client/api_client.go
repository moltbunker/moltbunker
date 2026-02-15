package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient communicates with the moltbunker HTTP API using wallet-signed requests.
// This enables requesters to use the CLI without running a daemon.
type APIClient struct {
	baseURL    string
	signer     *WalletSigner
	httpClient *http.Client
}

// NewAPIClient creates a new HTTP API client with wallet signing.
func NewAPIClient(baseURL string, signer *WalletSigner) *APIClient {
	return &APIClient{
		baseURL: baseURL,
		signer:  signer,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// do performs an authenticated HTTP request and decodes the JSON response into out.
func (c *APIClient) do(method, path string, body interface{}, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseURL+path, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Sign request with wallet
	if c.signer != nil {
		addr, sig, msg, err := c.signer.SignAuth()
		if err != nil {
			return fmt.Errorf("failed to sign request: %w", err)
		}
		req.Header.Set("X-Wallet-Address", addr)
		req.Header.Set("X-Wallet-Signature", sig)
		req.Header.Set("X-Wallet-Message", msg)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		// Try to extract error message from JSON response
		var errResp struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &errResp) == nil {
			msg := errResp.Error
			if msg == "" {
				msg = errResp.Message
			}
			if msg != "" {
				return fmt.Errorf("API error (%d): %s", resp.StatusCode, msg)
			}
		}
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

// Status retrieves node/network status.
func (c *APIClient) Status() (*StatusResponse, error) {
	var status StatusResponse
	if err := c.do("GET", "/v1/status", nil, &status); err != nil {
		return nil, err
	}
	return &status, nil
}

// Deploy deploys a container via the HTTP API.
func (c *APIClient) Deploy(req *DeployRequest) (*DeployResponse, error) {
	var resp DeployResponse
	if err := c.do("POST", "/v1/deploy", req, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// List lists containers via the HTTP API.
func (c *APIClient) List() ([]ContainerInfo, error) {
	var containers []ContainerInfo
	if err := c.do("GET", "/v1/containers", nil, &containers); err != nil {
		return nil, err
	}
	return containers, nil
}

// GetLogs retrieves container logs via the HTTP API.
func (c *APIClient) GetLogs(containerID string, follow bool, tail int) (string, error) {
	path := fmt.Sprintf("/v1/containers/%s/logs?tail=%d", containerID, tail)
	var result struct {
		Logs string `json:"logs"`
	}
	if err := c.do("GET", path, nil, &result); err != nil {
		return "", err
	}
	return result.Logs, nil
}

// Stop stops a container via the HTTP API.
func (c *APIClient) Stop(containerID string) error {
	return c.do("POST", fmt.Sprintf("/v1/containers/%s/stop", containerID), nil, nil)
}

// Delete deletes a container via the HTTP API.
func (c *APIClient) Delete(containerID string) error {
	return c.do("DELETE", fmt.Sprintf("/v1/containers/%s", containerID), nil, nil)
}

// Balance retrieves BUNKER token balance.
func (c *APIClient) Balance() (*BalanceResponse, error) {
	var resp BalanceResponse
	if err := c.do("GET", "/v1/balance", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// Health retrieves system health.
func (c *APIClient) Health(_ string) (*HealthResponse, error) {
	var resp HealthResponse
	if err := c.do("GET", "/v1/health", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetContainerDetail retrieves detailed container info.
func (c *APIClient) GetContainerDetail(containerID string) (*ContainerDetail, error) {
	var detail ContainerDetail
	if err := c.do("GET", fmt.Sprintf("/v1/containers/%s", containerID), nil, &detail); err != nil {
		return nil, err
	}
	return &detail, nil
}

// BalanceResponse contains token balance information.
type BalanceResponse struct {
	Address   string `json:"address"`
	Available string `json:"available"`
	Reserved  string `json:"reserved"`
	Staked    string `json:"staked"`
	Total     string `json:"total"`
}
