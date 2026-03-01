package client

import (
	"encoding/json"
	"fmt"
	"time"
)

// SubdomainRegisterRequest contains subdomain registration parameters.
type SubdomainRegisterRequest struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
}

// SubdomainRegisterResponse contains subdomain registration result.
type SubdomainRegisterResponse struct {
	Name         string `json:"name"`
	DeploymentID string `json:"deployment_id"`
	URL          string `json:"url"`
	TxHash       string `json:"tx_hash,omitempty"`
}

// SubdomainInfo contains subdomain details.
type SubdomainInfo struct {
	Name         string    `json:"name"`
	DeploymentID string    `json:"deployment_id"`
	Owner        string    `json:"owner"`
	URL          string    `json:"url"`
	RegisteredAt time.Time `json:"registered_at"`
}

// SubdomainRegister registers a vanity subdomain.
func (c *DaemonClient) SubdomainRegister(name, deploymentID string) (*SubdomainRegisterResponse, error) {
	resp, err := c.call("subdomain_register", SubdomainRegisterRequest{
		Name:         name,
		DeploymentID: deploymentID,
	})
	if err != nil {
		return nil, err
	}
	var result SubdomainRegisterResponse
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse register response: %w", err)
	}
	return &result, nil
}

// SubdomainRelease releases a vanity subdomain.
func (c *DaemonClient) SubdomainRelease(name string) error {
	_, err := c.call("subdomain_release", map[string]string{"name": name})
	return err
}

// SubdomainList lists subdomains owned by this node's wallet.
func (c *DaemonClient) SubdomainList() ([]SubdomainInfo, error) {
	resp, err := c.call("subdomain_list", nil)
	if err != nil {
		return nil, err
	}
	var result []SubdomainInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse subdomain list: %w", err)
	}
	return result, nil
}

// SubdomainResolve resolves a subdomain name.
func (c *DaemonClient) SubdomainResolve(name string) (*SubdomainInfo, error) {
	resp, err := c.call("subdomain_resolve", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var result SubdomainInfo
	if err := json.Unmarshal(resp.Result, &result); err != nil {
		return nil, fmt.Errorf("failed to parse subdomain info: %w", err)
	}
	return &result, nil
}

// SubdomainTransfer transfers subdomain ownership.
func (c *DaemonClient) SubdomainTransfer(name, newOwner string) error {
	_, err := c.call("subdomain_transfer", map[string]string{
		"name":      name,
		"new_owner": newOwner,
	})
	return err
}

// SubdomainUpdate updates the deployment ID for a subdomain.
func (c *DaemonClient) SubdomainUpdate(name, deploymentID string) error {
	_, err := c.call("subdomain_update", map[string]string{
		"name":          name,
		"deployment_id": deploymentID,
	})
	return err
}

// SubdomainRenew extends a subdomain's expiration by 365 days.
func (c *DaemonClient) SubdomainRenew(name string) error {
	_, err := c.call("subdomain_renew", map[string]string{"name": name})
	return err
}

// SubdomainReserve reserves a subdomain name for 48 hours.
func (c *DaemonClient) SubdomainReserve(name string) error {
	_, err := c.call("subdomain_reserve", map[string]string{"name": name})
	return err
}

// SubdomainClaim finalizes a reserved subdomain with a deployment ID.
func (c *DaemonClient) SubdomainClaim(name, deploymentID string) error {
	_, err := c.call("subdomain_claim", map[string]string{
		"name":          name,
		"deployment_id": deploymentID,
	})
	return err
}

// SubdomainCancel cancels a pending subdomain reservation.
func (c *DaemonClient) SubdomainCancel(name string) error {
	_, err := c.call("subdomain_cancel", map[string]string{"name": name})
	return err
}

// SubdomainSetMetadata sets description and avatar URL for a subdomain.
func (c *DaemonClient) SubdomainSetMetadata(name, description, avatarURL string) error {
	_, err := c.call("subdomain_metadata", map[string]interface{}{
		"name":        name,
		"description": description,
		"avatar_url":  avatarURL,
	})
	return err
}

// SubdomainSetPrimary sets a subdomain as the primary name for reverse resolution.
func (c *DaemonClient) SubdomainSetPrimary(name string) error {
	_, err := c.call("subdomain_primary", map[string]string{"name": name})
	return err
}

// SubdomainReclaim reclaims a squatted subdomain name.
func (c *DaemonClient) SubdomainReclaim(name string) error {
	_, err := c.call("subdomain_reclaim", map[string]string{"name": name})
	return err
}
