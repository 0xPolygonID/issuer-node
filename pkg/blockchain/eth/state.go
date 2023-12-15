package eth

import (
	"context"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/log"
)

// StateServiceConfig is a config for StateService
type StateServiceConfig struct {
	EthClient       *Client
	StateAddress    ethCommon.Address
	ResponseTimeout time.Duration
}

// StateService is a service for working with state contract
type StateService struct {
	rw              *sync.RWMutex
	contractBinding *abi.State
	rpcConfig       StateServiceConfig
}

// NewStateService creates new instance of StateService
func NewStateService(rpcConfig StateServiceConfig) (*StateService, error) {
	contractBinding, err := abi.NewState(rpcConfig.StateAddress, rpcConfig.EthClient.GetEthereumClient())
	if err != nil {
		log.Error(context.Background(), "Failed to create state contract binding", "error", err)
		return &StateService{}, err
	}
	return &StateService{
		contractBinding: contractBinding,
		rw:              &sync.RWMutex{},
		rpcConfig:       rpcConfig,
	}, nil
}

// GetLatestStateByDID returns latest state info for DID
func (ss *StateService) GetLatestStateByDID(ctx context.Context, did *w3c.DID) (abi.IStateStateInfo, error) {
	var (
		latestState abi.IStateStateInfo
		err         error
	)
	id, err := core.IDFromDID(*did)
	if err != nil {
		return abi.IStateStateInfo{}, err
	}
	latestState, err = ss.contractBinding.GetStateInfoById(&bind.CallOpts{Context: ctx}, id.BigInt())
	if err != nil {
		return latestState, err
	}
	return latestState, nil
}

// GetGistRootInfo returns global state info
func (ss *StateService) GetGistRootInfo(ctx context.Context, did *w3c.DID, gist *big.Int) (abi.IStateGistRootInfo, error) {
	var err error

	// get chain ID from DID
	globalStateInfo, err := ss.contractBinding.GetGISTRootInfo(&bind.CallOpts{Context: ctx}, gist)
	if err != nil {
		return abi.IStateGistRootInfo{}, err
	}
	return globalStateInfo, nil
}

// GetGistProof returns proof for global state
func (ss *StateService) GetGistProof(ctx context.Context, did *w3c.DID) (abi.IStateGistProof, error) {
	var err error
	id, err := core.IDFromDID(*did)
	if err != nil {
		return abi.IStateGistProof{}, err
	}

	gistProof, err := ss.contractBinding.GetGISTProof(&bind.CallOpts{Context: ctx}, id.BigInt())
	if err != nil {
		log.Error(ctx, "Failed to get gist proof", "error", err)
		return abi.IStateGistProof{}, err
	}

	return gistProof, nil
}
