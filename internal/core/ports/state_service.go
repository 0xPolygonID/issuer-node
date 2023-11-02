package ports

import (
	"context"
	"math/big"

	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

// StateService is a service for working with state contract
type StateService interface {
	GetLatestStateByDID(ctx context.Context, did *w3c.DID) (abi.IStateStateInfo, error)
	GetGistRootInfo(ctx context.Context, did *w3c.DID, gist *big.Int) (abi.IStateGistRootInfo, error)
	GetGistProof(ctx context.Context, did *w3c.DID) (abi.IStateGistProof, error)
}
