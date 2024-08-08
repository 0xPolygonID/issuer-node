package gateways

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/network"
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
	networkResolver network.Resolver
}

// NewTransaction new transaction gateway
func NewTransaction(networkResolver network.Resolver) (*transaction, error) {
	return &transaction{networkResolver: networkResolver}, nil
}

// CheckConfirmation check tx confirmation status
func (tr *transaction) CheckConfirmation(ctx context.Context, identity *domain.Identity, receipt *types.Receipt, confirmationBlockCount int64) (bool, error) {
	client, err := getEthClient(ctx, identity, tr.networkResolver)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return false, err
	}

	currentBlock, err := client.CurrentBlock(ctx)
	if err != nil {
		return false, err
	}

	blocks := currentBlock.Sub(currentBlock, receipt.BlockNumber)

	if blocks.Int64() < confirmationBlockCount {
		return false, nil
	}

	return true, err
}

// WaitForTransactionReceipt wait for ETH tx receipt
func (tr *transaction) WaitForTransactionReceipt(ctx context.Context, identity *domain.Identity, txID string) (*types.Receipt, error) {
	client, err := getEthClient(ctx, identity, tr.networkResolver)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return nil, err
	}
	receipt, err := client.WaitTransactionReceiptByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// GetHeaderByNumber get Eth block header by block number
func (tr *transaction) GetHeaderByNumber(ctx context.Context, identity *domain.Identity, blockNumber *big.Int) (*types.Header, error) {
	client, err := getEthClient(ctx, identity, tr.networkResolver)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return nil, err
	}
	header, err := client.HeaderByNumber(ctx, blockNumber)
	if err != nil {
		return nil, err
	}
	return header, err
}

// GetTransactionReceiptByID  returns tx receipt
func (tr *transaction) GetTransactionReceiptByID(ctx context.Context, identity *domain.Identity, txID string) (*types.Receipt, error) {
	client, err := getEthClient(ctx, identity, tr.networkResolver)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return nil, err
	}
	receipt, err := client.GetTransactionReceiptByID(ctx, txID)
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

// WaitForConfirmation wait until transaction will be confirmed
func (tr *transaction) WaitForConfirmation(ctx context.Context, identity *domain.Identity, receipt *types.Receipt) (bool, error) {
	client, err := getEthClient(ctx, identity, tr.networkResolver)
	if err != nil {
		log.Error(ctx, "failed to get client", "err", err)
		return false, err
	}

	confirmationBlockCount, err := tr.getConfirmationBlockCount(ctx, identity)
	if err != nil {
		log.Error(ctx, "failed to get confirmation block count", "err", err)
		return false, err
	}

	confirmationBlock := big.NewInt(confirmationBlockCount)
	confirmationBlock = confirmationBlock.Add(confirmationBlock, receipt.BlockNumber)
	err = client.WaitForBlock(ctx, confirmationBlock)
	if err != nil {
		return false, err
	}
	return true, err
}

func (tr *transaction) getConfirmationBlockCount(ctx context.Context, identity *domain.Identity) (int64, error) {
	resolverPrefix, err := identity.GetResolverPrefix()
	if err != nil {
		log.Error(ctx, "failed to get networkResolver prefix", "err", err)
		return 0, err
	}

	confirmationBlockCount, err := tr.networkResolver.GetConfirmationBlockCount(resolverPrefix)
	if err != nil {
		log.Error(ctx, "failed to get confirmation block count", "err", err)
		return 0, err
	}

	return confirmationBlockCount, nil
}
