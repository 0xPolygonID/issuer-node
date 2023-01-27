package gateways

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// ETHClient defines interface for ethereum client
type ETHClient interface {
	GetTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error)
	WaitTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error)
	CurrentBlock(ctx context.Context) (*big.Int, error)
	BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	WaitForBlock(ctx context.Context, confirmationBlock *big.Int) error
}

// TransactionService blockchain tx service
type transaction struct {
	client                 ETHClient
	confirmationBlockCount int64
}

// NewTransaction new transaction gateway
func NewTransaction(_client ETHClient, confirmationBlockCount int64) (*transaction, error) {
	return &transaction{client: _client, confirmationBlockCount: confirmationBlockCount}, nil
}

// CheckConfirmation check tx confirmation status
func (tr *transaction) CheckConfirmation(ctx context.Context, receipt *types.Receipt) (bool, error) {
	currentBlock, err := tr.client.CurrentBlock(ctx)
	if err != nil {
		return false, err
	}

	blocks := currentBlock.Sub(currentBlock, receipt.BlockNumber)

	if blocks.Int64() < tr.confirmationBlockCount {
		return false, nil
	}

	return true, err
}

// WaitForTransactionReceipt wait for ETH tx receipt
func (tr *transaction) WaitForTransactionReceipt(ctx context.Context, txID string) (*types.Receipt, error) {
	receipt, err := tr.client.WaitTransactionReceiptByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetHeaderByNumber get Eth block header by block number
func (tr *transaction) GetHeaderByNumber(ctx context.Context, blockNumber *big.Int) (*types.Header, error) {
	header, err := tr.client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return nil, err
	}
	return header, err
}

// GetTransactionReceiptByID  returns tx receipt
func (tr *transaction) GetTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error) {
	receipt, err := tr.client.GetTransactionReceiptByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// WaitForConfirmation wait until transaction will be confirmed
func (tr *transaction) WaitForConfirmation(ctx context.Context, receipt *types.Receipt) (bool, error) {
	confirmationBlock := big.NewInt(tr.confirmationBlockCount)
	confirmationBlock = confirmationBlock.Add(confirmationBlock, receipt.BlockNumber)
	err := tr.client.WaitForBlock(ctx, confirmationBlock)
	if err != nil {
		return false, err
	}
	return true, err
}
