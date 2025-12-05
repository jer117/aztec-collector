package collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains Prometheus metrics for the Aztec collector
type Metrics struct {
	// Block metrics
	BlockHeight        *prometheus.GaugeVec
	ProvenBlockHeight  *prometheus.GaugeVec
	PendingBlockHeight *prometheus.GaugeVec
	ReferenceHeight    *prometheus.GaugeVec
	BlocksBehind       *prometheus.GaugeVec
	BlocksAhead        *prometheus.GaugeVec
	ProofLag           *prometheus.GaugeVec

	// Transaction metrics
	PendingTxCount *prometheus.GaugeVec

	// Sync metrics
	IsSynced  *prometheus.GaugeVec
	IsSyncing *prometheus.GaugeVec
	IsReady   *prometheus.GaugeVec

	// P2P metrics
	NumPeers *prometheus.GaugeVec

	// Gas metrics
	BaseFeePerDaGas *prometheus.GaugeVec
	BaseFeePerL2Gas *prometheus.GaugeVec

	// Health metrics
	HealthStatus *prometheus.GaugeVec

	// Collection metrics
	CollectionDuration *prometheus.HistogramVec
	CollectionErrors   *prometheus.CounterVec
	LastCollectionTime *prometheus.GaugeVec
}

// NewMetrics creates new Prometheus metrics
func NewMetrics(r prometheus.Registerer) *Metrics {
	namespace := "aztec"

	return &Metrics{
		BlockHeight: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "block_height",
				Help:      "Current block height",
			}, []string{}),

		ProvenBlockHeight: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "proven_block_height",
				Help:      "Current proven block height",
			}, []string{}),

		PendingBlockHeight: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pending_block_height",
				Help:      "Current pending block height",
			}, []string{}),

		ReferenceHeight: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "reference_height",
				Help:      "Reference block height from external source",
			}, []string{}),

		BlocksBehind: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "blocks_behind",
				Help:      "Number of blocks behind reference",
			}, []string{}),

		BlocksAhead: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "blocks_ahead",
				Help:      "Number of blocks ahead of reference",
			}, []string{}),

		ProofLag: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "proof_lag",
				Help:      "Gap between latest and proven block height",
			}, []string{}),

		PendingTxCount: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pending_tx_count",
				Help:      "Number of pending transactions in mempool",
			}, []string{}),

		IsSynced: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "is_synced",
				Help:      "Whether the node is synced (1 = synced, 0 = not synced)",
			}, []string{}),

		IsSyncing: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "is_syncing",
				Help:      "Whether the node is currently syncing (1 = syncing, 0 = not syncing)",
			}, []string{}),

		IsReady: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "is_ready",
				Help:      "Whether the node is ready (1 = ready, 0 = not ready)",
			}, []string{}),

		NumPeers: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "num_peers",
				Help:      "Number of connected peers",
			}, []string{}),

		BaseFeePerDaGas: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "base_fee_per_da_gas",
				Help:      "Current base fee per DA gas",
			}, []string{}),

		BaseFeePerL2Gas: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "base_fee_per_l2_gas",
				Help:      "Current base fee per L2 gas",
			}, []string{}),

		HealthStatus: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "health_status",
				Help:      "Health status of the node (0=ok, 1=warn, 2=error, 3=fatal)",
			}, []string{}),

		CollectionDuration: promauto.With(r).NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "collection_duration_seconds",
				Help:      "Time taken to collect metrics",
				Buckets:   prometheus.DefBuckets,
			}, []string{}),

		CollectionErrors: promauto.With(r).NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "collection_errors_total",
				Help:      "Total number of collection errors",
			}, []string{"method"}),

		LastCollectionTime: promauto.With(r).NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "last_collection_timestamp",
				Help:      "Unix timestamp of last successful collection",
			}, []string{}),
	}
}

// UpdateFromState updates metrics from collector state
func (m *Metrics) UpdateFromState(state *CollectorState) {
	if state.Chain != nil {
		m.BlockHeight.WithLabelValues().Set(float64(state.Chain.HeadHeight))
		m.ProvenBlockHeight.WithLabelValues().Set(float64(state.Chain.ProvenHeight))
		m.PendingBlockHeight.WithLabelValues().Set(float64(state.Chain.PendingHeight))

		if state.Chain.ReferenceHeight > 0 {
			m.ReferenceHeight.WithLabelValues().Set(float64(state.Chain.ReferenceHeight))
		}

		if behind, ok := state.GetBlocksBehind(); ok {
			m.BlocksBehind.WithLabelValues().Set(float64(behind))
		}

		if ahead, ok := state.GetBlocksAhead(); ok {
			m.BlocksAhead.WithLabelValues().Set(float64(ahead))
		}

		if lag, ok := state.GetProofLag(); ok {
			m.ProofLag.WithLabelValues().Set(float64(lag))
		}
	}

	if state.Node != nil {
		if state.Node.IsReady {
			m.IsReady.WithLabelValues().Set(1)
		} else {
			m.IsReady.WithLabelValues().Set(0)
		}

		if state.Node.IsSyncing {
			m.IsSyncing.WithLabelValues().Set(1)
			m.IsSynced.WithLabelValues().Set(0)
		} else {
			m.IsSyncing.WithLabelValues().Set(0)
			m.IsSynced.WithLabelValues().Set(1)
		}
	}

	if state.Aztec != nil {
		m.PendingTxCount.WithLabelValues().Set(float64(state.Aztec.PendingTxCount))

		if state.Aztec.L2Tips != nil {
			m.PendingBlockHeight.WithLabelValues().Set(float64(state.Aztec.L2Tips.Pending.Number))
		}
	}

	// Health status
	healthValue := 0.0
	switch state.Health.Level {
	case HealthOK:
		healthValue = 0
	case HealthWarn:
		healthValue = 1
	case HealthError:
		healthValue = 2
	case HealthFatal:
		healthValue = 3
	}
	m.HealthStatus.WithLabelValues().Set(healthValue)

	m.LastCollectionTime.WithLabelValues().Set(float64(state.CollectedAt.Unix()))
}

// RecordCollectionError records a collection error for a specific method
func (m *Metrics) RecordCollectionError(method string) {
	m.CollectionErrors.WithLabelValues(method).Inc()
}
