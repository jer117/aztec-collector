package collector

import (
	"fmt"

	"github.com/aztec-collector/pkg/aztec"
	"github.com/aztec-collector/pkg/jsonrpc"
)

// RPC method names for the Aztec node API
const (
	// Block queries
	MethodGetBlockNumber       = "node_getBlockNumber"
	MethodGetProvenBlockNumber = "node_getProvenBlockNumber"
	MethodGetL2Tips            = "node_getL2Tips"
	MethodGetBlock             = "node_getBlock"
	MethodGetBlocks            = "node_getBlocks"
	MethodGetBlockHeader       = "node_getBlockHeader"

	// Transaction operations
	MethodSendTx              = "node_sendTx"
	MethodGetTxReceipt        = "node_getTxReceipt"
	MethodGetTxEffect         = "node_getTxEffect"
	MethodGetTxByHash         = "node_getTxByHash"
	MethodGetPendingTxs       = "node_getPendingTxs"
	MethodGetPendingTxCount   = "node_getPendingTxCount"
	MethodIsValidTx           = "node_isValidTx"
	MethodSimulatePublicCalls = "node_simulatePublicCalls"

	// State queries
	MethodGetPublicStorageAt      = "node_getPublicStorageAt"
	MethodGetWorldStateSyncStatus = "node_getWorldStateSyncStatus"

	// Merkle tree queries
	MethodFindLeavesIndexes        = "node_findLeavesIndexes"
	MethodGetNullifierSiblingPath  = "node_getNullifierSiblingPath"
	MethodGetNoteHashSiblingPath   = "node_getNoteHashSiblingPath"
	MethodGetArchiveSiblingPath    = "node_getArchiveSiblingPath"
	MethodGetPublicDataSiblingPath = "node_getPublicDataSiblingPath"

	// Membership witnesses
	MethodGetNullifierMembershipWitness    = "node_getNullifierMembershipWitness"
	MethodGetLowNullifierMembershipWitness = "node_getLowNullifierMembershipWitness"
	MethodGetPublicDataWitness             = "node_getPublicDataWitness"
	MethodGetArchiveMembershipWitness      = "node_getArchiveMembershipWitness"
	MethodGetNoteHashMembershipWitness     = "node_getNoteHashMembershipWitness"

	// L1 to L2 messages
	MethodGetL1ToL2MessageMembershipWitness = "node_getL1ToL2MessageMembershipWitness"
	MethodGetL1ToL2MessageBlock             = "node_getL1ToL2MessageBlock"
	MethodIsL1ToL2MessageSynced             = "node_isL1ToL2MessageSynced"
	MethodGetL2ToL1Messages                 = "node_getL2ToL1Messages"

	// Log queries
	MethodGetPrivateLogs      = "node_getPrivateLogs"
	MethodGetPublicLogs       = "node_getPublicLogs"
	MethodGetContractClassLogs = "node_getContractClassLogs"
	MethodGetLogsByTags       = "node_getLogsByTags"

	// Contract queries
	MethodGetContractClass = "node_getContractClass"
	MethodGetContract      = "node_getContract"

	// Node information
	MethodIsReady                     = "node_isReady"
	MethodGetNodeInfo                 = "node_getNodeInfo"
	MethodGetNodeVersion              = "node_getNodeVersion"
	MethodGetVersion                  = "node_getVersion"
	MethodGetChainId                  = "node_getChainId"
	MethodGetL1ContractAddresses      = "node_getL1ContractAddresses"
	MethodGetProtocolContractAddresses = "node_getProtocolContractAddresses"
	MethodGetEncodedEnr               = "node_getEncodedEnr"
	MethodGetCurrentBaseFees          = "node_getCurrentBaseFees"

	// Validator queries
	MethodGetValidatorsStats = "node_getValidatorsStats"
	MethodGetValidatorStats  = "node_getValidatorStats"

	// Debug operations
	MethodRegisterContractFunctionSignatures = "node_registerContractFunctionSignatures"
	MethodGetAllowedPublicSetup              = "node_getAllowedPublicSetup"

	// Admin API
	MethodAdminGetConfig            = "nodeAdmin_getConfig"
	MethodAdminSetConfig            = "nodeAdmin_setConfig"
	MethodAdminPauseSync            = "nodeAdmin_pauseSync"
	MethodAdminResumeSync           = "nodeAdmin_resumeSync"
	MethodAdminRollbackTo           = "nodeAdmin_rollbackTo"
	MethodAdminStartSnapshotUpload  = "nodeAdmin_startSnapshotUpload"
	MethodAdminGetSlashPayloads     = "nodeAdmin_getSlashPayloads"
	MethodAdminGetSlashOffenses     = "nodeAdmin_getSlashOffenses"
)

// AztecRPC provides methods to query the Aztec node
type AztecRPC struct {
	client *jsonrpc.Client
}

// NewAztecRPC creates a new Aztec RPC client
func NewAztecRPC(client *jsonrpc.Client) *AztecRPC {
	return &AztecRPC{client: client}
}

// GetBlockNumber returns the latest block number
func (r *AztecRPC) GetBlockNumber() (int64, error) {
	var result int64
	if err := r.client.Call(MethodGetBlockNumber, []interface{}{}, &result); err != nil {
		return 0, fmt.Errorf("GetBlockNumber: %w", err)
	}
	return result, nil
}

// GetProvenBlockNumber returns the latest proven block number
func (r *AztecRPC) GetProvenBlockNumber() (int64, error) {
	var result int64
	if err := r.client.Call(MethodGetProvenBlockNumber, []interface{}{}, &result); err != nil {
		return 0, fmt.Errorf("GetProvenBlockNumber: %w", err)
	}
	return result, nil
}

// GetL2Tips returns the L2 chain tips
func (r *AztecRPC) GetL2Tips() (*aztec.L2Tips, error) {
	var result aztec.L2Tips
	if err := r.client.Call(MethodGetL2Tips, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetL2Tips: %w", err)
	}
	return &result, nil
}

// GetBlock returns a block by number
func (r *AztecRPC) GetBlock(blockNumber interface{}) (*aztec.L2Block, error) {
	var result aztec.L2Block
	if err := r.client.Call(MethodGetBlock, []interface{}{blockNumber}, &result); err != nil {
		return nil, fmt.Errorf("GetBlock: %w", err)
	}
	return &result, nil
}

// GetBlocks returns multiple blocks in a range
func (r *AztecRPC) GetBlocks(from int64, limit int) ([]aztec.L2Block, error) {
	var result []aztec.L2Block
	if err := r.client.Call(MethodGetBlocks, []interface{}{from, limit}, &result); err != nil {
		return nil, fmt.Errorf("GetBlocks: %w", err)
	}
	return result, nil
}

// GetBlockHeader returns a block header
func (r *AztecRPC) GetBlockHeader(blockNumber interface{}) (*aztec.BlockHeader, error) {
	var result aztec.BlockHeader
	if err := r.client.Call(MethodGetBlockHeader, []interface{}{blockNumber}, &result); err != nil {
		return nil, fmt.Errorf("GetBlockHeader: %w", err)
	}
	return &result, nil
}

// GetTxReceipt returns a transaction receipt
func (r *AztecRPC) GetTxReceipt(txHash string) (*aztec.TxReceipt, error) {
	var result aztec.TxReceipt
	if err := r.client.Call(MethodGetTxReceipt, []interface{}{txHash}, &result); err != nil {
		return nil, fmt.Errorf("GetTxReceipt: %w", err)
	}
	return &result, nil
}

// GetTxEffect returns the transaction effect
func (r *AztecRPC) GetTxEffect(txHash string) (*aztec.IndexedTxEffect, error) {
	var result aztec.IndexedTxEffect
	if err := r.client.Call(MethodGetTxEffect, []interface{}{txHash}, &result); err != nil {
		return nil, fmt.Errorf("GetTxEffect: %w", err)
	}
	return &result, nil
}

// GetTxByHash returns a pending transaction by hash
func (r *AztecRPC) GetTxByHash(txHash string) (*aztec.Tx, error) {
	var result aztec.Tx
	if err := r.client.Call(MethodGetTxByHash, []interface{}{txHash}, &result); err != nil {
		return nil, fmt.Errorf("GetTxByHash: %w", err)
	}
	return &result, nil
}

// GetPendingTxs returns pending transactions
func (r *AztecRPC) GetPendingTxs(limit int, after *string) ([]aztec.Tx, error) {
	params := []interface{}{limit}
	if after != nil {
		params = append(params, *after)
	}
	var result []aztec.Tx
	if err := r.client.Call(MethodGetPendingTxs, params, &result); err != nil {
		return nil, fmt.Errorf("GetPendingTxs: %w", err)
	}
	return result, nil
}

// GetPendingTxCount returns the count of pending transactions
func (r *AztecRPC) GetPendingTxCount() (int, error) {
	var result int
	if err := r.client.Call(MethodGetPendingTxCount, []interface{}{}, &result); err != nil {
		return 0, fmt.Errorf("GetPendingTxCount: %w", err)
	}
	return result, nil
}

// GetWorldStateSyncStatus returns the sync status
func (r *AztecRPC) GetWorldStateSyncStatus() (*aztec.WorldStateSyncStatus, error) {
	var result aztec.WorldStateSyncStatus
	if err := r.client.Call(MethodGetWorldStateSyncStatus, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetWorldStateSyncStatus: %w", err)
	}
	return &result, nil
}

// IsReady checks if the node is ready
func (r *AztecRPC) IsReady() (bool, error) {
	var result bool
	if err := r.client.Call(MethodIsReady, []interface{}{}, &result); err != nil {
		return false, fmt.Errorf("IsReady: %w", err)
	}
	return result, nil
}

// GetNodeInfo returns node information
func (r *AztecRPC) GetNodeInfo() (*aztec.NodeInfo, error) {
	var result aztec.NodeInfo
	if err := r.client.Call(MethodGetNodeInfo, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetNodeInfo: %w", err)
	}
	return &result, nil
}

// GetNodeVersion returns the node version
func (r *AztecRPC) GetNodeVersion() (string, error) {
	var result string
	if err := r.client.Call(MethodGetNodeVersion, []interface{}{}, &result); err != nil {
		return "", fmt.Errorf("GetNodeVersion: %w", err)
	}
	return result, nil
}

// GetVersion returns the protocol version
func (r *AztecRPC) GetVersion() (int64, error) {
	var result int64
	if err := r.client.Call(MethodGetVersion, []interface{}{}, &result); err != nil {
		return 0, fmt.Errorf("GetVersion: %w", err)
	}
	return result, nil
}

// GetChainID returns the chain ID
func (r *AztecRPC) GetChainID() (int64, error) {
	var result int64
	if err := r.client.Call(MethodGetChainId, []interface{}{}, &result); err != nil {
		return 0, fmt.Errorf("GetChainID: %w", err)
	}
	return result, nil
}

// GetCurrentBaseFees returns the current base fees
func (r *AztecRPC) GetCurrentBaseFees() (*aztec.CurrentBaseFees, error) {
	var result aztec.CurrentBaseFees
	if err := r.client.Call(MethodGetCurrentBaseFees, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetCurrentBaseFees: %w", err)
	}
	return &result, nil
}

// GetValidatorsStats returns validator statistics
func (r *AztecRPC) GetValidatorsStats() (*aztec.ValidatorsStatsResponse, error) {
	var result aztec.ValidatorsStatsResponse
	if err := r.client.Call(MethodGetValidatorsStats, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetValidatorsStats: %w", err)
	}
	return &result, nil
}

// GetPublicStorageAt returns a public storage value
func (r *AztecRPC) GetPublicStorageAt(blockNumber interface{}, contract, slot string) (string, error) {
	var result string
	if err := r.client.Call(MethodGetPublicStorageAt, []interface{}{blockNumber, contract, slot}, &result); err != nil {
		return "", fmt.Errorf("GetPublicStorageAt: %w", err)
	}
	return result, nil
}

// GetL1ContractAddresses returns L1 contract addresses
func (r *AztecRPC) GetL1ContractAddresses() (*aztec.L1ContractAddresses, error) {
	var result aztec.L1ContractAddresses
	if err := r.client.Call(MethodGetL1ContractAddresses, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetL1ContractAddresses: %w", err)
	}
	return &result, nil
}

// GetProtocolContractAddresses returns protocol contract addresses
func (r *AztecRPC) GetProtocolContractAddresses() (*aztec.ProtocolContractAddresses, error) {
	var result aztec.ProtocolContractAddresses
	if err := r.client.Call(MethodGetProtocolContractAddresses, []interface{}{}, &result); err != nil {
		return nil, fmt.Errorf("GetProtocolContractAddresses: %w", err)
	}
	return &result, nil
}

// GetEncodedEnr returns the encoded ENR
func (r *AztecRPC) GetEncodedEnr() (string, error) {
	var result string
	if err := r.client.Call(MethodGetEncodedEnr, []interface{}{}, &result); err != nil {
		return "", fmt.Errorf("GetEncodedEnr: %w", err)
	}
	return result, nil
}

