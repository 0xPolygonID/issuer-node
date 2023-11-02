package gateways

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/iden3/contracts-abi/onchain-credential-status-resolver/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// OnChainCredStatusResolverService - on chain credential status resolver service
type OnChainCredStatusResolverService struct {
	ethClient       *eth.Client
	responseTimeout time.Duration
}

// NewOnChainCredStatusResolverService - create new instance of OnChainCredStatusResolverService
func NewOnChainCredStatusResolverService(ethClient *eth.Client, rpcTimeout time.Duration) *OnChainCredStatusResolverService {
	return &OnChainCredStatusResolverService{
		ethClient:       ethClient,
		responseTimeout: rpcTimeout,
	}
}

// GetRevocationStatus - get revocation status
func (s *OnChainCredStatusResolverService) GetRevocationStatus(ctx context.Context, state *big.Int, nonce uint64, did *w3c.DID, address ethCommon.Address) (abi.IOnchainCredentialStatusResolverCredentialStatus, error) {
	id, err := core.IDFromDID(*did)
	if err != nil {
		log.Error(ctx, "cannot get id from DID", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}
	resolver, err := abi.NewOnchainCredentialStatusResolver(address, s.ethClient.GetEthereumClient())
	if err != nil {
		log.Error(ctx, "cannot get resolver", "err", err)
		return abi.IOnchainCredentialStatusResolverCredentialStatus{}, err
	}

	revocationStatus, err := resolver.GetRevocationStatusByIdAndState(&bind.CallOpts{Context: ctx}, id.BigInt(), state, nonce)
	if err != nil {
		log.Error(ctx, "cannot get revocation status", "err", err)
		return revocationStatus, err
	}
	return revocationStatus, nil
}
