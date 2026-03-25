package ingress

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/moltbunker/moltbunker/internal/logging"
)

// DNSSync manages Cloudflare DNS records for subdomain registration.
// It creates/deletes A records via the Cloudflare API v4 when subdomains
// are registered or released, eliminating the need for a manual wildcard record.
type DNSSync struct {
	apiToken   string
	zoneID     string
	ingressIP  string
	domain     string
	httpClient *http.Client
}

// NewDNSSync creates a new Cloudflare DNS sync manager.
func NewDNSSync(apiToken, zoneID, ingressIP, domain string) *DNSSync {
	return &DNSSync{
		apiToken:  apiToken,
		zoneID:    zoneID,
		ingressIP: ingressIP,
		domain:    domain,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// cfDNSRecord represents a Cloudflare DNS record.
type cfDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

// cfAPIResponse wraps a Cloudflare API response.
type cfAPIResponse struct {
	Success bool            `json:"success"`
	Errors  []cfAPIError    `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

// cfAPIError represents a Cloudflare API error.
type cfAPIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// CreateRecord creates an A record for the subdomain pointing to the ingress IP.
// Idempotent: if a record already exists with the correct IP, it's a no-op.
func (d *DNSSync) CreateRecord(ctx context.Context, subdomain string) error {
	fqdn := subdomain + "." + d.domain

	// Check if record already exists
	existing, err := d.listRecords(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("list DNS records: %w", err)
	}
	for _, rec := range existing {
		if rec.Type == "A" && rec.Content == d.ingressIP {
			logging.Debug("DNS record already exists, skipping",
				"subdomain", subdomain,
				"fqdn", fqdn,
				logging.Component("dns-sync"))
			return nil
		}
	}

	// Create new A record
	body := map[string]interface{}{
		"type":    "A",
		"name":    fqdn,
		"content": d.ingressIP,
		"ttl":     300,
		"proxied": false,
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal record: %w", err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records", d.zoneID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyJSON))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+d.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("cloudflare API call: %w", err)
	}
	defer resp.Body.Close()

	var apiResp cfAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Success {
		return fmt.Errorf("cloudflare API error: %v", apiResp.Errors)
	}

	logging.Info("DNS record created",
		"subdomain", subdomain,
		"fqdn", fqdn,
		"ip", d.ingressIP,
		logging.Component("dns-sync"))
	return nil
}

// DeleteRecord removes A records for the subdomain.
func (d *DNSSync) DeleteRecord(ctx context.Context, subdomain string) error {
	fqdn := subdomain + "." + d.domain

	existing, err := d.listRecords(ctx, fqdn)
	if err != nil {
		return fmt.Errorf("list DNS records: %w", err)
	}

	for _, rec := range existing {
		if rec.Type != "A" {
			continue
		}
		if err := d.deleteRecord(ctx, rec.ID); err != nil {
			return fmt.Errorf("delete record %s: %w", rec.ID, err)
		}
	}

	logging.Info("DNS record deleted",
		"subdomain", subdomain,
		"fqdn", fqdn,
		logging.Component("dns-sync"))
	return nil
}

// listRecords fetches DNS records matching the given FQDN.
func (d *DNSSync) listRecords(ctx context.Context, fqdn string) ([]cfDNSRecord, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records?name=%s&type=A", d.zoneID, fqdn)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+d.apiToken)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp cfAPIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Success {
		return nil, fmt.Errorf("cloudflare API error: %v", apiResp.Errors)
	}

	var records []cfDNSRecord
	if err := json.Unmarshal(apiResp.Result, &records); err != nil {
		return nil, fmt.Errorf("decode records: %w", err)
	}
	return records, nil
}

// deleteRecord removes a single DNS record by ID.
func (d *DNSSync) deleteRecord(ctx context.Context, recordID string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/dns_records/%s", d.zoneID, recordID)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+d.apiToken)

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var apiResp cfAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	if !apiResp.Success {
		return fmt.Errorf("cloudflare API error: %v", apiResp.Errors)
	}
	return nil
}
