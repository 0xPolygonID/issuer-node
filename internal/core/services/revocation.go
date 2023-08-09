package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/verifiable"
	proof "github.com/iden3/merkletree-proof"

	"github.com/polygonid/sh-id-platform/internal/log"
	client "github.com/polygonid/sh-id-platform/pkg/http"
	"github.com/polygonid/sh-id-platform/pkg/protocol"
)

const (
	defaultRevocationTime = 30
	stateChildrenLength   = 3
)

// StateStore TBD
type StateStore interface {
	GetLatestStateByID(ctx context.Context, addr common.Address, id *big.Int) (abi.IStateStateInfo, error)
}

// Revocation TBD
type Revocation struct {
	eth      StateStore
	contract common.Address
}

// NewRevocationService returns the Revocation struct
func NewRevocationService(ethStore StateStore, contract common.Address) *Revocation {
	return &Revocation{
		eth:      ethStore,
		contract: contract,
	}
}

// Status returns the current revocation status
func (r *Revocation) Status(ctx context.Context, credStatus interface{}, issuerDID *core.DID) (*verifiable.RevocationStatus, error) {
	switch status := credStatus.(type) {
	case *verifiable.RHSCredentialStatus:
		latestStateInfo, err := r.eth.GetLatestStateByID(ctx, r.contract, issuerDID.ID.BigInt())
		if err != nil && strings.Contains(err.Error(), protocol.ErrStateNotFound.Error()) {
			return nil, protocol.ErrStateNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("failed get latest state for '%s': %v", issuerDID, err)
		}

		hashedRevNonce, err := merkletree.NewHashFromBigInt(big.NewInt(int64(status.RevocationNonce)))
		if err != nil {
			return nil, fmt.Errorf("failed calculate mt hash for revocation nonce '%d': '%s'",
				status.RevocationNonce, err)
		}

		hashedIssuerRoot, err := merkletree.NewHashFromBigInt(latestStateInfo.State)
		if err != nil {
			return nil, fmt.Errorf("failed calcilate mt hash for issuer state '%s': '%s'",
				latestStateInfo.State, err)
		}

		rs, err := getNonRevocationProofFromRHS(ctx, status.ID, hashedRevNonce, hashedIssuerRoot)
		if err != nil && status.StatusIssuer.Type == verifiable.SparseMerkleTreeProof {
			// try to get proof from issuer
			log.Warn(ctx, "failed build revocation status from enabled RHS. Then try to fetch from issuer")
			revocStatus, err := getRevocationProofFromIssuer(ctx, status.StatusIssuer.ID)
			if err != nil {
				return nil, err
			}
			return revocStatus, nil
		}
		return rs, nil
	case *verifiable.CredentialStatus:
		return getRevocationProofFromIssuer(ctx, status.ID)
	case verifiable.RHSCredentialStatus:
		return r.Status(ctx, &status, issuerDID)
	case verifiable.CredentialStatus:
		return r.Status(ctx, &status, issuerDID)
	case map[string]interface{}:
		credStatusType, ok := status["type"].(string)
		if !ok {
			return nil, errors.New("credential status doesn't contain type")
		}
		marshaledStatus, err := json.Marshal(status)
		if err != nil {
			return nil, err
		}
		var s interface{}
		switch verifiable.CredentialStatusType(credStatusType) {
		case verifiable.Iden3ReverseSparseMerkleTreeProof:
			s = &verifiable.RHSCredentialStatus{}
		case verifiable.SparseMerkleTreeProof:
			s = &verifiable.CredentialStatus{}
		default:
			return nil, fmt.Errorf("credential status type %s id not supported", credStatusType)
		}

		err = json.Unmarshal(marshaledStatus, s)
		if err != nil {
			return nil, err
		}
		return r.Status(ctx, s, issuerDID)

	default:
		return nil, errors.New("unknown credential status format")
	}
}

func getRevocationProofFromIssuer(ctx context.Context, url string) (*verifiable.RevocationStatus, error) {
	b, err := client.NewClient(*http.DefaultClient).Get(ctx, url)
	if err != nil {
		return nil, err
	}

	rs := &verifiable.RevocationStatus{}
	if err := json.Unmarshal(b, rs); err != nil {
		return nil, err
	}
	return rs, nil
}

func getNonRevocationProofFromRHS(ctx context.Context, rhsURL string, data, issuerRoot *merkletree.Hash) (*verifiable.RevocationStatus, error) {
	rhsCli := proof.HTTPReverseHashCli{
		URL:         rhsURL,
		HTTPTimeout: time.Second * defaultRevocationTime,
	}

	treeRoots, err := rhsCli.GetNode(ctx, issuerRoot)
	if err != nil {
		return nil, err
	}
	if len(treeRoots.Children) != stateChildrenLength {
		return nil, fmt.Errorf("state should has tree children")
	}

	var (
		s    = issuerRoot.Hex()
		CTR  = treeRoots.Children[0].Hex()
		RTR  = treeRoots.Children[1].Hex()
		RoTR = treeRoots.Children[2].Hex()
	)

	rtrHashed, err := merkletree.NewHashFromString(RTR)
	if err != nil {
		return nil, err
	}
	nonRevProof, err := rhsCli.GenerateProof(ctx, rtrHashed, data)
	if err != nil {
		return nil, fmt.Errorf("failed generate proof for root '%s' and element '%s': %s",
			issuerRoot, data, err)
	}

	return &verifiable.RevocationStatus{
		Issuer: struct {
			State              *string `json:"state,omitempty"`
			RootOfRoots        *string `json:"rootOfRoots,omitempty"`
			ClaimsTreeRoot     *string `json:"claimsTreeRoot,omitempty"`
			RevocationTreeRoot *string `json:"revocationTreeRoot,omitempty"`
		}{
			State:              &s,
			ClaimsTreeRoot:     &CTR,
			RevocationTreeRoot: &RTR,
			RootOfRoots:        &RoTR,
		},
		MTP: *nonRevProof,
	}, nil
}
