package gateways

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"sync"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/blockchain/eth"
)

// PublisherEthGateway interact with blockchain
type PublisherEthGateway struct {
	rw              *sync.RWMutex
	client          *eth.Client
	contract        ethCommon.Address
	kms             *kms.KMS
	publishingKeyID kms.KeyID
}

// NewPublisherEthGateway creates new instance of publishing service
func NewPublisherEthGateway(_client *eth.Client, contract ethCommon.Address, keyStore *kms.KMS, publishingKeyPath string) (*PublisherEthGateway, error) {
	if publishingKeyPath == "" {
		return nil, errors.New("publishing key path is required")
	}
	return &PublisherEthGateway{
		client:   _client,
		contract: contract,
		rw:       &sync.RWMutex{},
		kms:      keyStore,
		publishingKeyID: kms.KeyID{
			Type: kms.KeyTypeEthereum,
			ID:   publishingKeyPath,
		},
	}, nil
}

// PublishState creates or updates state in the blockchain
func (pb *PublisherEthGateway) PublishState(ctx context.Context, identifier *core.DID, latestState, newState *merkletree.Hash, isOldStateGenesis bool, proof *domain.ZKProof) (*string, error) {
	pb.rw.Lock()
	defer pb.rw.Unlock()

	if common.CompareMerkleTreeHash(newState, latestState) {
		return nil, errors.New("state hasn't been changed")
	}

	fromAddress, err := pb.getAddressForTxInitiator()
	if err != nil {
		return nil, err
	}

	payload, err := pb.getStatePayload(identifier, latestState, newState, isOldStateGenesis, proof)
	if err != nil {
		return nil, err
	}

	txParams := eth.TransactionParams{
		FromAddress: fromAddress,
		ToAddress:   pb.contract,
		Payload:     payload,
	}
	tx, err := pb.client.CreateRawTx(ctx, txParams)
	if err != nil {
		return nil, err
	}

	cid, err := pb.client.ChainID(ctx)
	if err != nil {
		return nil, err
	}

	s := types.LatestSignerForChainID(cid)

	h := s.Hash(tx)
	sig, err := pb.kms.Sign(ctx, pb.publishingKeyID, h[:])
	if err != nil {
		return nil, err
	}

	signedTx, err := tx.WithSignature(s, sig)
	if err != nil {
		return nil, fmt.Errorf("failed sign transaction: %w", err)
	}

	err = pb.client.SendRawTx(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	txID := signedTx.Hash().Hex()

	var (
		gasTip            = signedTx.GasTipCap()
		maxGasPricePerFee = signedTx.GasFeeCap()
		baseFee           = big.NewInt(0).Sub(maxGasPricePerFee, gasTip)
	)
	log.Debug(ctx, "Prices for tx '%s' Basefee: %s; Tip: %s; MaxPrice: %s", txID, baseFee, gasTip, maxGasPricePerFee)
	return &txID, nil
}

func (pb *PublisherEthGateway) getAddressForTxInitiator() (ethCommon.Address, error) {
	bytesPubKey, err := pb.kms.PublicKey(pb.publishingKeyID)
	if err != nil {
		return ethCommon.Address{}, err
	}
	var pubKey *ecdsa.PublicKey
	bytesPubKeyLen := 33
	switch len(bytesPubKey) {
	case bytesPubKeyLen:
		pubKey, err = crypto.DecompressPubkey(bytesPubKey)
	default:
		pubKey, err = crypto.UnmarshalPubkey(bytesPubKey)
	}
	if err != nil {
		return ethCommon.Address{}, err
	}
	fromAddress := crypto.PubkeyToAddress(*pubKey)
	return fromAddress, nil
}

func (pb *PublisherEthGateway) getStatePayload(identifier *core.DID, latestState, newState *merkletree.Hash, isOldStateGenesis bool, proof *domain.ZKProof) ([]byte, error) {
	a, b, c, err := proof.ProofToBigInts()
	if err != nil {
		return nil, err
	}
	proofA := [2]*big.Int{a[0], a[1]}
	proofB := [2][2]*big.Int{
		{b[0][1], b[0][0]},
		{b[1][1], b[1][0]},
	}
	proofC := [2]*big.Int{c[0], c[1]}

	ab, err := eth.StateMetaData.GetAbi()
	if err != nil {
		return nil, err
	}

	payload, err := ab.Pack("transitState", identifier.ID.BigInt(), latestState.BigInt(), newState.BigInt(), isOldStateGenesis,
		proofA, proofB, proofC)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
