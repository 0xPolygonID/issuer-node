package services

import (
	"context"
	"math/big"

	ethCommon "github.com/ethereum/go-ethereum/common"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/network"
)

// AccountService is a service for account operations
type AccountService struct {
	networkResolver network.Resolver
}

// NewAccountService creates a new instance of AccountService
func NewAccountService(networkResolver network.Resolver) *AccountService {
	return &AccountService{
		networkResolver: networkResolver,
	}
}

// GetBalanceByDID returns balance by DID
func (as *AccountService) GetBalanceByDID(ctx context.Context, did *w3c.DID) (*big.Int, error) {
	id, err := core.IDFromDID(*did)
	if err != nil {
		log.Error(ctx, "cannot get id from DID", "err", err)
		return nil, err
	}

	resolverPrefix, err := common.ResolverPrefix(did)
	if err != nil {
		log.Error(ctx, "cannot get networkResolver prefix", "err", err)
		return nil, err
	}

	ethClient, err := as.networkResolver.GetEthClient(resolverPrefix)
	if err != nil {
		log.Error(ctx, "cannot get eth client", "err", err)
		return nil, err
	}
	ethAddress, err := core.EthAddressFromID(id)
	if err != nil {
		log.Error(ctx, "cannot get eth address from id", "err", err)
		return nil, err
	}
	commonAddress := ethCommon.BytesToAddress(ethAddress[:])
	return ethClient.BalanceAt(ctx, commonAddress)
}
