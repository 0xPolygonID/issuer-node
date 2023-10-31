package ports

import (
	"context"
	"math/big"

	"github.com/iden3/go-iden3-core/v2/w3c"
)

// AccountService is a service for account operations
type AccountService interface {
	GetBalanceByDID(ctx context.Context, did *w3c.DID) (*big.Int, error)
}
