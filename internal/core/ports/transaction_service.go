package ports

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// TransactionService interface
type TransactionService interface {
	WaitForTransactionReceipt(ctx context.Context, identity *domain.Identity, txID string) (*types.Receipt, error)
	WaitForConfirmation(ctx context.Context, identity *domain.Identity, receipt *types.Receipt) (bool, error)
	GetHeaderByNumber(ctx context.Context, identity *domain.Identity, blockNumber *big.Int) (*types.Header, error)
	CheckConfirmation(ctx context.Context, identity *domain.Identity, receipt *types.Receipt, confirmationBlockCount int64) (bool, error)
	GetTransactionReceiptByID(ctx context.Context, identity *domain.Identity, txID string) (*types.Receipt, error)
}
