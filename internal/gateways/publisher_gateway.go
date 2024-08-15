package gateways

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	rstypes "github.com/iden3/go-rapidsnark/types"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
	"github.com/polygonid/sh-id-platform/pkg/network"
)

// PublisherEthGateway interact with blockchain
type PublisherEthGateway struct {
	rw                    *sync.RWMutex
	kms                   *kms.KMS
	publishingKeyID       kms.KeyID
	ethRPCResponseTimeout time.Duration
	networkResolver       network.Resolver
}

const rpcTimeout = 10 * time.Second

// NewPublisherEthGateway creates new instance of publishing service
func NewPublisherEthGateway(resolver network.Resolver, keyStore *kms.KMS, publishingKeyPath string) (*PublisherEthGateway, error) {
	// TODO: make timeout configurable

	return newStateService(resolver, rpcTimeout, keyStore, kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   publishingKeyPath,
	})
}

func newStateService(resolver network.Resolver, to time.Duration, kServ *kms.KMS, kPath kms.KeyID) (*PublisherEthGateway, error) {
	return &PublisherEthGateway{
		networkResolver:       resolver,
		rw:                    &sync.RWMutex{},
		kms:                   kServ,
		publishingKeyID:       kPath,
		ethRPCResponseTimeout: to,
	}, nil
}

// PublishState creates or updates state in the blockchain
func (pb *PublisherEthGateway) PublishState(ctx context.Context, identifier *w3c.DID, latestState, newState *merkletree.Hash, isOldStateGenesis bool, proof *rstypes.ProofData, identity *domain.Identity) (*string, error) {
	pb.rw.Lock()
	defer pb.rw.Unlock()

	if common.CompareMerkleTreeHash(newState, latestState) {
		return nil, errors.New("state hasn't been changed")
	}

	var tx *types.Transaction

	id, err := core.IDFromDID(*identifier)
	if err != nil {
		return nil, err
	}

	switch identity.KeyType {
	case string(kms.KeyTypeEthereum):
		keyIDs, err := pb.kms.KeysByIdentity(ctx, *identifier)
		if err != nil {
			return nil, err
		}

		var sigKeyID kms.KeyID
		for _, v := range keyIDs {
			if v.Type == kms.KeyTypeEthereum {
				sigKeyID = v
				break
			}
		}

		ctxWT, cancel := context.WithTimeout(ctx, pb.ethRPCResponseTimeout)
		defer cancel()

		client, err := getEthClient(ctx, identity, pb.networkResolver)
		if err != nil {
			log.Error(ctx, "failed to get client", "err", err)
			return nil, err
		}

		opts, err := client.CreateTxOpts(ctxWT, sigKeyID)
		if err != nil {
			log.Error(ctx, "failed to create tx opts", "err", err)
			return nil, err
		}
		log.Info(ctx, "Transaction metadata", "opts.GasPrice:", opts.GasPrice)
		log.Info(ctx, "Transaction metadata", "opts.GasLimit:", opts.GasLimit)
		log.Info(ctx, "Transaction metadata", "opts.GasTipCap:", opts.GasTipCap)

		resolverPrefix, err := identity.GetResolverPrefix()
		if err != nil {
			log.Error(ctx, "failed to get networkResolver prefix", "err", err)
			return nil, err
		}

		contractBinding, err := getContractBinding(client, resolverPrefix, pb.networkResolver)
		if err != nil {
			log.Error(ctx, "failed to get contract binding", "err", err)
			return nil, err
		}

		tx, err = contractBinding.TransitStateGeneric(opts, id.BigInt(), latestState.BigInt(), newState.BigInt(), isOldStateGenesis, big.NewInt(1), []byte{})
		if err != nil {
			return nil, err
		}

	case string(kms.KeyTypeBabyJubJub):
		ctxWT, cancel := context.WithTimeout(ctx, pb.ethRPCResponseTimeout)
		defer cancel()

		client, err := getEthClient(ctx, identity, pb.networkResolver)
		if err != nil {
			log.Error(ctx, "failed to get client", "err", err)
			return nil, err
		}

		opts, err := client.CreateTxOpts(ctxWT, pb.publishingKeyID)
		if err != nil {
			log.Error(ctx, "failed to create tx opts", "err", err)
			return nil, err
		}
		log.Info(ctx, "Transaction metadata", "opts.GasPrice:", opts.GasPrice)
		log.Info(ctx, "Transaction metadata", "opts.GasLimit:", opts.GasLimit)
		log.Info(ctx, "Transaction metadata", "opts.GasTipCap:", opts.GasTipCap)

		a, b, c, err := pb.adaptProofToAbi(proof)
		if err != nil {
			return nil, err
		}

		resolverPrefix, err := identity.GetResolverPrefix()
		if err != nil {
			log.Error(ctx, "failed to get networkResolver prefix", "err", err)
			return nil, err
		}

		contractBinding, err := getContractBinding(client, resolverPrefix, pb.networkResolver)
		if err != nil {
			log.Error(ctx, "failed to get contract binding", "err", err)
			return nil, err
		}
		tx, err = contractBinding.TransitState(opts, id.BigInt(), latestState.BigInt(), newState.BigInt(), isOldStateGenesis, a, b, c)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("unsupported key type for publishing")
	}

	txID := tx.Hash().Hex()

	return &txID, nil
}

func (pb *PublisherEthGateway) adaptProofToAbi(proof *rstypes.ProofData) (proofA [2]*big.Int, proofB [2][2]*big.Int, proofC [2]*big.Int, err error) {
	a, err := common.ArrayStringToBigInt(proof.A)
	if err != nil {
		return
	}
	b, err := common.ArrayOfStringArraysToBigInt(proof.B)
	if err != nil {
		return
	}
	c, err := common.ArrayStringToBigInt(proof.C)
	if err != nil {
		return
	}
	proofA = [2]*big.Int{a[0], a[1]}
	proofB = [2][2]*big.Int{
		{b[0][1], b[0][0]},
		{b[1][1], b[1][0]},
	}
	proofC = [2]*big.Int{c[0], c[1]}

	return
}

func getContractBinding(ethClient *eth.Client, resolverPrefix string, resolver network.Resolver) (*abi.State, error) {
	c := ethClient.GetEthereumClient()
	addr, err := resolver.GetContractAddress(resolverPrefix)
	if err != nil {
		return nil, err
	}
	binding, err := abi.NewState(*addr, c)
	if err != nil {
		return nil, err
	}

	return binding, nil
}
