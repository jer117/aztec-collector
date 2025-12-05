package collector

import (
	"time"

	"github.com/aztec-collector/pkg/aztec"
)

// HealthLevel represents the health level of the node
type HealthLevel string

const (
	HealthOK    HealthLevel = "ok"
	HealthWarn  HealthLevel = "warn"
	HealthError HealthLevel = "error"
	HealthFatal HealthLevel = "fatal"
)

// Message represents a status message with level
type Message struct {
	Level   HealthLevel `json:"level"`
	Message string      `json:"message"`
	Time    time.Time   `json:"time"`
}

// ChainState represents the state of the blockchain
type ChainState struct {
	HeadHeight      int64     `json:"headHeight"`
	HeadHash        string    `json:"headHash,omitempty"`
	BlockTime       time.Time `json:"blockTime,omitempty"`
	ProvenHeight    int64     `json:"provenHeight,omitempty"`
	PendingHeight   int64     `json:"pendingHeight,omitempty"`
	NetworkHeight   int64     `json:"networkHeight,omitempty"`
	ReferenceHeight int64     `json:"referenceHeight,omitempty"`
}

// P2PState represents P2P network state
type P2PState struct {
	NumPeers int    `json:"numPeers"`
	PeerID   string `json:"peerId,omitempty"`
	EnrURI   string `json:"enrUri,omitempty"`
}

// NodeState represents node-specific state
type NodeState struct {
	NodeVersion     string `json:"nodeVersion,omitempty"`
	ProtocolVersion int64  `json:"protocolVersion,omitempty"`
	ChainID         int64  `json:"chainId,omitempty"`
	IsReady         bool   `json:"isReady"`
	IsSyncing       bool   `json:"isSyncing"`
}

// HealthState represents the health of the node
type HealthState struct {
	Level      HealthLevel `json:"level"`
	Conditions []string    `json:"conditions,omitempty"`
}

// CollectorState represents the full state collected from an Aztec node
type CollectorState struct {
	// Core state
	Chain  *ChainState  `json:"chain,omitempty"`
	P2P    *P2PState    `json:"p2p,omitempty"`
	Node   *NodeState   `json:"node,omitempty"`
	Health *HealthState `json:"health,omitempty"`

	// Aztec-specific state
	Aztec *aztec.AztecState `json:"aztec,omitempty"`

	// Reference service info
	RefService *RefServiceInfo `json:"refService,omitempty"`

	// Messages collected during polling
	Messages []Message `json:"messages,omitempty"`

	// Collection metadata
	CollectedAt time.Time `json:"collectedAt"`
}

// NewCollectorState creates a new collector state
func NewCollectorState() *CollectorState {
	return &CollectorState{
		Chain:       &ChainState{},
		P2P:         &P2PState{},
		Node:        &NodeState{},
		Health:      &HealthState{Level: HealthOK},
		Aztec:       &aztec.AztecState{},
		Messages:    []Message{},
		CollectedAt: time.Now(),
	}
}

// AddMessage adds a message to the state
func (s *CollectorState) AddMessage(level HealthLevel, msg string) {
	s.Messages = append(s.Messages, Message{
		Level:   level,
		Message: msg,
		Time:    time.Now(),
	})

	// Update health level if this message is worse
	if level > s.Health.Level {
		s.Health.Level = level
	}
}

// SetErrorCondition adds an error condition to health state
func (s *CollectorState) SetErrorCondition(condition string) {
	s.Health.Conditions = append(s.Health.Conditions, condition)
	if s.Health.Level < HealthError {
		s.Health.Level = HealthError
	}
}

// SetWarnCondition adds a warning condition to health state
func (s *CollectorState) SetWarnCondition(condition string) {
	s.Health.Conditions = append(s.Health.Conditions, condition)
	if s.Health.Level < HealthWarn {
		s.Health.Level = HealthWarn
	}
}

// GetBlocksBehind calculates how many blocks behind the node is compared to reference
func (s *CollectorState) GetBlocksBehind() (int64, bool) {
	if s.Chain == nil {
		return 0, false
	}

	// Use reference height if available, otherwise use network height
	ref := s.Chain.ReferenceHeight
	if ref == 0 {
		ref = s.Chain.NetworkHeight
	}
	if ref == 0 {
		return 0, false
	}

	behind := ref - s.Chain.HeadHeight
	if behind < 0 {
		behind = 0
	}
	return behind, true
}

// GetBlocksAhead calculates how many blocks ahead the node is compared to reference
func (s *CollectorState) GetBlocksAhead() (int64, bool) {
	if s.Chain == nil {
		return 0, false
	}

	// Use reference height if available, otherwise use network height
	ref := s.Chain.ReferenceHeight
	if ref == 0 {
		ref = s.Chain.NetworkHeight
	}
	if ref == 0 {
		return 0, false
	}

	ahead := s.Chain.HeadHeight - ref
	if ahead < 0 {
		ahead = 0
	}
	return ahead, true
}

// GetProofLag calculates the gap between latest and proven blocks
func (s *CollectorState) GetProofLag() (int64, bool) {
	if s.Chain == nil || s.Chain.ProvenHeight == 0 {
		return 0, false
	}

	lag := s.Chain.HeadHeight - s.Chain.ProvenHeight
	if lag < 0 {
		lag = 0
	}
	return lag, true
}

// UpdateFromL2Tips updates state from L2 tips
func (s *CollectorState) UpdateFromL2Tips(tips *aztec.L2Tips) {
	if tips == nil {
		return
	}

	s.Chain.HeadHeight = tips.Latest.Number
	s.Chain.HeadHash = tips.Latest.Hash
	s.Chain.ProvenHeight = tips.Proven.Number
	s.Chain.PendingHeight = tips.Pending.Number

	if tips.Latest.Timestamp > 0 {
		s.Chain.BlockTime = time.Unix(tips.Latest.Timestamp, 0)
	}
}

// UpdateFromNodeInfo updates state from node info
func (s *CollectorState) UpdateFromNodeInfo(info *aztec.NodeInfo) {
	if info == nil {
		return
	}

	s.Node.NodeVersion = info.NodeVersion
	s.Node.ProtocolVersion = info.ProtocolVersion
	s.Node.ChainID = info.L1ChainID
	s.P2P.EnrURI = info.EnrURI
}

// UpdateFromSyncStatus updates state from sync status
func (s *CollectorState) UpdateFromSyncStatus(status *aztec.WorldStateSyncStatus) {
	if status == nil {
		return
	}

	s.Node.IsSyncing = !status.Synced
	s.Chain.NetworkHeight = status.LatestBlock
}
