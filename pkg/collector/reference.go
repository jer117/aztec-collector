package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aztec-collector/pkg/jsonrpc"
)

// ReferenceService queries an external source for the reference block height
type ReferenceService struct {
	config     *RefServiceConfig
	httpClient *http.Client
	rpcClient  *jsonrpc.Client
}

// NewReferenceService creates a new reference service
func NewReferenceService(config *RefServiceConfig) (*ReferenceService, error) {
	rs := &ReferenceService{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}

	// If using external RPC, create an RPC client
	if config.UseExternalRPC && config.ExternalRPCURL != "" {
		client, err := jsonrpc.New(
			jsonrpc.WithURL(config.ExternalRPCURL),
			jsonrpc.WithHTTPClient(rs.httpClient),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create external RPC client: %w", err)
		}
		rs.rpcClient = client
	}

	return rs, nil
}

// GetReferenceHeight gets the reference block height from the configured source
func (rs *ReferenceService) GetReferenceHeight(ctx context.Context) (int64, error) {
	if rs.config.UseExternalRPC {
		return rs.getReferenceHeightFromRPC(ctx)
	}
	return rs.getReferenceHeightFromService(ctx)
}

// getReferenceHeightFromRPC queries another Aztec node for the block height
func (rs *ReferenceService) getReferenceHeightFromRPC(ctx context.Context) (int64, error) {
	if rs.rpcClient == nil {
		return 0, fmt.Errorf("external RPC client not configured")
	}

	// Query the external node's L2 tips
	var tips struct {
		Latest struct {
			Number int64 `json:"number"`
		} `json:"latest"`
	}

	if err := rs.rpcClient.Call("node_getL2Tips", []interface{}{}, &tips); err != nil {
		return 0, fmt.Errorf("failed to get L2 tips from reference node: %w", err)
	}

	return tips.Latest.Number, nil
}

// getReferenceHeightFromService queries an HTTP reference service
// This supports a simple REST API that returns block height
func (rs *ReferenceService) getReferenceHeightFromService(ctx context.Context) (int64, error) {
	if rs.config.URL == "" {
		return 0, fmt.Errorf("reference service URL not configured")
	}

	// Build the request URL
	url := fmt.Sprintf("%s/v1/protocols/%s/networks/%s/height",
		rs.config.URL, rs.config.Protocol, rs.config.Network)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := rs.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to query reference service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("reference service returned status %d", resp.StatusCode)
	}

	// Try to parse the response
	var result struct {
		Height      int64 `json:"height"`
		BlockHeight int64 `json:"block_height"`
		Number      int64 `json:"number"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to parse reference service response: %w", err)
	}

	// Return the first non-zero height
	if result.Height > 0 {
		return result.Height, nil
	}
	if result.BlockHeight > 0 {
		return result.BlockHeight, nil
	}
	if result.Number > 0 {
		return result.Number, nil
	}

	return 0, fmt.Errorf("no valid height in reference service response")
}

// RefServiceInfo contains information about the reference service query
type RefServiceInfo struct {
	Enabled         bool      `json:"enabled"`
	Source          string    `json:"source,omitempty"`
	ReferenceHeight int64     `json:"referenceHeight,omitempty"`
	QueryTime       time.Time `json:"queryTime,omitempty"`
	Error           string    `json:"error,omitempty"`
}

// GetInfo returns information about the reference service configuration
func (rs *ReferenceService) GetInfo() *RefServiceInfo {
	info := &RefServiceInfo{
		Enabled: rs.config.Enabled,
	}

	if rs.config.UseExternalRPC {
		info.Source = rs.config.ExternalRPCURL
	} else {
		info.Source = rs.config.URL
	}

	return info
}

