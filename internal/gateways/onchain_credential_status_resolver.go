package gateways

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/contracts-abi/onchain-credential-status-resolver/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/network"
)

// OnChainCredStatusResolverService - on chain credential status networkResolver service
type OnChainCredStatusResolverService struct {
	networkResolver *network.Resolver
}

// NewOnChainCredStatusResolverService - create new instance of OnChainCredStatusResolverService
func NewOnChainCredStatusResolverService(networkResolver *network.Resolver) *OnChainCredStatusResolverService {
	return &OnChainCredStatusResolverService{
		networkResolver: networkResolver,
	}
}

// GetRevocationStatus - get revocation status
func (s *OnChainCredStatusResolverService) GetRevocationStatus(ctx context.Context, state *big.Int, nonce uint64, did *w3c.DID, address ethCommon.Address) (abi.IOnchainCredentialStatusResolverCredentialStatus, error) {
	id, err := core.IDFromDID(*did)
	if err != nil {
		log.Error(ctx, "cannot get id from DID", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}

	resolverPrefix, err := common.ResolverPrefix(did)
	if err != nil {
		log.Error(ctx, "cannot get networkResolver prefix", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}

	client, err := s.networkResolver.GetEthClient(resolverPrefix)
	if err != nil {
		log.Error(ctx, "cannot get eth client", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}

	resolver, err := abi.NewOnchainCredentialStatusResolver(address, client.GetEthereumClient())
	if err != nil {
		log.Error(ctx, "cannot get networkResolver", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}

	revocationStatus, err := resolver.GetRevocationStatusByIdAndState(&bind.CallOpts{Context: ctx}, id.BigInt(), state, nonce)
	if err != nil {
		log.Error(ctx, "cannot get revocation status", "err", err)
		return revocationStatus, err
	}
	return revocationStatus, nil
}
