// Package aztec provides data models for the Aztec node API
package aztec

import (
	"encoding/json"
	"time"
)

// L2Tips represents the tips of the L2 chain
type L2Tips struct {
	Latest  BlockTip `json:"latest"`
	Pending BlockTip `json:"pending"`
	Proven  BlockTip `json:"proven"`
}

// BlockTip represents information about a block tip
type BlockTip struct {
	Number    int64  `json:"number"`
	Hash      string `json:"hash,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

// L2Block represents a full L2 block
type L2Block struct {
	Header       *BlockHeader `json:"header,omitempty"`
	Body         *BlockBody   `json:"body,omitempty"`
	Archive      string       `json:"archive,omitempty"`
	Number       int64        `json:"number,omitempty"`
	Hash         string       `json:"hash,omitempty"`
	TxCount      int          `json:"txCount,omitempty"`
	TxEffects    []TxEffect   `json:"txEffects,omitempty"`
	NumTxs       int          `json:"numTxs,omitempty"`
	BlockNumber  int64        `json:"blockNumber,omitempty"`
	Transactions []string     `json:"transactions,omitempty"`
}

// BlockHeader represents the header of an L2 block
type BlockHeader struct {
	LastArchive            *AppendOnlyTreeSnapshot `json:"lastArchive,omitempty"`
	ContentCommitment      *ContentCommitment      `json:"contentCommitment,omitempty"`
	State                  *StateReference         `json:"state,omitempty"`
	GlobalVariables        *GlobalVariables        `json:"globalVariables,omitempty"`
	TotalFees              string                  `json:"totalFees,omitempty"`
	TotalManaUsed          string                  `json:"totalManaUsed,omitempty"`
	BodyHash               string                  `json:"bodyHash,omitempty"`
	BlockNumber            int64                   `json:"blockNumber,omitempty"`
	SlotNumber             int64                   `json:"slotNumber,omitempty"`
	Timestamp              int64                   `json:"timestamp,omitempty"`
	Coinbase               string                  `json:"coinbase,omitempty"`
	FeeRecipient           string                  `json:"feeRecipient,omitempty"`
	GasFees                *GasFees                `json:"gasFees,omitempty"`
	ProofClaim             json.RawMessage         `json:"proofClaim,omitempty"`
}

// AppendOnlyTreeSnapshot represents a snapshot of an append-only tree
type AppendOnlyTreeSnapshot struct {
	Root            string `json:"root"`
	NextAvailableId int64  `json:"nextAvailableId"`
}

// ContentCommitment represents the content commitment of a block
type ContentCommitment struct {
	NumTxs        int64  `json:"numTxs"`
	TxsHash       string `json:"txsHash"`
	InHash        string `json:"inHash"`
	OutHash       string `json:"outHash"`
	BlobsHash     string `json:"blobsHash,omitempty"`
	BlobPublicSum string `json:"blobPublicSum,omitempty"`
}

// StateReference represents the state reference of a block
type StateReference struct {
	L1ToL2MessageTree *AppendOnlyTreeSnapshot `json:"l1ToL2MessageTree,omitempty"`
	PartialStateRef   *PartialStateReference  `json:"partial,omitempty"`
}

// PartialStateReference represents a partial state reference
type PartialStateReference struct {
	NoteHashTree   *AppendOnlyTreeSnapshot `json:"noteHashTree,omitempty"`
	NullifierTree  *AppendOnlyTreeSnapshot `json:"nullifierTree,omitempty"`
	PublicDataTree *AppendOnlyTreeSnapshot `json:"publicDataTree,omitempty"`
}

// GlobalVariables represents global variables in a block
type GlobalVariables struct {
	ChainID      int64    `json:"chainId"`
	Version      int64    `json:"version"`
	BlockNumber  int64    `json:"blockNumber"`
	SlotNumber   int64    `json:"slotNumber"`
	Timestamp    int64    `json:"timestamp"`
	Coinbase     string   `json:"coinbase"`
	FeeRecipient string   `json:"feeRecipient"`
	GasFees      *GasFees `json:"gasFees,omitempty"`
}

// GasFees represents gas fee information
type GasFees struct {
	FeePerDaGas   string `json:"feePerDaGas"`
	FeePerL2Gas   string `json:"feePerL2Gas"`
	CongestionCost string `json:"congestionCost,omitempty"`
}

// BlockBody represents the body of a block
type BlockBody struct {
	TxEffects []TxEffect `json:"txEffects,omitempty"`
}

// TxEffect represents the effect of a transaction
type TxEffect struct {
	TxHash           string          `json:"txHash,omitempty"`
	NoteHashes       []string        `json:"noteHashes,omitempty"`
	Nullifiers       []string        `json:"nullifiers,omitempty"`
	L2ToL1Msgs       []string        `json:"l2ToL1Msgs,omitempty"`
	PublicDataWrites []string        `json:"publicDataWrites,omitempty"`
	Logs             json.RawMessage `json:"logs,omitempty"`
}

// TxReceipt represents a transaction receipt
type TxReceipt struct {
	TxHash      string `json:"txHash"`
	Status      string `json:"status"` // mined, pending, dropped
	BlockHash   string `json:"blockHash,omitempty"`
	BlockNumber int64  `json:"blockNumber,omitempty"`
	Error       string `json:"error,omitempty"`
}

// IndexedTxEffect represents an indexed transaction effect
type IndexedTxEffect struct {
	TxEffect    *TxEffect `json:"txEffect,omitempty"`
	BlockNumber int64     `json:"blockNumber"`
	TxIndex     int       `json:"txIndex"`
}

// Tx represents a transaction
type Tx struct {
	Data     string          `json:"data,omitempty"`
	Hash     string          `json:"hash,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

// NodeInfo represents information about the node
type NodeInfo struct {
	NodeVersion         string          `json:"nodeVersion"`
	L1ChainID           int64           `json:"l1ChainId"`
	ProtocolVersion     int64           `json:"protocolVersion"`
	EnrURI              string          `json:"enrUri,omitempty"`
	L1ContractAddresses json.RawMessage `json:"l1ContractAddresses,omitempty"`
	ProtocolContracts   json.RawMessage `json:"protocolContracts,omitempty"`
}

// WorldStateSyncStatus represents the sync status of world state
type WorldStateSyncStatus struct {
	Synced        bool  `json:"synced"`
	LatestBlock   int64 `json:"latestBlock"`
	ProvenBlock   int64 `json:"provenBlock"`
	SyncedToBlock int64 `json:"syncedToBlock,omitempty"`
}

// L1ContractAddresses represents L1 contract addresses
type L1ContractAddresses struct {
	RollupAddress       string `json:"rollupAddress,omitempty"`
	RegistryAddress     string `json:"registryAddress,omitempty"`
	InboxAddress        string `json:"inboxAddress,omitempty"`
	OutboxAddress       string `json:"outboxAddress,omitempty"`
	FeeJuiceAddress     string `json:"feeJuiceAddress,omitempty"`
	FeeJuicePortal      string `json:"feeJuicePortal,omitempty"`
	CoinIssuer          string `json:"coinIssuer,omitempty"`
	RewardDistributor   string `json:"rewardDistributor,omitempty"`
	GovernanceProposer  string `json:"governanceProposer,omitempty"`
	Governance          string `json:"governance,omitempty"`
	SlashFactory        string `json:"slashFactory,omitempty"`
	StakingAsset        string `json:"stakingAsset,omitempty"`
}

// ProtocolContractAddresses represents protocol contract addresses
type ProtocolContractAddresses struct {
	ClassRegisterer   string `json:"classRegisterer,omitempty"`
	FeeJuice          string `json:"feeJuice,omitempty"`
	InstanceDeployer  string `json:"instanceDeployer,omitempty"`
	MultiCallEntrypoint string `json:"multiCallEntrypoint,omitempty"`
}

// ValidatorsStatsResponse represents the response from node_getValidatorsStats
type ValidatorsStatsResponse struct {
	Stats             map[string]*ValidatorStats `json:"stats"`
	LastProcessedSlot string                     `json:"lastProcessedSlot"`
	InitialSlot       string                     `json:"initialSlot"`
	SlotWindow        int64                      `json:"slotWindow"`
}

// ValidatorStats represents statistics about a single validator
type ValidatorStats struct {
	Address            string              `json:"address"`
	LastProposal       *SlotEvent          `json:"lastProposal,omitempty"`
	LastAttestation    *SlotEvent          `json:"lastAttestation,omitempty"`
	TotalSlots         int64               `json:"totalSlots"`
	MissedProposals    *MissedStats        `json:"missedProposals"`
	MissedAttestations *MissedStats        `json:"missedAttestations"`
	History            []SlotHistoryEntry  `json:"history,omitempty"`
}

// SlotEvent represents a slot event (proposal or attestation)
type SlotEvent struct {
	Timestamp string `json:"timestamp"`
	Slot      string `json:"slot"`
	Date      string `json:"date"`
}

// MissedStats represents missed proposal/attestation statistics
type MissedStats struct {
	CurrentStreak int64   `json:"currentStreak"`
	Rate          float64 `json:"rate"`
	Count         int64   `json:"count"`
	Total         int64   `json:"total"`
}

// SlotHistoryEntry represents a single slot history entry
type SlotHistoryEntry struct {
	Slot   string `json:"slot"`
	Status string `json:"status"` // block-mined, attestation-sent, block-missed, attestation-missed, block-proposed
}

// ValidatorInfo is a simplified view of validator status (for backwards compatibility)
type ValidatorInfo struct {
	Address       string `json:"address,omitempty"`
	Stake         string `json:"stake,omitempty"`
	Status        string `json:"status,omitempty"`
	AttestCount   int64  `json:"attestCount,omitempty"`
	MissedSlots   int64  `json:"missedSlots,omitempty"`
	ProposedCount int64  `json:"proposedCount,omitempty"`
}

// CurrentBaseFees represents current base fees
type CurrentBaseFees struct {
	FeePerDaGas string `json:"feePerDaGas,omitempty"`
	FeePerL2Gas string `json:"feePerL2Gas,omitempty"`
}

// AztecState combines all Aztec-specific state for collection
type AztecState struct {
	L2Tips            *L2Tips                  `json:"l2Tips,omitempty"`
	LatestBlock       *L2Block                 `json:"latestBlock,omitempty"`
	NodeInfo          *NodeInfo                `json:"nodeInfo,omitempty"`
	SyncStatus        *WorldStateSyncStatus    `json:"syncStatus,omitempty"`
	PendingTxCount    int                      `json:"pendingTxCount,omitempty"`
	CurrentBaseFees   *CurrentBaseFees         `json:"currentBaseFees,omitempty"`
	ValidatorsStats   *ValidatorsStatsResponse `json:"validatorsStats,omitempty"`
	CollectionTime    time.Time                `json:"collectionTime"`
}

// ToJSON converts AztecState to JSON
func (s *AztecState) ToJSON() ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}

