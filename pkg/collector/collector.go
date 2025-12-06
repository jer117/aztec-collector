package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aztec-collector/pkg/alerting"
	"github.com/aztec-collector/pkg/aztec"
	"github.com/aztec-collector/pkg/jsonrpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Collector is the main Aztec node metrics collector
type Collector struct {
	config     *Config
	rpc        *AztecRPC
	metrics    *Metrics
	logger     *log.Logger
	refService *ReferenceService
	alerter    *alerting.Alerter

	// State
	mu          sync.RWMutex
	latestState *CollectorState

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Collector with the given options
func New(options ...func(*Collector)) (*Collector, error) {
	c := &Collector{
		logger: log.New(os.Stdout, "[aztec-collector] ", log.LstdFlags),
	}

	for _, opt := range options {
		opt(c)
	}

	if c.config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if c.rpc == nil {
		return nil, fmt.Errorf("RPC client is required")
	}

	if c.metrics == nil {
		c.metrics = NewMetrics(prometheus.DefaultRegisterer)
	}

	// Initialize reference service if enabled
	if c.config.RefService.Enabled {
		refService, err := NewReferenceService(&c.config.RefService)
		if err != nil {
			return nil, fmt.Errorf("failed to create reference service: %w", err)
		}
		c.refService = refService
		c.logger.Printf("Reference service enabled: %s", refService.GetInfo().Source)
	}

	// Initialize alerter if enabled
	if c.config.Alerting != nil && c.config.Alerting.Enabled {
		c.alerter = alerting.NewAlerter(c.config.Alerting, c.logger)
		c.logger.Printf("Alerting enabled with cooldown: %s", c.config.Alerting.Cooldown)
	}

	return c, nil
}

// WithConfig sets the configuration
func WithConfig(config *Config) func(*Collector) {
	return func(c *Collector) {
		c.config = config
	}
}

// WithRPCClient sets the RPC client
func WithRPCClient(client *jsonrpc.Client) func(*Collector) {
	return func(c *Collector) {
		c.rpc = NewAztecRPC(client)
	}
}

// WithMetrics sets the metrics
func WithMetrics(metrics *Metrics) func(*Collector) {
	return func(c *Collector) {
		c.metrics = metrics
	}
}

// WithLogger sets the logger
func WithLogger(logger *log.Logger) func(*Collector) {
	return func(c *Collector) {
		c.logger = logger
	}
}

// Run starts the collector
func (c *Collector) Run(ctx context.Context) error {
	c.ctx, c.cancel = context.WithCancel(ctx)

	// Start API server if enabled
	if c.config.API.Enabled {
		go c.startAPIServer()
	}

	// Start collection loop
	ticker := time.NewTicker(c.config.Interval)
	defer ticker.Stop()

	// Collect immediately on start
	c.collect()

	for {
		select {
		case <-c.ctx.Done():
			return c.ctx.Err()
		case <-ticker.C:
			c.collect()
		}
	}
}

// Stop stops the collector
func (c *Collector) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
}

// collect performs a single collection cycle
func (c *Collector) collect() {
	startTime := time.Now()
	state := NewCollectorState()

	c.logger.Println("Starting collection cycle...")

	// Collect L2 tips (most important)
	if err := c.collectL2Tips(state); err != nil {
		c.logger.Printf("Error collecting L2 tips: %v", err)
		state.AddMessage(HealthError, fmt.Sprintf("Failed to get L2 tips: %v", err))
		c.metrics.RecordCollectionError("getL2Tips")
	}

	// Collect node info
	if err := c.collectNodeInfo(state); err != nil {
		c.logger.Printf("Error collecting node info: %v", err)
		state.AddMessage(HealthWarn, fmt.Sprintf("Failed to get node info: %v", err))
		c.metrics.RecordCollectionError("getNodeInfo")
	}

	// Collect sync status
	if err := c.collectSyncStatus(state); err != nil {
		c.logger.Printf("Error collecting sync status: %v", err)
		state.AddMessage(HealthWarn, fmt.Sprintf("Failed to get sync status: %v", err))
		c.metrics.RecordCollectionError("getSyncStatus")
	}

	// Check if node is ready
	if err := c.collectReadyStatus(state); err != nil {
		c.logger.Printf("Error checking ready status: %v", err)
		state.AddMessage(HealthWarn, fmt.Sprintf("Failed to check ready status: %v", err))
		c.metrics.RecordCollectionError("isReady")
	}

	// Collect pending tx count
	if err := c.collectPendingTxCount(state); err != nil {
		c.logger.Printf("Error collecting pending tx count: %v", err)
		// Don't set error - this is not critical
		c.metrics.RecordCollectionError("getPendingTxCount")
	}

	// Collect base fees
	if err := c.collectBaseFees(state); err != nil {
		c.logger.Printf("Error collecting base fees: %v", err)
		// Don't set error - this is not critical
		c.metrics.RecordCollectionError("getBaseFees")
	}

	// Collect reference height if enabled
	if err := c.collectReferenceHeight(state); err != nil {
		c.logger.Printf("Error collecting reference height: %v", err)
		// Don't set error - this is not critical
		c.metrics.RecordCollectionError("getReferenceHeight")
	}

	// Collect validator stats if enabled
	if c.config.Validator.Enabled {
		if err := c.collectValidatorStats(state); err != nil {
			c.logger.Printf("Error collecting validator stats: %v", err)
			c.metrics.RecordCollectionError("getValidatorsStats")
		}
	}

	// Evaluate health
	c.evaluateHealth(state)

	// Update state timestamp
	state.CollectedAt = time.Now()
	state.Aztec.CollectionTime = state.CollectedAt

	// Record collection duration
	duration := time.Since(startTime)
	c.metrics.CollectionDuration.WithLabelValues().Observe(duration.Seconds())

	// Update metrics from state
	c.metrics.UpdateFromState(state)

	// Store latest state
	c.mu.Lock()
	c.latestState = state
	c.mu.Unlock()

	// Write to JSON log if configured
	if c.config.JSONLog.Path != "" {
		c.writeJSONLog(state)
	}

	c.logger.Printf("Collection cycle completed in %v", duration)
}

// collectL2Tips collects L2 chain tips
func (c *Collector) collectL2Tips(state *CollectorState) error {
	tips, err := c.rpc.GetL2Tips()
	if err != nil {
		return err
	}

	state.Aztec.L2Tips = tips
	state.UpdateFromL2Tips(tips)
	return nil
}

// collectNodeInfo collects node information
func (c *Collector) collectNodeInfo(state *CollectorState) error {
	info, err := c.rpc.GetNodeInfo()
	if err != nil {
		return err
	}

	state.Aztec.NodeInfo = info
	state.UpdateFromNodeInfo(info)
	return nil
}

// collectSyncStatus collects world state sync status
func (c *Collector) collectSyncStatus(state *CollectorState) error {
	status, err := c.rpc.GetWorldStateSyncStatus()
	if err != nil {
		return err
	}

	state.Aztec.SyncStatus = status
	state.UpdateFromSyncStatus(status)
	return nil
}

// collectReadyStatus checks if the node is ready
func (c *Collector) collectReadyStatus(state *CollectorState) error {
	ready, err := c.rpc.IsReady()
	if err != nil {
		return err
	}

	state.Node.IsReady = ready
	
	// If node is ready, it's considered synced
	if ready {
		state.Node.IsSyncing = false
	}
	
	return nil
}

// collectPendingTxCount collects pending transaction count
func (c *Collector) collectPendingTxCount(state *CollectorState) error {
	count, err := c.rpc.GetPendingTxCount()
	if err != nil {
		return err
	}

	state.Aztec.PendingTxCount = count
	return nil
}

// collectBaseFees collects current base fees
func (c *Collector) collectBaseFees(state *CollectorState) error {
	fees, err := c.rpc.GetCurrentBaseFees()
	if err != nil {
		return err
	}

	state.Aztec.CurrentBaseFees = fees
	return nil
}

// collectReferenceHeight collects reference block height from external source
func (c *Collector) collectReferenceHeight(state *CollectorState) error {
	if c.refService == nil || !c.config.RefService.Enabled {
		return nil
	}

	ctx, cancel := context.WithTimeout(c.ctx, c.config.RefService.Timeout)
	defer cancel()

	height, err := c.refService.GetReferenceHeight(ctx)
	if err != nil {
		state.RefService = &RefServiceInfo{
			Enabled:   true,
			Source:    c.refService.GetInfo().Source,
			QueryTime: time.Now(),
			Error:     err.Error(),
		}
		return err
	}

	state.Chain.ReferenceHeight = height
	state.RefService = &RefServiceInfo{
		Enabled:         true,
		Source:          c.refService.GetInfo().Source,
		ReferenceHeight: height,
		QueryTime:       time.Now(),
	}

	c.logger.Printf("Reference height: %d (local: %d)", height, state.Chain.HeadHeight)
	return nil
}

// collectValidatorStats collects validator statistics and checks for missed attestations/proposals
func (c *Collector) collectValidatorStats(state *CollectorState) error {
	stats, err := c.rpc.GetValidatorsStats()
	if err != nil {
		return err
	}

	state.Aztec.ValidatorsStats = stats

	if stats.Stats == nil {
		return nil
	}

	// If no specific addresses configured, monitor ALL validators
	if len(c.config.Validator.Addresses) == 0 {
		for _, validatorStats := range stats.Stats {
			c.checkValidatorAlerts(validatorStats, stats.LastProcessedSlot)
		}
	} else {
		// Monitor only the specified addresses
		for _, configAddr := range c.config.Validator.Addresses {
			// Normalize the address for lookup (lowercase)
			addr := strings.ToLower(strings.TrimSpace(configAddr))
			
			if validatorStats, ok := stats.Stats[addr]; ok {
				c.checkValidatorAlerts(validatorStats, stats.LastProcessedSlot)
			} else {
				c.logger.Printf("Validator %s not found in stats", configAddr)
			}
		}
	}

	return nil
}

// checkValidatorAlerts checks for missed attestations/proposals and sends alerts
func (c *Collector) checkValidatorAlerts(stats *aztec.ValidatorStats, lastSlot string) {
	// Check for missed attestations
	if c.config.Validator.AlertOnMissedAttestation && stats.MissedAttestations != nil {
		if stats.MissedAttestations.CurrentStreak > 0 {
			c.sendAlert(alerting.Alert{
				Type:             alerting.AlertMissedAttestation,
				Level:            alerting.AlertLevelError,
				Title:            "Missed Attestation",
				Message:          fmt.Sprintf("Validator %s has missed %d attestation(s) in a row!", stats.Address, stats.MissedAttestations.CurrentStreak),
				ValidatorAddress: stats.Address,
				Slot:             lastSlot,
				MissedStreak:     stats.MissedAttestations.CurrentStreak,
			})
		} else if c.alerter != nil && c.alerter.IsAlertActive(alerting.AlertMissedAttestation) {
			c.sendRecoveryAlert(alerting.AlertMissedAttestation, fmt.Sprintf("Validator %s is now attesting normally", stats.Address))
		}
	}

	// Check for missed proposals
	if c.config.Validator.AlertOnMissedProposal && stats.MissedProposals != nil {
		if stats.MissedProposals.CurrentStreak > 0 {
			c.sendAlert(alerting.Alert{
				Type:             alerting.AlertMissedProposal,
				Level:            alerting.AlertLevelCritical,
				Title:            "Missed Block Proposal",
				Message:          fmt.Sprintf("Validator %s has missed %d block proposal(s) in a row!", stats.Address, stats.MissedProposals.CurrentStreak),
				ValidatorAddress: stats.Address,
				Slot:             lastSlot,
				MissedStreak:     stats.MissedProposals.CurrentStreak,
			})
		} else if c.alerter != nil && c.alerter.IsAlertActive(alerting.AlertMissedProposal) {
			c.sendRecoveryAlert(alerting.AlertMissedProposal, fmt.Sprintf("Validator %s is now proposing normally", stats.Address))
		}
	}
}

// evaluateHealth evaluates the health of the node based on collected state
func (c *Collector) evaluateHealth(state *CollectorState) {
	// Check if node is ready
	if !state.Node.IsReady {
		state.SetErrorCondition("node_not_ready")
		c.sendAlert(alerting.Alert{
			Type:    alerting.AlertNodeNotReady,
			Level:   alerting.AlertLevelError,
			Title:   "Node Not Ready",
			Message: "The Aztec node is not ready to serve requests.",
		})
	} else if c.alerter != nil && c.alerter.IsAlertActive(alerting.AlertNodeNotReady) {
		c.sendRecoveryAlert(alerting.AlertNodeNotReady, "Node is now ready")
	}

	// Check if node is syncing
	if state.Node.IsSyncing {
		state.SetWarnCondition("syncing")
		c.sendAlert(alerting.Alert{
			Type:        alerting.AlertNodeNotSynced,
			Level:       alerting.AlertLevelWarning,
			Title:       "Node Syncing",
			Message:     "The Aztec node is currently syncing.",
			BlockHeight: state.Chain.HeadHeight,
		})
	} else if c.alerter != nil && c.alerter.IsAlertActive(alerting.AlertNodeNotSynced) {
		c.sendRecoveryAlert(alerting.AlertNodeNotSynced, "Node sync complete")
	}

	// Check blocks behind using reference height
	if behind, ok := state.GetBlocksBehind(); ok {
		if behind > c.config.Health.MaxBlocksBehind {
			state.SetErrorCondition("blocks_behind")
			state.AddMessage(HealthError, fmt.Sprintf("Node is %d blocks behind reference", behind))
			c.sendAlert(alerting.Alert{
				Type:         alerting.AlertBlocksBehind,
				Level:        alerting.AlertLevelError,
				Title:        "Node Falling Behind",
				Message:      fmt.Sprintf("Node is %d blocks behind the reference (threshold: %d).", behind, c.config.Health.MaxBlocksBehind),
				BlockHeight:  state.Chain.HeadHeight,
				BlocksBehind: behind,
			})
		} else if behind > 0 {
			state.SetWarnCondition("blocks_behind")
		} else if c.alerter != nil && c.alerter.IsAlertActive(alerting.AlertBlocksBehind) {
			c.sendRecoveryAlert(alerting.AlertBlocksBehind, "Node is now in sync")
		}
	}

	// Check blocks ahead using reference height
	if ahead, ok := state.GetBlocksAhead(); ok {
		if ahead > c.config.Health.MaxBlocksAhead {
			state.SetWarnCondition("blocks_ahead")
			state.AddMessage(HealthWarn, fmt.Sprintf("Node is %d blocks ahead of reference", ahead))
			c.sendAlert(alerting.Alert{
				Type:        alerting.AlertBlocksAhead,
				Level:       alerting.AlertLevelWarning,
				Title:       "Node Ahead of Reference",
				Message:     fmt.Sprintf("Node is %d blocks ahead of the reference (threshold: %d).", ahead, c.config.Health.MaxBlocksAhead),
				BlockHeight: state.Chain.HeadHeight,
			})
		}
	}

	// Check proven vs latest
	if state.Chain != nil && state.Chain.ProvenHeight > 0 {
		gap := state.Chain.HeadHeight - state.Chain.ProvenHeight
		if gap > 100 { // More than 100 blocks unproven
			state.SetWarnCondition("proof_lag")
			c.sendAlert(alerting.Alert{
				Type:        alerting.AlertProofLag,
				Level:       alerting.AlertLevelWarning,
				Title:       "Proof Lag Detected",
				Message:     fmt.Sprintf("There are %d unproven blocks (proven: %d, latest: %d).", gap, state.Chain.ProvenHeight, state.Chain.HeadHeight),
				BlockHeight: state.Chain.HeadHeight,
			})
		}
	}
}

// sendAlert sends an alert if alerting is enabled
func (c *Collector) sendAlert(alert alerting.Alert) {
	if c.alerter == nil {
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	if err := c.alerter.SendAlert(ctx, alert); err != nil {
		c.logger.Printf("Failed to send alert: %v", err)
	}
}

// sendRecoveryAlert sends a recovery notification
func (c *Collector) sendRecoveryAlert(alertType alerting.AlertType, message string) {
	if c.alerter == nil {
		return
	}

	c.alerter.ClearAlert(alertType)

	alert := alerting.Alert{
		Type:    alerting.AlertRecovered,
		Level:   alerting.AlertLevelInfo,
		Title:   "Recovery: " + string(alertType),
		Message: message,
	}

	ctx, cancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer cancel()

	if err := c.alerter.SendAlert(ctx, alert); err != nil {
		c.logger.Printf("Failed to send recovery alert: %v", err)
	}
}

// writeJSONLog writes the state to the JSON log file
func (c *Collector) writeJSONLog(state *CollectorState) {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		c.logger.Printf("Error marshaling state to JSON: %v", err)
		return
	}

	if err := os.WriteFile(c.config.JSONLog.Path, data, 0644); err != nil {
		c.logger.Printf("Error writing JSON log: %v", err)
	}
}

// startAPIServer starts the HTTP API server
func (c *Collector) startAPIServer() {
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// State endpoint
	mux.HandleFunc("/state", c.handleState)

	// Health endpoint
	mux.HandleFunc("/health", c.handleHealth)

	// Ready endpoint
	mux.HandleFunc("/ready", c.handleReady)

	// Config endpoint (for debugging)
	mux.HandleFunc("/config", c.handleConfig)

	server := &http.Server{
		Addr:    c.config.API.Listen,
		Handler: mux,
	}

	c.logger.Printf("Starting API server on %s", c.config.API.Listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		c.logger.Printf("API server error: %v", err)
	}
}

// handleState returns the current collector state as JSON
func (c *Collector) handleState(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	state := c.latestState
	c.mu.RUnlock()

	if state == nil {
		http.Error(w, "No state available", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(state)
}

// handleHealth returns the health status
func (c *Collector) handleHealth(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	state := c.latestState
	c.mu.RUnlock()

	if state == nil {
		http.Error(w, "No state available", http.StatusServiceUnavailable)
		return
	}

	status := http.StatusOK
	switch state.Health.Level {
	case HealthError, HealthFatal:
		status = http.StatusServiceUnavailable
	case HealthWarn:
		status = http.StatusOK // Still healthy, just warning
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(state.Health)
}

// handleReady returns whether the node is ready
func (c *Collector) handleReady(w http.ResponseWriter, r *http.Request) {
	c.mu.RLock()
	state := c.latestState
	c.mu.RUnlock()

	if state == nil || !state.Node.IsReady {
		http.Error(w, "Not ready", http.StatusServiceUnavailable)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// handleConfig returns the current configuration (sanitized)
func (c *Collector) handleConfig(w http.ResponseWriter, r *http.Request) {
	// Return a sanitized version of the config
	sanitized := struct {
		Protocol struct {
			Name     string `json:"name"`
			RPCURL   string `json:"rpc_url"`
			AdminURL string `json:"admin_url,omitempty"`
		} `json:"protocol"`
		Interval   string `json:"interval"`
		RefService struct {
			Enabled        bool   `json:"enabled"`
			UseExternalRPC bool   `json:"use_external_rpc"`
			Source         string `json:"source,omitempty"`
		} `json:"ref_service"`
	}{
		Protocol: struct {
			Name     string `json:"name"`
			RPCURL   string `json:"rpc_url"`
			AdminURL string `json:"admin_url,omitempty"`
		}{
			Name:     c.config.Protocol.Name,
			RPCURL:   c.config.Protocol.RPCURL,
			AdminURL: c.config.Protocol.AdminURL,
		},
		Interval: c.config.Interval.String(),
		RefService: struct {
			Enabled        bool   `json:"enabled"`
			UseExternalRPC bool   `json:"use_external_rpc"`
			Source         string `json:"source,omitempty"`
		}{
			Enabled:        c.config.RefService.Enabled,
			UseExternalRPC: c.config.RefService.UseExternalRPC,
		},
	}

	if c.config.RefService.UseExternalRPC {
		sanitized.RefService.Source = c.config.RefService.ExternalRPCURL
	} else if c.config.RefService.URL != "" {
		sanitized.RefService.Source = c.config.RefService.URL
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sanitized)
}

// GetLatestState returns the latest collected state
func (c *Collector) GetLatestState() *CollectorState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.latestState
}
