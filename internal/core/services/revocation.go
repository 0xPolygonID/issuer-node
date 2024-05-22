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
	"strconv"
	"strings"
	"time"

	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	abiOnchain "github.com/iden3/contracts-abi/onchain-credential-status-resolver/go/abi"
	"github.com/iden3/contracts-abi/state/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	proofHttp "github.com/iden3/merkletree-proof/http"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	client "github.com/polygonid/sh-id-platform/pkg/http"
)

const (
	defaultRevocationTime = 30
	stateChildrenLength   = 3
	contractPartsLength   = 2
)

// ErrIdentityDoesNotExist  - identity does not exist
var ErrIdentityDoesNotExist = errors.New("identity does not exist")

// StateStore TBD
type StateStore interface {
	GetLatestStateByID(ctx context.Context, addr ethCommon.Address, id *big.Int) (abi.IStateStateInfo, error)
}

type onChainStatusService interface {
	GetRevocationStatus(ctx context.Context, state *big.Int, nonce uint64, did *w3c.DID, address ethCommon.Address) (abiOnchain.IOnchainCredentialStatusResolverCredentialStatus, error)
}

// Revocation TBD
type Revocation struct {
	stateService         ports.StateService
	onChainStatusService onChainStatusService
	contract             ethCommon.Address
}

// NewRevocationService returns the Revocation struct
func NewRevocationService(contract ethCommon.Address, stateService ports.StateService, onChainStatusService onChainStatusService) *Revocation {
	return &Revocation{
		contract:             contract,
		stateService:         stateService,
		onChainStatusService: onChainStatusService,
	}
}

// Status returns the current revocation status
func (r *Revocation) Status(ctx context.Context, credStatus interface{}, issuerDID *w3c.DID, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error) {
	status, err := convertCredentialStatus(credStatus)
	if err != nil {
		log.Error(ctx, "failed convert credential status", "error", err)
		return nil, err
	}
	switch status.Type {
	case verifiable.Iden3ReverseSparseMerkleTreeProof:
		return r.getRevocationStatusFromRHS(ctx, issuerDID, status, issuerData)
	case verifiable.SparseMerkleTreeProof:
		return getRevocationProofFromIssuer(ctx, status.ID)
	case verifiable.Iden3commRevocationStatusV1:
		return getRevocationStatusFromAgent(ctx, issuerDID.String(),
			issuerDID.String(), status.ID, status.RevocationNonce)
	case verifiable.Iden3OnchainSparseMerkleTreeProof2023:
		return r.getRevocationStatusFromOnchainCredStatusResolver(ctx, issuerDID, status)
	default:
		return nil, fmt.Errorf("%s type not supported", status.Type)
	}
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
		return nil, err
	}

	iden3commMsg, err := pkg.Pack(packers.MediaTypePlainMessage, bytesMsg, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Post(endpoint, "application/json", bytes.NewBuffer(iden3commMsg))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn(ctx, "failed to close response body: %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	basicMessage, _, err := pkg.Unpack(b)
	if err != nil {
		return nil, err
	}

	if basicMessage.Type != protocol.RevocationStatusResponseMessageType {
		return nil, fmt.Errorf("unexpected message type: %s", basicMessage.Type)
	}

	var revocationStatus protocol.RevocationStatusResponseMessageBody
	if err := json.Unmarshal(basicMessage.Body, &revocationStatus); err != nil {
		return nil, err
	}

	return &revocationStatus.RevocationStatus, nil
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
	rhsCli := proofHttp.ReverseHashCli{
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
		Issuer: verifiable.TreeState{
			State:              &s,
			ClaimsTreeRoot:     &CTR,
			RevocationTreeRoot: &RTR,
			RootOfRoots:        &RoTR,
		},
		MTP: *nonRevProof,
	}, nil
}

func (r *Revocation) getRevocationStatusFromRHS(ctx context.Context, issuerDID *w3c.DID, status verifiable.CredentialStatus, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error) {
	latestStateInfo, err := r.stateService.GetLatestStateByDID(ctx, issuerDID)
	if err != nil && strings.Contains(err.Error(), ErrIdentityDoesNotExist.Error()) {

		currentState, err := extractState(status.ID)
		if err != nil {
			log.Error(ctx, "failed extract state from rhs id", "error", err)
			return nil, err
		}
		if currentState == "" {
			return getRevocationStatusFromIssuerData(issuerDID, issuerData)
		} else {
			latestStateInfo.State, err = getGenesisState(issuerDID, currentState)
			if err != nil {
				return nil, fmt.Errorf("failed get genesis state for issuer '%s'", issuerDID)
			}
		}

	} else if err != nil {
		return nil, fmt.Errorf("failed get latest state by did '%s'", issuerDID.String())
	}

	hashedRevNonce, err := merkletree.NewHashFromBigInt(big.NewInt(int64(status.RevocationNonce)))
	if err != nil {
		return nil, fmt.Errorf("failed calculate mt hash for revocation nonce '%d': '%s'",
			status.RevocationNonce, err)
	}

	hashedIssuerState, err := merkletree.NewHashFromBigInt(latestStateInfo.State)
	if err != nil {
		return nil, fmt.Errorf("failed calcilate mt hash for issuer state '%s': '%s'",
			latestStateInfo.State, err)
	}

	u := strings.Split(status.ID, "/node")
	rs, err := getNonRevocationProofFromRHS(ctx, u[0], hashedRevNonce, hashedIssuerState)
	if err != nil && status.StatusIssuer.Type == verifiable.SparseMerkleTreeProof {
		// try to get proof from issuer
		log.Warn(ctx, "failed build revocation status from enabled RHS. Then try to fetch from issuer. RHS error: %v", err)
		revocStatus, err := getRevocationProofFromIssuer(ctx, status.StatusIssuer.ID)
		if err != nil {
			return nil, err
		}
		return revocStatus, nil
	}
	return rs, nil
}

func extractState(id string) (string, error) {
	rhsULR, err := url.Parse(id)
	if err != nil {
		return "", fmt.Errorf("invalid rhs id filed '%s'", id)
	}
	params, err := url.ParseQuery(rhsULR.RawQuery)
	if err != nil {
		return "", fmt.Errorf("invalid rhs params '%s'", rhsULR.RawQuery)
	}
	return params.Get("state"), nil
}

func getRevocationStatusFromIssuerData(did *w3c.DID, issuerData *verifiable.IssuerData) (*verifiable.RevocationStatus, error) {
	if issuerData == nil || issuerData.State.Value == nil {
		return nil, errors.New("issuer data state is empty. is not possible verify revocation status")
	}
	h, err := merkletree.NewHashFromHex(*issuerData.State.Value)
	if err != nil {
		return nil, fmt.Errorf("failed parse hex '%s'", *issuerData.State.Value)
	}
	err = common.CheckGenesisStateDID(did, h.BigInt())
	if err != nil {
		return nil, fmt.Errorf("failed check genesis state for issuer '%s'", did)
	}

	return &verifiable.RevocationStatus{
		Issuer: verifiable.TreeState{
			State:              issuerData.State.Value,
			RootOfRoots:        issuerData.State.RootOfRoots,
			ClaimsTreeRoot:     issuerData.State.ClaimsTreeRoot,
			RevocationTreeRoot: issuerData.State.RevocationTreeRoot,
		},
		MTP: merkletree.Proof{Existence: false},
	}, nil
}

func getGenesisState(did *w3c.DID, currentState string) (*big.Int, error) {
	h, err := merkletree.NewHashFromHex(currentState)
	if err != nil {
		return nil, fmt.Errorf("failed parse hex '%s'", currentState)
	}
	err = common.CheckGenesisStateDID(did, h.BigInt())
	if err != nil {
		return nil, fmt.Errorf("failed check genesis state for issuer '%s'", did)
	}
	return h.BigInt(), nil
}

func (r *Revocation) getRevocationStatusFromOnchainCredStatusResolver(ctx context.Context, issuerDID *w3c.DID, status verifiable.CredentialStatus) (*verifiable.RevocationStatus, error) {
	issuerID, err := core.IDFromDID(*issuerDID)
	if err != nil {
		return nil, fmt.Errorf("failed get issuer id from '%s'", issuerDID)
	}

	onchainRevStatus, err := newOnchainRevStatusFromURI(status.ID)
	if err != nil {
		return nil, err
	}

	if onchainRevStatus.revNonce != nil && *onchainRevStatus.revNonce != status.RevocationNonce {
		return nil, fmt.Errorf("revocationNonce is not equal to the one in OnChainCredentialStatus ID {%d} {%d}", onchainRevStatus.revNonce, status.RevocationNonce)
	}

	// get latest state from contract
	var stateToProof *big.Int
	latestStateInfo, err := r.stateService.GetLatestStateByDID(ctx, issuerDID)
	//nolint:gocritic //reason: work with errors
	if err != nil && strings.Contains(err.Error(), ErrIdentityDoesNotExist.Error()) {
		if onchainRevStatus.state == nil {
			return nil, errors.New(`latest state not found and state parameter is not present in credentialStatus.id`)
		}
		err = common.CheckGenesisStateDID(issuerDID, onchainRevStatus.state)
		if err != nil {
			return nil, err
		}
		stateToProof = onchainRevStatus.state

	} else if err != nil {
		return nil, fmt.Errorf("failed get latest state by id '%s'", issuerID)
	} else {
		stateToProof = latestStateInfo.State
	}

	rs, err := r.onChainStatusService.GetRevocationStatus(
		ctx, stateToProof, status.RevocationNonce, issuerDID, onchainRevStatus.contractAddress,
	)
	if err != nil {
		return nil, errors.New("failed get revocation status from onchain cred status resolver")
	}

	smProof, err := common.SmartContractProofToMtProofAdapter(common.SmartContractProof{
		Root:         rs.Mtp.Root,
		Existence:    rs.Mtp.Existence,
		Siblings:     rs.Mtp.Siblings,
		Index:        rs.Mtp.Index,
		Value:        rs.Mtp.Value,
		AuxExistence: rs.Mtp.AuxExistence,
		AuxIndex:     rs.Mtp.AuxIndex,
		AuxValue:     rs.Mtp.AuxValue,
	})
	if err != nil {
		log.Error(ctx, "failed convert smart contract proof to merkle tree proof", "error", err)
		return nil, err
	}

	state, err := merkletree.NewHashFromBigInt(rs.Issuer.State)
	if err != nil {
		log.Error(ctx, "failed convert state to merkle tree hash", "error", err)
		return nil, err
	}
	stateHex := state.Hex()
	ctr, err := merkletree.NewHashFromBigInt(rs.Issuer.ClaimsTreeRoot)
	if err != nil {
		return nil, err
	}
	ctrHex := ctr.Hex()

	rtr, err := merkletree.NewHashFromBigInt(rs.Issuer.RevocationTreeRoot)
	if err != nil {
		log.Error(ctx, "failed convert revocation tree root to merkle tree hash", "error", err)
		return nil, err
	}
	rtrHex := rtr.Hex()

	ror, err := merkletree.NewHashFromBigInt(rs.Issuer.RootOfRoots)
	if err != nil {
		log.Error(ctx, "failed convert root of roots to merkle tree hash", "error", err)
		return nil, err
	}
	rorHex := ror.Hex()

	return &verifiable.RevocationStatus{
		Issuer: verifiable.TreeState{
			State:              &stateHex,
			ClaimsTreeRoot:     &ctrHex,
			RevocationTreeRoot: &rtrHex,
			RootOfRoots:        &rorHex,
		},
		MTP: *smProof,
	}, nil
}

type onchainRevStatus struct {
	contractAddress ethCommon.Address
	revNonce        *uint64
	state           *big.Int
}

func newOnchainRevStatusFromURI(id string) (onchainRevStatus, error) {
	var s onchainRevStatus

	uri, err := url.Parse(id)
	if err != nil {
		return s, errors.New("OnChainCredentialStatus ID is not a valid URI")
	}

	contract := uri.Query().Get("contractAddress")
	if contract == "" {
		return s, errors.New("OnChainCredentialStatus contract address is empty")
	}

	contractParts := strings.Split(contract, ":")
	if len(contractParts) != contractPartsLength {
		return s, errors.New(
			"OnChainCredentialStatus contract address is not valid")
	}
	if !ethCommon.IsHexAddress(contractParts[1]) {
		return s, errors.New(
			"OnChainCredentialStatus incorrect contract address")
	}
	s.contractAddress = ethCommon.HexToAddress(contractParts[1])

	revocationNonce := uri.Query().Get("revocationNonce")
	// revnonce may be nil if params is absent in query
	if revocationNonce != "" {
		n, err := strconv.ParseUint(revocationNonce, 10, 64)
		if err != nil {
			return s, errors.New("revocationNonce is not a number in OnChainCredentialStatus ID")
		}
		s.revNonce = &n
	}
	// state may be nil if params is absent in query
	stateParam := uri.Query().Get("state")
	if stateParam == "" {
		s.state = nil
	} else {
		stateHash, err := merkletree.NewHashFromHex(stateParam)
		if err != nil {
			return s, err
		}
		s.state = stateHash.BigInt()
	}

	return s, nil
}
