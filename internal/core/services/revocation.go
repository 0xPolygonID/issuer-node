package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	proof "github.com/iden3/merkletree-proof"

	comm "github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/log"
	client "github.com/polygonid/sh-id-platform/pkg/http"
)

const (
	defaultRevocationTime = 30
	stateChildrenLength   = 3
)

// ErrStateNotFound issuer state is genesis state.
var ErrStateNotFound = errors.New("Identity does not exist")

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
func (r *Revocation) Status(ctx context.Context, credStatus interface{}, userDID, issuerDID *w3c.DID, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error) {
	status, err := convertCredentialStatus(credStatus)
	if err != nil {
		return nil, err
	}

	switch status.Type {
	case verifiable.Iden3ReverseSparseMerkleTreeProof:
		issuerID, err := core.IDFromDID(*issuerDID)
		if err != nil {
			log.Error(ctx, "failed get issuer id from", "issuerDID", issuerDID)
			return nil, errors.Join(err, errors.New("failed get issuer id"))
		}

		latestStateInfo, err := r.eth.GetLatestStateByID(ctx, r.contract, issuerID.BigInt())
		if err != nil && strings.Contains(err.Error(), ErrStateNotFound.Error()) {

			currentState, err := extractState(status.ID)
			if err != nil {
				return nil, err
			}
			if currentState == "" {
				return getRevocationStatusFromIssuerData(ctx, issuerDID, issuerData)
			} else {
				latestStateInfo.State, err = getGenesisState(issuerDID, currentState)
				if err != nil {
					return nil, errors.Join(err, fmt.Errorf("failed get genesis state for issuer '%s'", issuerDID))
				}
			}

		} else if err != nil {
			return nil, errors.Join(err, fmt.Errorf("failed get latest state by id '%s'", issuerID))
		}

		hashedRevNonce, err := merkletree.NewHashFromBigInt(big.NewInt(int64(status.RevocationNonce)))
		if err != nil {
			log.Error(ctx, "failed calculate mt hash for revocation nonce", "revocationNonce", status.RevocationNonce)
			return nil, errors.Join(err, fmt.Errorf("failed calculate mt hash for revocation nonce '%d'", status.RevocationNonce))
		}

		hashedIssuerRoot, err := merkletree.NewHashFromBigInt(latestStateInfo.State)
		if err != nil {
			return nil, fmt.Errorf("failed calcilate mt hash for issuer state '%s'", latestStateInfo.State)
		}

		u := strings.Split(status.ID, "/node")
		rs, err := getNonRevocationProofFromRHS(ctx, u[0], hashedRevNonce, hashedIssuerRoot)
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
	case verifiable.SparseMerkleTreeProof:
		return getRevocationProofFromIssuer(ctx, status.ID)
	case verifiable.Iden3commRevocationStatusV1:
		return getRevocationStatusFromAgent(ctx, userDID.String(), issuerDID.String(), status.ID, status.RevocationNonce)
	default:
		return nil, fmt.Errorf("%s type not supported", status.Type)
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

func convertCredentialStatus(credStatus interface{}) (verifiable.CredentialStatus, error) {
	status, ok := credStatus.(verifiable.CredentialStatus)
	if ok {
		return status, nil
	}
	pointedStatus, ok := credStatus.(*verifiable.CredentialStatus)
	if ok {
		return *pointedStatus, nil
	}
	_, ok = credStatus.(map[string]interface{})
	if ok {
		b, err := json.Marshal(credStatus)
		if err != nil {
			return verifiable.CredentialStatus{}, err
		}
		var status verifiable.CredentialStatus
		err = json.Unmarshal(b, &status)
		if err != nil {
			return verifiable.CredentialStatus{}, err
		}
		return status, nil
	}

	return verifiable.CredentialStatus{}, errors.New("failed cast credential status to verifiable.CredentialStatus")
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

func getRevocationStatusFromIssuerData(ctx context.Context, did *w3c.DID, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error) {
	if issuerData == nil || issuerData.State.Value == nil {
		log.Error(ctx, "issuer data state is empty. is not possible verify revocation status")
		return nil, errors.New("issuer data state is empty. is not possible verify revocation status")
	}
	h, err := merkletree.NewHashFromHex(*issuerData.State.Value)
	if err != nil {
		log.Error(ctx, "failed parse hex", "state", *issuerData.State.Value)
		return nil, errors.Join(err, fmt.Errorf("failed parse hex '%s'", *issuerData.State.Value))
	}
	err = comm.CheckGenesisStateDID(did, h.BigInt())
	if err != nil {
		log.Error(ctx, "failed check genesis state for issuer", "did", did, "state", h.BigInt())
		return nil, errors.Join(err, fmt.Errorf("failed check genesis state for issuer '%s'", did))
	}

	return &verifiable.RevocationStatus{
		Issuer: struct {
			State              *string `json:"state,omitempty"`
			RootOfRoots        *string `json:"rootOfRoots,omitempty"`
			ClaimsTreeRoot     *string `json:"claimsTreeRoot,omitempty"`
			RevocationTreeRoot *string `json:"revocationTreeRoot,omitempty"`
		}{
			State:              issuerData.State.Value,
			RootOfRoots:        issuerData.State.RootOfRoots,
			ClaimsTreeRoot:     issuerData.State.ClaimsTreeRoot,
			RevocationTreeRoot: issuerData.State.RevocationTreeRoot,
		},
		MTP: merkletree.Proof{Existence: false},
	}, nil
}

func extractState(id string) (string, error) {
	rhsULR, err := url.Parse(id)
	if err != nil {
		return "", errors.Join(err, errors.New("failed parse rhs url"))
	}
	params, err := url.ParseQuery(rhsULR.RawQuery)
	if err != nil {
		return "", errors.Join(err, errors.New("failed parse rhs params"))
	}
	return params.Get("state"), nil
}

func getGenesisState(did *w3c.DID, currentState string) (*big.Int, error) {
	h, err := merkletree.NewHashFromHex(currentState)
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed parse hex '%s'", currentState))
	}
	err = comm.CheckGenesisStateDID(did, h.BigInt())
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("failed check genesis state for issuer '%s'", did))
	}
	return h.BigInt(), nil
}

func getRevocationStatusFromAgent(ctx context.Context, from, to, endpoint string, nonce uint64) (*verifiable.RevocationStatus, error) {
	pkg := iden3comm.NewPackageManager()
	if err := pkg.RegisterPackers(&packers.PlainMessagePacker{}); err != nil {
		return nil, err
	}

	revocationBody := protocol.RevocationStatusRequestMessageBody{
		RevocationNonce: nonce,
	}
	rawBody, err := json.Marshal(revocationBody)
	if err != nil {
		return nil, err
	}

	// TODO(illia-korotia): create simple message builder?
	msg := iden3comm.BasicMessage{
		ID:       uuid.New().String(),
		ThreadID: uuid.New().String(),
		From:     from,
		To:       to,
		Type:     protocol.RevocationStatusRequestMessageType,
		Body:     rawBody,
	}
	bytesMsg, err := json.Marshal(msg)
	if err != nil {
		log.Error(ctx, "failed marshal message", "message", msg, "err", err)
		return nil, err
	}

	iden3commMsg, err := pkg.Pack(packers.MediaTypePlainMessage, bytesMsg, nil)
	if err != nil {
		log.Error(ctx, "failed pack message", "message", msg, "err", err)
		return nil, err
	}

	resp, err := http.DefaultClient.Post(endpoint, "application/json", bytes.NewBuffer(iden3commMsg))
	if err != nil {
		log.Error(ctx, "failed send request", "endpoint", endpoint, "err", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn(ctx, "failed close response body", "err", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Error(ctx, "bad status code", "statusCode", resp.StatusCode)
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	basicMessage, _, err := pkg.Unpack(b)
	if err != nil {
		log.Error(ctx, "failed unpack message", "err", err)
		return nil, err
	}

	if basicMessage.Type != protocol.RevocationStatusResponseMessageType {
		return nil, fmt.Errorf("unexpected message type: %s", basicMessage.Type)
	}

	var revocationStatus protocol.RevocationStatusResponseMessageBody
	if err := json.Unmarshal(basicMessage.Body, &revocationStatus); err != nil {
		log.Error(ctx, "failed unmarshal message", "err", err)
		return nil, err
	}

	return &revocationStatus.RevocationStatus, nil
}
