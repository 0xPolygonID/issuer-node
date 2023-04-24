package gateways

import (
	"context"
	"errors"
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
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/sync_ttl_map"
)

type jobIDType string

var (
	// ErrNoStatesToProcess No states to process
	ErrNoStatesToProcess = errors.New("no states to process or previous state transaction failed")
	// ErrStateIsBeingProcessed State is being processed
	ErrStateIsBeingProcessed = errors.New("the state is being processed")
	// ErrNoFailedStatesToProcess - No fialed states to process
	ErrNoFailedStatesToProcess = errors.New("no failed states to process")
)

const (
	jobID              jobIDType = "job-id"
	ttl                          = 60 * time.Second
	transactionCleanup           = 3600 * time.Second
)

// PublisherGateway - Define the interface for publishers.
type PublisherGateway interface {
	PublishState(ctx context.Context, identifier *core.DID, latestState *merkletree.Hash, newState *merkletree.Hash, isOldStateGenesis bool, proof *domain.ZKProof) (*string, error)
}

type publisher struct {
	storage               *db.Storage
	identityService       ports.IdentityService
	claimService          ports.ClaimsService
	mtService             ports.MtService
	kms                   kms.KMSType
	transactionService    ports.TransactionService
	confirmationTimeout   time.Duration
	zkService             ports.ZKGenerator
	publisherGateway      PublisherGateway
	pendingTransactions   *sync_ttl_map.TTLMap
	notificationPublisher pubsub.Publisher
}

// NewPublisher - Constructor
func NewPublisher(storage *db.Storage, identityService ports.IdentityService, claimService ports.ClaimsService, mtService ports.MtService, kms kms.KMSType, transactionService ports.TransactionService, zkService ports.ZKGenerator, publisherGateway PublisherGateway, confirmationTimeout time.Duration, notificationPublisher pubsub.Publisher) *publisher {
	pendingTransactions := sync_ttl_map.New(ttl)
	pendingTransactions.CleaningBackground(transactionCleanup)

	return &publisher{
		identityService:       identityService,
		claimService:          claimService,
		storage:               storage,
		mtService:             mtService,
		kms:                   kms,
		transactionService:    transactionService,
		zkService:             zkService,
		publisherGateway:      publisherGateway,
		confirmationTimeout:   confirmationTimeout,
		pendingTransactions:   pendingTransactions,
		notificationPublisher: notificationPublisher,
	}
}

func (p *publisher) PublishState(ctx context.Context, identifier *core.DID) (*domain.PublishedState, error) {
	idStr := identifier.String()
	processingEntity := p.pendingTransactions.Load(idStr)
	if processingEntity != nil {
		return nil, ErrStateIsBeingProcessed
	}

	p.pendingTransactions.Store(idStr, true)
	newState, err := p.publishState(ctx, identifier)
	if err != nil {
		p.pendingTransactions.Delete(idStr)
	}

	return newState, err
}

func (p *publisher) RetryPublishState(ctx context.Context, identifier *core.DID) (*domain.PublishedState, error) {
	idStr := identifier.String()
	processingEntity := p.pendingTransactions.Load(idStr)
	if processingEntity != nil {
		return nil, ErrStateIsBeingProcessed
	}

	p.pendingTransactions.Store(idStr, true)
	newState, err := p.retrypublishFailedState(ctx, identifier)
	if err != nil {
		p.pendingTransactions.Delete(idStr)
	}

	return newState, err
}

func (p *publisher) publishState(ctx context.Context, identifier *core.DID) (*domain.PublishedState, error) {
	exists, err := p.identityService.HasUnprocessedStatesByID(ctx, *identifier)
	if err != nil {
		log.Error(ctx, "error fetching unprocessed issuers did", "err", err)
		return nil, err
	}

	if !exists {
		log.Info(ctx, "no unprocessed states for the given issuer id")
		return nil, ErrNoStatesToProcess
	}

	// 4. Calculate new states and publish them synchronously
	updatedState, err := p.identityService.UpdateState(ctx, *identifier)
	if err != nil {
		log.Error(ctx, "Error during processing claims", "err", err, "did", identifier.String())
		return nil, err
	}

	txID, err := p.publishProof(ctx, identifier, *updatedState)
	if err != nil {
		log.Error(ctx, "Error during publishing proof:", "err", err, "did", identifier.String())
		updatedState.Status = domain.StatusFailed
		errUpdating := p.identityService.UpdateIdentityState(ctx, updatedState)
		if errUpdating != nil {
			log.Error(ctx, "Error saving the state as failed:", "err", err, "did", identifier.String())
			return nil, errUpdating
		}
		return nil, err
	}

	return &domain.PublishedState{
		TxID:               txID,
		ClaimsTreeRoot:     updatedState.ClaimsTreeRoot,
		State:              updatedState.State,
		RevocationTreeRoot: updatedState.RevocationTreeRoot,
		RootOfRoots:        updatedState.RootOfRoots,
	}, nil
}

func (p *publisher) retrypublishFailedState(ctx context.Context, identifier *core.DID) (*domain.PublishedState, error) {
	failedState, err := p.identityService.GetFailedState(ctx, *identifier)
	if err != nil {
		log.Error(ctx, "error fetching failed state", "err", err)
		return nil, err
	}

	if failedState == nil {
		log.Info(ctx, "no failed state for the given issuer id")
		return nil, ErrNoFailedStatesToProcess
	}

	txID, err := p.publishProof(ctx, identifier, *failedState)
	if err != nil {
		log.Error(ctx, "Error during publishing proof:", "err", err, "did", identifier.String())
		return nil, err
	}

	return &domain.PublishedState{
		TxID:               txID,
		ClaimsTreeRoot:     failedState.ClaimsTreeRoot,
		State:              failedState.State,
		RevocationTreeRoot: failedState.RevocationTreeRoot,
		RootOfRoots:        failedState.RootOfRoots,
	}, nil
}

// PublishProof publishes new proof using the latest state
func (p *publisher) publishProof(ctx context.Context, identifier *core.DID, newState domain.IdentityState) (*string, error) {
	did, err := core.ParseDID(newState.Identifier)
	if err != nil {
		return nil, err
	}

	// 1. Get latest transacted state
	latestState, err := p.identityService.GetLatestStateByID(ctx, *did)
	if err != nil {
		return nil, err
	}

	latestStateHash, err := merkletree.NewHashFromHex(*latestState.State)
	if err != nil {
		return nil, err
	}

	// TODO: core.IdenState should be calculated before state stored to db
	newStateHash, err := merkletree.NewHashFromHex(*newState.State)
	if err != nil {
		return nil, err
	}

	newTreeState, err := newState.ToTreeState()
	if err != nil {
		return nil, err
	}

	authClaim, err := p.claimService.GetAuthClaimForPublishing(ctx, did, *newState.State)
	if err != nil {
		return nil, err
	}

	claimKeyID, err := p.identityService.GetKeyIDFromAuthClaim(ctx, authClaim)
	if err != nil {
		return nil, err
	}

	oldTreeState, err := latestState.ToTreeState()
	if err != nil {
		return nil, err
	}

	circuitAuthClaim, circuitAuthClaimNewStateIncProof, err := p.fillAuthClaimData(ctx, did, authClaim, newState)
	if err != nil {
		return nil, err
	}

	hashOldAndNewStates, err := poseidon.Hash([]*big.Int{oldTreeState.State.BigInt(), newStateHash.BigInt()})
	if err != nil {
		return nil, err
	}

	sigDigest := kms.BJJDigest(hashOldAndNewStates)
	sigBytes, err := p.kms.Sign(ctx, claimKeyID, sigDigest)
	if err != nil {
		return nil, err
	}

	signature, err := kms.DecodeBJJSignature(sigBytes)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	// TODO: Integrate when it's finished
	fullProof, err := p.zkService.Generate(ctx, jsonInputs, string(circuits.StateTransitionCircuitID))
	if err != nil {
		return nil, err
	}

	// 7. Publish state and receive txID

	txID, err := p.publisherGateway.PublishState(ctx, did, latestStateHash, newStateHash, isLatestStateGenesis, fullProof.Proof)
	if err != nil {
		return nil, err
	}

	log.Info(ctx, "Success!", "TxID", txID)

	// 8. Update state with txID value (block values are still default because tx is not confirmed)

	newState.Status = domain.StatusTransacted
	newState.TxID = txID

	err = p.identityService.UpdateIdentityState(ctx, &newState)
	if err != nil {
		return nil, err
	}

	// add go routine that will listen for transaction status update

	go func(ctx context.Context) {
		if err := p.updateTransactionStatus(ctx, newState, *txID); err != nil {
			log.Error(ctx, "cannot update transaction status", "err", err)
		}
		p.pendingTransactions.Delete(identifier.String())
	}(ctx)

	return txID, nil
}

func (p *publisher) fillAuthClaimData(ctx context.Context, identifier *core.DID, authClaim *domain.Claim, newState domain.IdentityState) (
	authClaimData *circuits.ClaimWithMTPProof, authClaimNewStateIncProof *merkletree.Proof, err error,
) {
	err = p.storage.Pgx.BeginFunc(
		ctx, func(tx pgx.Tx) error {
			var errIn error

			var idState *domain.IdentityState
			idState, errIn = p.identityService.GetLatestStateByID(ctx, *identifier)
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
		log.Error(ctx, "error during receipt receiving: ", "err", err, "txID", txID)
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
		log.Error(ctx, "updating identity state", "err", err, "txID", txID)
		return err
	}

	log.Info(ctx, "transaction status updated", "tx", txID)
	return nil
}

func (p *publisher) updateIdentityStateTxStatus(ctx context.Context, state *domain.IdentityState, receipt *types.Receipt) error {
	header, err := p.transactionService.GetHeaderByNumber(ctx, receipt.BlockNumber)
	if err != nil {
		log.Error(ctx, "couldn't find receipt block: ", "err", err, "block", receipt.BlockNumber)
		return err
	}

	blockNumber := int(receipt.BlockNumber.Int64())
	state.BlockNumber = &blockNumber

	blockTime := int(header.Time)
	state.BlockTimestamp = &blockTime

	if receipt.Status == types.ReceiptStatusSuccessful {
		state.Status = domain.StatusConfirmed
		err = p.claimService.UpdateClaimsMTPAndState(ctx, state)
		did, err := core.ParseDID(state.Identifier)
		if err != nil {
			log.Error(ctx, "error getting did from state: ", "err", err, "state", state.StateID)
			return err
		}
		claimsToNotify, err := p.claimService.GetByStateIDWithMTPProof(ctx, did, *state.State)
		if err != nil {
			log.Error(ctx, "couldn't fetch the credentials to send notifications: ", "err", err, "state", state.StateID)
			return err
		}
		log.Info(ctx, "sending notifications:", "numberOfClaims", len(claimsToNotify))

		grupedCredentials := groupByUserId(claimsToNotify)
		for _, claims := range grupedCredentials {
			err = p.notificationPublisher.Publish(ctx, event.CreateCredentialEvent, &event.CreateCredential{CredentialIDs: claims, IssuerID: state.Identifier})
			if err != nil {
				log.Error(ctx, "publish EventCreateCredential", "err", err.Error(), "credential", claims)
				continue
			}
		}
	} else {
		state.Status = domain.StatusFailed
		err = p.identityService.UpdateIdentityState(ctx, state)
	}

	if err != nil {
		log.Error(ctx, "state is not updated", "err", err)
		return err
	}

	return nil
}

// groupByUserId - groups claims by user id
func groupByUserId(claims []*domain.Claim) map[string][]string {
	grouped := make(map[string][]string)
	for _, c := range claims {
		grouped[c.OtherIdentifier] = append(grouped[c.OtherIdentifier], c.ID.String())
	}
	return grouped
}

// CheckTransactionStatus - checks transaction status
func (p *publisher) CheckTransactionStatus(ctx context.Context) {
	jobIDValue, err := uuid.NewUUID()
	if err != nil {
		log.Error(ctx, "Check transaction status", "err", err)
		return
	}
	ctx = context.WithValue(ctx, jobID, jobIDValue.String())
	log.Info(ctx, "checker status job started", "job-id", jobIDValue.String())
	// Get all issuers that have claims not included in any state
	states, err := p.identityService.GetTransactedStates(ctx)
	if err != nil {
		log.Error(ctx, "Error during get transacted states", "err", err)
		return
	}
	// we shouldn't process states which go routines are still in progress

	var toCheck []domain.IdentityState
	for i, state := range states {
		log.Debug(ctx, "examining state", "id", state.StateID, "identifier", state.Identifier, "prev", state.PreviousState, "created_at", state.CreatedAt, "updated_at", state.ModifiedAt)
		if time.Now().Unix() > states[i].ModifiedAt.Add(p.confirmationTimeout).Unix() {
			toCheck = append(toCheck, states[i])
			log.Debug(ctx, "considering state", "id", state.StateID, "identifier", state.Identifier, "prev", state.PreviousState, "created_at", state.CreatedAt, "updated_at", state.ModifiedAt)
		}
	}

	// 4. Calculate new states and publish them synchronously
	for i := range toCheck {
		err := p.checkStatus(ctx, &toCheck[i])
		if err != nil {
			log.Error(ctx, "transaction check status", "err", err, "state id", *states[i].State)
			continue
		}
	}

	log.Info(ctx, "checker status job finished", "job-id", jobIDValue.String())
}

func (p *publisher) checkStatus(ctx context.Context, state *domain.IdentityState) error {
	// Get receipt and check status
	receipt, err := p.transactionService.GetTransactionReceiptByID(ctx, *state.TxID)
	if err != nil {
		log.Error(ctx, "error during receipt receiving:", "err", err, "state-id", *state.TxID)
		return fmt.Errorf("error during receipt receiving::%s: %w", *state.TxID, err)
	}

	// Check if transaction has enough confirmation blocks
	confirmed, err := p.transactionService.CheckConfirmation(ctx, receipt)
	if err != nil {
		log.Error(ctx, fmt.Sprintf("transaction receipt is found, but confirmation is not checked - %s", *state.TxID), "err", err)
		return fmt.Errorf("transaction receipt is found, but confirmation is not checked:%s - %w", *state.TxID, err)
	}

	if !confirmed {
		log.Debug(ctx, "transaction receipt is found, but it is not confirmed yet", "TxID", *state.TxID)
		return ErrStateIsBeingProcessed
	}

	err = p.updateIdentityStateTxStatus(ctx, state, receipt)
	if err != nil {
		log.Error(ctx, "error during identity state update: ", "err", err)
		return err
	}

	log.Info(ctx, "transaction status updated", "tx", *state.TxID)
	return nil
}
