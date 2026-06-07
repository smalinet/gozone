// Package pdns provides a client for the PowerDNS Authoritative Server REST API.
//
// The Client wraps HTTP calls to the /api/v1 endpoints exposed by PowerDNS,
// supporting zone and record management (CRUD), DNSSEC rectification, and
// slave notification.
package pdns

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/babykart/gozone/internal/config"
	"github.com/babykart/gozone/internal/logger"
	"github.com/babykart/gozone/internal/models"
)

// Client provides access to the PowerDNS REST API.
type Client struct {
	baseURL  string
	apiKey   string
	serverID string
	http     *http.Client
}

// NewClient creates a new PowerDNS API client from configuration.
//
// It normalizes the API URL to ensure it ends with "/api/v1" and configures
// an HTTP client with a 30-second request timeout.
func NewClient(cfg *config.PowerDNSConfig) *Client {
	baseURL := strings.TrimRight(cfg.APIURL, "/")
	if !strings.HasSuffix(baseURL, "/api/v1") {
		baseURL += "/api/v1"
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableKeepAlives:   false,
	}

	return &Client{
		baseURL:  baseURL,
		apiKey:   cfg.APIKey,
		serverID: cfg.ServerID,
		http: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}
}

func (c *Client) do(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	logger.Info("pdns request", "method", method, "path", path, "status", resp.StatusCode, "bytes", len(respBody))
	return respBody, resp.StatusCode, nil
}

// GetServers returns the list of PowerDNS servers.
func (c *Client) GetServers(ctx context.Context) ([]models.ServerInfo, error) {
	body, status, err := c.do(ctx, "GET", "/servers", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var servers []models.ServerInfo
	if err := json.Unmarshal(body, &servers); err != nil {
		return nil, fmt.Errorf("unmarshal servers: %w", err)
	}
	return servers, nil
}

// GetServer returns a single server's info.
func (c *Client) GetServer(ctx context.Context) (*models.ServerInfo, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var server models.ServerInfo
	if err := json.Unmarshal(body, &server); err != nil {
		return nil, fmt.Errorf("unmarshal server: %w", err)
	}
	return &server, nil
}

// GetStatistics returns global PowerDNS statistics.
func (c *Client) GetStatistics(ctx context.Context) ([]models.StatisticItem, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/statistics", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var stats []models.StatisticItem
	if err := json.Unmarshal(body, &stats); err != nil {
		return nil, fmt.Errorf("unmarshal statistics: %w", err)
	}
	return stats, nil
}

// ListZones returns all zones without their rrsets.
//
// ?rrsets=false prevents PowerDNS from including record sets in the response,
// keeping the payload small regardless of zone size.
func (c *Client) ListZones(ctx context.Context) ([]models.Zone, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/zones?rrsets=false", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var zones []models.Zone
	if err := json.Unmarshal(body, &zones); err != nil {
		return nil, fmt.Errorf("unmarshal zones: %w", err)
	}
	return zones, nil
}

// ListZonesWithInfo returns all zones in a single request.
//
// Record counts are not included; the former approach made one additional
// HTTP request per zone (N+1) which is replaced by a single ListZones call.
func (c *Client) ListZonesWithInfo(ctx context.Context) ([]models.ZoneWithInfo, error) {
	zones, err := c.ListZones(ctx)
	if err != nil {
		return nil, err
	}

	info := make([]models.ZoneWithInfo, len(zones))
	for i, z := range zones {
		info[i] = models.ZoneWithInfo{Zone: z}
	}
	return info, nil
}

// GetZone returns a specific zone.
func (c *Client) GetZone(ctx context.Context, zoneID string) (*models.Zone, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/zones/"+zoneID, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var zone models.Zone
	if err := json.Unmarshal(body, &zone); err != nil {
		return nil, fmt.Errorf("unmarshal zone: %w", err)
	}
	return &zone, nil
}

// CreateZone creates a new zone.
func (c *Client) CreateZone(ctx context.Context, req models.ZoneCreateRequest) (*models.Zone, error) {
	if req.Kind == "" {
		req.Kind = "Native"
	}
	if req.Nameservers == nil {
		req.Nameservers = []string{}
	}

	body, status, err := c.do(ctx, "POST", "/servers/"+c.serverID+"/zones", req)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var zone models.Zone
	if err := json.Unmarshal(body, &zone); err != nil {
		return nil, fmt.Errorf("unmarshal zone: %w", err)
	}
	return &zone, nil
}

// DeleteZone deletes a zone.
func (c *Client) DeleteZone(ctx context.Context, zoneID string) error {
	_, status, err := c.do(ctx, "DELETE", "/servers/"+c.serverID+"/zones/"+zoneID, nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// ListRecords returns all records (RRSets) for a zone.
func (c *Client) ListRecords(ctx context.Context, zoneID string) ([]models.RRSet, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/zones/"+zoneID, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var full struct {
		RRSets []models.RRSet `json:"rrsets"`
	}
	if err := json.Unmarshal(body, &full); err != nil {
		return nil, fmt.Errorf("unmarshal rrsets: %w", err)
	}
	return full.RRSets, nil
}

// CreateRecord creates a new RRSet in a zone.
func (c *Client) CreateRecord(ctx context.Context, zoneID string, rrset models.RRSet) error {
	rrset.ChangeType = "REPLACE"
	return c.patchZone(ctx, zoneID, []models.RRSet{rrset})
}

// UpdateRecord updates an existing RRSet.
func (c *Client) UpdateRecord(ctx context.Context, zoneID string, rrset models.RRSet) error {
	rrset.ChangeType = "REPLACE"
	return c.patchZone(ctx, zoneID, []models.RRSet{rrset})
}

// DeleteRecord deletes an RRSet from a zone.
func (c *Client) DeleteRecord(ctx context.Context, zoneID string, name, recordType string) error {
	rrset := models.RRSet{
		Name:       name,
		Type:       recordType,
		ChangeType: "DELETE",
	}
	return c.patchZone(ctx, zoneID, []models.RRSet{rrset})
}

// CreateRecords creates multiple RRSets in a zone in a single PATCH call.
func (c *Client) CreateRecords(ctx context.Context, zoneID string, rrsets []models.RRSet) error {
	if len(rrsets) == 0 {
		return nil
	}
	for i := range rrsets {
		rrsets[i].ChangeType = "REPLACE"
	}
	return c.patchZone(ctx, zoneID, rrsets)
}

func (c *Client) patchZone(ctx context.Context, zoneID string, rrsets []models.RRSet) error {
	payload := map[string]interface{}{
		"rrsets": rrsets,
	}

	_, status, err := c.do(ctx, "PATCH", "/servers/"+c.serverID+"/zones/"+zoneID, payload)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// RectifyZone triggers DNSSEC rectification for a zone.
func (c *Client) RectifyZone(ctx context.Context, zoneID string) error {
	_, status, err := c.do(ctx, "PUT", "/servers/"+c.serverID+"/zones/"+zoneID+"/rectify", nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// NotifySlaves sends NOTIFY to slave servers for a zone.
func (c *Client) NotifySlaves(ctx context.Context, zoneID string) error {
	_, status, err := c.do(ctx, "PUT", "/servers/"+c.serverID+"/zones/"+zoneID+"/notify", nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// GetMetadata returns all zone metadata entries.
func (c *Client) GetMetadata(ctx context.Context, zoneID string) ([]models.Metadata, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/zones/"+zoneID+"/metadata", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var metadata []models.Metadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return metadata, nil
}

// SetMetadata creates or replaces a zone metadata entry.
// Uses PUT with the kind in the URL path for broader compatibility across
// PowerDNS versions.
func (c *Client) SetMetadata(ctx context.Context, zoneID string, meta models.Metadata) error {
	if meta.Metadata == nil {
		meta.Metadata = []string{}
	}
	payload := map[string][]string{"metadata": meta.Metadata}
	_, status, err := c.do(ctx, "PUT", "/servers/"+c.serverID+"/zones/"+zoneID+"/metadata/"+meta.Kind, payload)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// DeleteMetadata removes a zone metadata entry by kind.
func (c *Client) DeleteMetadata(ctx context.Context, zoneID string, kind string) error {
	_, status, err := c.do(ctx, "DELETE", "/servers/"+c.serverID+"/zones/"+zoneID+"/metadata/"+kind, nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// ServerID returns the configured server ID.
func (c *Client) ServerID() string {
	return c.serverID
}

// ListTSIGKeys returns all TSIG keys for the server.
func (c *Client) ListTSIGKeys(ctx context.Context) ([]models.TSIGKey, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/tsigkeys", nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var keys []models.TSIGKey
	if err := json.Unmarshal(body, &keys); err != nil {
		return nil, fmt.Errorf("unmarshal tsigkeys: %w", err)
	}
	return keys, nil
}

// GetTSIGKey returns a single TSIG key.
func (c *Client) GetTSIGKey(ctx context.Context, id string) (*models.TSIGKey, error) {
	body, status, err := c.do(ctx, "GET", "/servers/"+c.serverID+"/tsigkeys/"+id, nil)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var key models.TSIGKey
	if err := json.Unmarshal(body, &key); err != nil {
		return nil, fmt.Errorf("unmarshal tsigkey: %w", err)
	}
	return &key, nil
}

// CreateTSIGKey creates a new TSIG key.
func (c *Client) CreateTSIGKey(ctx context.Context, key models.TSIGKey) (*models.TSIGKey, error) {
	body, status, err := c.do(ctx, "POST", "/servers/"+c.serverID+"/tsigkeys", key)
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s", status, string(body))
	}

	var created models.TSIGKey
	if err := json.Unmarshal(body, &created); err != nil {
		return nil, fmt.Errorf("unmarshal tsigkey: %w", err)
	}
	return &created, nil
}

// UpdateTSIGKey updates an existing TSIG key.
func (c *Client) UpdateTSIGKey(ctx context.Context, id string, key models.TSIGKey) error {
	_, status, err := c.do(ctx, "PUT", "/servers/"+c.serverID+"/tsigkeys/"+id, key)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}

// DeleteTSIGKey deletes a TSIG key.
func (c *Client) DeleteTSIGKey(ctx context.Context, id string) error {
	_, status, err := c.do(ctx, "DELETE", "/servers/"+c.serverID+"/tsigkeys/"+id, nil)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return fmt.Errorf("unexpected status %d", status)
	}
	return nil
}
