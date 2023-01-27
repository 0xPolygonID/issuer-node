package ports

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
)

// TransactionService interface
type TransactionService interface {
	WaitForTransactionReceipt(ctx context.Context, txID string) (*types.Receipt, error)
	WaitForConfirmation(ctx context.Context, receipt *types.Receipt) (bool, error)
	GetHeaderByNumber(ctx context.Context, blockNumber *big.Int) (*types.Header, error)
	CheckConfirmation(ctx context.Context, receipt *types.Receipt) (bool, error)
	GetTransactionReceiptByID(ctx context.Context, txID string) (*types.Receipt, error)
}
