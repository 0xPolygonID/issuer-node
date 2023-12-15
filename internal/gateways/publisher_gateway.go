package gateways

import (
	"context"
	"errors"
	"math/big"
	"sync"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	rstypes "github.com/iden3/go-rapidsnark/types"

	"github.com/polygonid/issuer-node/internal/common"
	"github.com/polygonid/issuer-node/internal/core/domain"
	"github.com/polygonid/issuer-node/internal/kms"
	"github.com/polygonid/issuer-node/internal/log"
	"github.com/polygonid/issuer-node/pkg/blockchain/eth"
)

// PublisherEthGateway interact with blockchain
type PublisherEthGateway struct {
	rw                    *sync.RWMutex
	client                *eth.Client
	kms                   *kms.KMS
	publishingKeyID       kms.KeyID
	ethRPCResponseTimeout time.Duration
	contractBinding       *abi.State
}

const rpcTimeout = 10 * time.Second

// NewPublisherEthGateway creates new instance of publishing service
func NewPublisherEthGateway(_client *eth.Client, contract ethCommon.Address, keyStore *kms.KMS, publishingKeyPath string) (*PublisherEthGateway, error) {
	// TODO: make timeout configurable
	return newStateService(_client, contract, rpcTimeout, keyStore, kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   publishingKeyPath,
	})
}

func newStateService(client *eth.Client, addr ethCommon.Address, to time.Duration, kServ *kms.KMS, kPath kms.KeyID) (*PublisherEthGateway, error) {
	c := client.GetEthereumClient()

	binding, err := abi.NewState(addr, c)
	if err != nil {
		return nil, err
	}

	return &PublisherEthGateway{
		client:                client,
		contractBinding:       binding,
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
		opts, err := pb.client.CreateTxOpts(ctxWT, sigKeyID)
		if err != nil {
			log.Error(ctx, "failed to create tx opts", "err", err)
			return nil, err
		}

		tx, err = pb.contractBinding.TransitStateGeneric(opts, id.BigInt(), latestState.BigInt(), newState.BigInt(), isOldStateGenesis, big.NewInt(1), []byte{})
		if err != nil {
			return nil, err
		}

	case string(kms.KeyTypeBabyJubJub):
		ctxWT, cancel := context.WithTimeout(ctx, pb.ethRPCResponseTimeout)
		defer cancel()
		opts, err := pb.client.CreateTxOpts(ctxWT, pb.publishingKeyID)
		if err != nil {
			log.Error(ctx, "failed to create tx opts", "err", err)
			return nil, err
		}

		a, b, c, err := pb.adaptProofToAbi(proof)
		if err != nil {
			return nil, err
		}

		tx, err = pb.contractBinding.TransitState(opts, id.BigInt(), latestState.BigInt(), newState.BigInt(), isOldStateGenesis, a, b, c)
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
