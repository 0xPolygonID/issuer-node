package gateways

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/iden3/go-circuits"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/poseidon"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
)

type jobIDType string

const jobID jobIDType = "job-id"

// PublisherGateway - Define the interface for publishers.
type PublisherGateway interface {
	PublishState(ctx context.Context, identifier *core.DID, latestState *merkletree.Hash, newState *merkletree.Hash, isOldStateGenesis bool, proof *domain.ZKProof) (*string, error)
}

type publisher struct {
	storage             *db.Storage
	identityService     ports.IdentityService
	claimService        ports.ClaimsService
	mtService           ports.MtService
	kms                 kms.KMSType
	transactionService  ports.TransactionService
	confirmationTimeout time.Duration
	zkService           ports.ZKGenerator
	publisherGateway    PublisherGateway
}

// NewPublisher - Constructor
func NewPublisher(storage *db.Storage, identityService ports.IdentityService, claimService ports.ClaimsService, mtService ports.MtService, kms kms.KMSType, transactionService ports.TransactionService, zkService ports.ZKGenerator, publisherGateway PublisherGateway, confirmationTimeout time.Duration) *publisher {
	return &publisher{
		identityService:     identityService,
		claimService:        claimService,
		storage:             storage,
		mtService:           mtService,
		kms:                 kms,
		transactionService:  transactionService,
		zkService:           zkService,
		publisherGateway:    publisherGateway,
		confirmationTimeout: confirmationTimeout,
	}
}

func (p *publisher) PublishState(ctx context.Context) {
	jobIDValue, err := uuid.NewUUID()
	if err != nil {
		log.Error(ctx, "error", err)
		return
	}
	ctx = context.WithValue(ctx, jobID, jobIDValue.String())
	log.Info(ctx, "publish state job started", jobID, jobIDValue.String())
	// TODO: make snapshot
	// make snapshot if rds was init

	// 1. Get all issuers that have claims not included in any state
	issuers, err := p.identityService.GetUnprocessedIssuersIDs(ctx)
	if err != nil {
		log.Error(ctx, "error fetching unprocessed issuers dids", err)
		return
	}
	log.Info(ctx, "GetUnprocessedIssuersIDs", "GetUnprocessedIssuersIDs-len", len(issuers))

	// 2. Get all states that were not transacted by some reason
	states, err := p.identityService.GetNonTransactedStates(ctx)
	if err != nil {
		log.Error(ctx, "error fetching non transacted states", err)
		return
	}

	// 3. Publish non -transacted states
	log.Info(ctx, "Non transacted states", "states len", len(states))
	for i := range states {
		err = p.publishProof(ctx, states[i])
		if err != nil {
			log.Error(ctx, "Error during publishing proof", err, "did", states[i].Identifier)
			continue
		}
	}

	// we shouldn't process IDs which had unpublished states.

	toCalculateAndPublish := []*core.DID{}
	for _, id := range issuers {
		if !domain.ContainsID(states, id) {
			toCalculateAndPublish = append(toCalculateAndPublish, id)
		}
	}

	// 4. Calculate new states and publish them synchronously
	log.Info(ctx, "toCalculateAndPublish", "toCalculateAndPublish-len", len(toCalculateAndPublish))
	for _, id := range toCalculateAndPublish {
		state, err := p.identityService.UpdateState(ctx, id)
		if err != nil {
			log.Error(ctx, "Error during processing claims", err, "did", id.String())
			continue
		}

		err = p.publishProof(ctx, *state)
		if err != nil {
			log.Error(ctx, "Error during publishing proof:", err, "did", id.String())
			continue
		}
	}

	log.Info(ctx, "publish state job finished", jobID, jobIDValue.String())
}

// PublishProof publishes new proof using the latest state
func (p *publisher) publishProof(ctx context.Context, newState domain.IdentityState) error {
	did, err := core.ParseDID(newState.Identifier)
	if err != nil {
		return err
	}

	// 1. Get latest transacted state
	latestState, err := p.identityService.GetLatestStateByID(ctx, did)
	if err != nil {
		return err
	}

	latestStateHash, err := merkletree.NewHashFromHex(*latestState.State)
	if err != nil {
		return err
	}

	// TODO: core.IdenState should be calculated before state stored to db
	newStateHash, err := merkletree.NewHashFromHex(*newState.State)
	if err != nil {
		return err
	}

	newTreeState, err := newState.ToTreeState()
	if err != nil {
		return err
	}

	authClaim, err := p.claimService.GetAuthClaimForPublishing(ctx, did, *newState.State)
	if err != nil {
		return err
	}

	claimKeyID, err := p.identityService.GetKeyIDFromAuthClaim(ctx, authClaim)
	if err != nil {
		return err
	}

	oldTreeState, err := latestState.ToTreeState()
	if err != nil {
		return err
	}

	circuitAuthClaim, circuitAuthClaimNewStateIncProof, err := p.fillAuthClaimData(ctx, did, authClaim, newState)
	if err != nil {
		return err
	}

	hashOldAndNewStates, err := poseidon.Hash([]*big.Int{oldTreeState.State.BigInt(), newStateHash.BigInt()})
	if err != nil {
		return err
	}

	sigDigest := kms.BJJDigest(hashOldAndNewStates)
	sigBytes, err := p.kms.Sign(ctx, claimKeyID, sigDigest)
	if err != nil {
		return err
	}

	signature, err := kms.DecodeBJJSignature(sigBytes)
	if err != nil {
		return err
	}

	isLatestStateGenesis := latestState.PreviousState == nil
	stateTransitionInputs := circuits.StateTransitionInputs{
		ID:                &did.ID,
		NewTreeState:      newTreeState,
		OldTreeState:      oldTreeState,
		IsOldStateGenesis: isLatestStateGenesis,

		AuthClaim:          circuitAuthClaim.Claim,
		AuthClaimIncMtp:    circuitAuthClaim.IncProof.Proof,
		AuthClaimNonRevMtp: circuitAuthClaim.NonRevProof.Proof,

		AuthClaimNewStateIncMtp: circuitAuthClaimNewStateIncProof,
		Signature:               signature,
	}

	jsonInputs, err := stateTransitionInputs.InputsMarshal()
	if err != nil {
		return err
	}

	// TODO: Integrate when it's finished
	fullProof, err := p.zkService.Generate(ctx, jsonInputs, string(circuits.StateTransitionCircuitID))
	if err != nil {
		return err
	}

	// 7. Publish state and receive txID

	txID, err := p.publisherGateway.PublishState(ctx, did, latestStateHash, newStateHash, isLatestStateGenesis, fullProof.Proof)
	if err != nil {
		return err
	}

	log.Info(ctx, "Success!", "TxID", txID)

	// 8. Update state with txID value (block values are still default because tx is not confirmed)

	newState.Status = domain.StatusTransacted
	newState.TxID = txID

	err = p.identityService.UpdateIdentityState(ctx, &newState)
	if err != nil {
		return err
	}

	// add go routine that will listen for transaction status update

	go func() {
		err2 := p.updateTransactionStatus(ctx, newState, *txID)
		if err2 != nil {
			log.Error(ctx, "can not update transaction status", err2)
		}
	}()

	return nil
}

func (p *publisher) fillAuthClaimData(ctx context.Context, identifier *core.DID, authClaim *domain.Claim, newState domain.IdentityState) (
	authClaimData *circuits.ClaimWithMTPProof, authClaimNewStateIncProof *merkletree.Proof, err error,
) {
	err = p.storage.Pgx.BeginFunc(
		ctx, func(tx pgx.Tx) error {
			var errIn error

			var idState *domain.IdentityState
			idState, errIn = p.identityService.GetLatestStateByID(ctx, identifier)
			if errIn != nil {
				return errIn
			}

			identityTrees, errIn := p.mtService.GetIdentityMerkleTrees(ctx, tx, identifier)
			if errIn != nil {
				return errIn
			}

			claimsTree, errIn := identityTrees.ClaimsTree()
			if errIn != nil {
				return errIn
			}
			// get index hash of authClaim
			coreClaim := authClaim.CoreClaim.Get()
			hIndex, errIn := coreClaim.HIndex()
			if errIn != nil {
				return errIn
			}

			authClaimMTP, _, errIn := claimsTree.GenerateProof(ctx, hIndex, idState.TreeState().ClaimsRoot)
			if errIn != nil {
				return errIn
			}

			authClaimData = &circuits.ClaimWithMTPProof{
				Claim: coreClaim,
			}

			authClaimData.IncProof = circuits.MTProof{
				Proof:     authClaimMTP,
				TreeState: idState.TreeState(),
			}

			// revocation / non revocation MTP for the latest identity state
			nonRevocationProof, errIn := identityTrees.
				GenerateRevocationProof(ctx, new(big.Int).SetUint64(uint64(authClaim.RevNonce)), idState.TreeState().RevocationRoot)
			if errIn != nil {
				return errIn
			}

			authClaimData.NonRevProof = circuits.MTProof{
				TreeState: idState.TreeState(),
				Proof:     nonRevocationProof,
			}

			// proof that auth key is included in new state claims tree
			authClaimNewStateIncProof, _, errIn = claimsTree.GenerateProof(ctx, hIndex, newState.TreeState().ClaimsRoot)
			if errIn != nil {
				return errIn
			}

			return errIn
		})

	if err != nil {
		return nil, nil, err
	}
	return authClaimData, authClaimNewStateIncProof, nil
}

// updateTransactionStatus update identity state with transaction status
func (p *publisher) updateTransactionStatus(ctx context.Context, state domain.IdentityState, txID string) error {
	receipt, err := p.transactionService.WaitForTransactionReceipt(ctx, txID)
	if err != nil {
		log.Error(ctx, "error during receipt receiving: ", err)
		return err
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		// wait until transaction will be confirmed if transaction has enough confirmation blocks
		log.Debug(ctx, "Waiting for confirmation", "tx", receipt.TxHash.Hex())
		confirmed, rErr := p.transactionService.WaitForConfirmation(ctx, receipt)
		if rErr != nil {
			return fmt.Errorf("transaction receipt is found, but not confirmed - %s", *state.TxID)
		}
		if !confirmed {
			return fmt.Errorf("transaction receipt is found, but tx is not confirmed yet - %s", *state.TxID)
		}
	} else {
		// do not wait for many confirmations, just save as failed
		log.Info(ctx, "transaction failed", "tx", *state.TxID)
	}

	err = p.updateIdentityStateTxStatus(ctx, &state, receipt)
	if err != nil {
		log.Error(ctx, "error during identity state update: ", err)
		return err
	}

	log.Info(ctx, "transaction status updated", "tx", txID)
	return nil
}

func (p *publisher) updateIdentityStateTxStatus(ctx context.Context, state *domain.IdentityState, receipt *types.Receipt) error {
	header, err := p.transactionService.GetHeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		log.Error(ctx, "couldn't find receipt block: ", err)
		return err
	}

	blockNumber := int(receipt.BlockNumber.Int64())
	state.BlockNumber = &blockNumber

	blockTime := int(header.Time)
	state.BlockTimestamp = &blockTime

	if receipt.Status == types.ReceiptStatusSuccessful {
		state.Status = domain.StatusConfirmed
		err = p.claimService.UpdateClaimsMTPAndState(ctx, state)
	} else {
		state.Status = domain.StatusFailed
		err = p.identityService.UpdateIdentityState(ctx, state)
	}

	if err != nil {
		log.Error(ctx, "state is not updated", err)
		return err
	}

	return nil
}

// CheckTransactionStatus - checks transaction status
func (p *publisher) CheckTransactionStatus(ctx context.Context) {
	jobIDValue, err := uuid.NewUUID()
	if err != nil {
		log.Error(ctx, "error", err)
		return
	}
	ctx = context.WithValue(ctx, jobID, jobIDValue.String())
	log.Info(ctx, "checker status job started", jobID, jobIDValue.String())
	// Get all issuers that have claims not included in any state
	states, err := p.identityService.GetTransactedStates(ctx)
	if err != nil {
		log.Error(ctx, "Error during get transacted states", err)
		return
	}

	// we shouldn't process states which go routines are still in progress

	toCheck := []domain.IdentityState{}
	for i := range states {
		if time.Now().Unix() > states[i].ModifiedAt.Add(p.confirmationTimeout).Unix() {
			toCheck = append(toCheck, states[i])
		}
	}

	// 4. Calculate new states and publish them synchronously
	for i := range toCheck {
		err := p.checkStatus(ctx, &toCheck[i])
		if err != nil {
			log.Error(ctx, "Error during transaction check status", err, "state id", *states[i].State)
			continue
		}
	}

	log.Info(ctx, "checker status job finished", jobID, jobIDValue.String())
}

func (p *publisher) checkStatus(ctx context.Context, state *domain.IdentityState) error {
	// Get receipt and check status
	receipt, err := p.transactionService.GetTransactionReceiptByID(ctx, *state.TxID)
	if err != nil {
		log.Error(ctx, "error during receipt receiving:", err, "state id", *state.TxID)
		return fmt.Errorf("error during receipt receiving::%s - %w", *state.TxID, err)
	}

	// Check if transaction has enough confirmation blocks
	confirmed, err := p.transactionService.CheckConfirmation(ctx, receipt)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("transaction receipt is found, but confirmation is not checked - %s", *state.TxID), err)
		return fmt.Errorf("transaction receipt is found, but confirmation is not checked:%s - %w", *state.TxID, err)
	}

	if !confirmed {
		return fmt.Errorf("transaction receipt is found, but it is not confirmed yet - %s", *state.TxID)
	}

	err = p.updateIdentityStateTxStatus(ctx, state, receipt)
	if err != nil {
		log.Error(ctx, "error during identity state update: ", err)
		return err
	}

	log.Info(ctx, "transaction status updated", "tx", *state.TxID)
	return nil
}
