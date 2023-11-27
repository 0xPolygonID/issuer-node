package services

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/go-iden3-auth/v2/pubsignals"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
	mtproof "github.com/iden3/merkletree-proof"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/config"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/event"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/pkg/credentials/revocation_status"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/circuit/signer"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/suite"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/suite/babyjubjub"
	"github.com/polygonid/sh-id-platform/pkg/primitive"
	"github.com/polygonid/sh-id-platform/pkg/pubsub"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	transitionDelay = time.Minute * 5
	serviceContext  = "https://www.w3.org/ns/did/v1"
	authReason      = "authentication"
)

// ErrWrongDIDMetada - represents an error in the identity metadata
var (
	// ErrWrongDIDMetada - represents an error in the identity metadata
	ErrWrongDIDMetada = errors.New("wrong DID Metadata")
	// ErrAssigningMTPProof - represents an error in the identity metadata
	ErrAssigningMTPProof = errors.New("error assigning the MTP Proof from Auth Claim. If this identity has keyType=ETH you must to publish the state first")
)

type identity struct {
	identityRepository      ports.IndentityRepository
	imtRepository           ports.IdentityMerkleTreeRepository
	identityStateRepository ports.IdentityStateRepository
	claimsRepository        ports.ClaimsRepository
	revocationRepository    ports.RevocationRepository
	connectionsRepository   ports.ConnectionsRepository
	sessionManager          ports.SessionRepository
	storage                 *db.Storage
	mtService               ports.MtService
	qrService               ports.QrStoreService
	kms                     kms.KMSType
	verifier                *auth.Verifier

	ignoreRHSErrors          bool
	pubsub                   pubsub.Publisher
	revocationStatusResolver *revocation_status.RevocationStatusResolver
	credentialStatusSettings config.CredentialStatus
	rhsFactory               reverse_hash.Factory
}

// NewIdentity creates a new identity
// nolint
func NewIdentity(kms kms.KMSType, identityRepository ports.IndentityRepository, imtRepository ports.IdentityMerkleTreeRepository, identityStateRepository ports.IdentityStateRepository, mtservice ports.MtService, qrService ports.QrStoreService, claimsRepository ports.ClaimsRepository, revocationRepository ports.RevocationRepository, connectionsRepository ports.ConnectionsRepository, storage *db.Storage, verifier *auth.Verifier, sessionRepository ports.SessionRepository, ps pubsub.Client, credentialStatusSettings config.CredentialStatus, rhsFactory reverse_hash.Factory, revocationStatusResolver *revocation_status.RevocationStatusResolver) ports.IdentityService {
	return &identity{
		identityRepository:       identityRepository,
		imtRepository:            imtRepository,
		identityStateRepository:  identityStateRepository,
		claimsRepository:         claimsRepository,
		revocationRepository:     revocationRepository,
		connectionsRepository:    connectionsRepository,
		sessionManager:           sessionRepository,
		storage:                  storage,
		mtService:                mtservice,
		qrService:                qrService,
		kms:                      kms,
		ignoreRHSErrors:          false,
		verifier:                 verifier,
		pubsub:                   ps,
		credentialStatusSettings: credentialStatusSettings,
		rhsFactory:               rhsFactory,
		revocationStatusResolver: revocationStatusResolver,
	}
}

func (i *identity) GetByDID(ctx context.Context, identifier w3c.DID) (*domain.Identity, error) {
	return i.identityRepository.GetByID(ctx, i.storage.Pgx, identifier)
}

func (i *identity) Create(ctx context.Context, hostURL string, didOptions *ports.DIDCreationOptions) (*domain.Identity, error) {
	var identifier *w3c.DID
	var err error
	err = i.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			var keyType kms.KeyType
			if didOptions == nil || didOptions.KeyType == "" {
				keyType = kms.KeyTypeBabyJubJub
			} else {
				keyType = didOptions.KeyType
			}

			switch keyType {
			case kms.KeyTypeEthereum:
				identifier, _, err = i.createEthIdentity(ctx, tx, hostURL, didOptions)
			case kms.KeyTypeBabyJubJub:
				identifier, _, err = i.createIdentity(ctx, tx, hostURL, didOptions)
			default:
				return fmt.Errorf("unsupported key type: %s", keyType)
			}
			return err
		})

	if err != nil {
		log.Error(ctx, "creating identity", "err", err, "id", identifier)
		return nil, fmt.Errorf("cannot create identity: %w", err)
	}

	identityDB, err := i.identityRepository.GetByID(ctx, i.storage.Pgx, *identifier)
	if err != nil {
		log.Error(ctx, "loading identity", "err", err, "id", identifier)
		return nil, fmt.Errorf("can't get identity: %w", err)
	}
	return identityDB, nil
}

func (i *identity) SignClaimEntry(ctx context.Context, authClaim *domain.Claim, claimEntry *core.Claim) (*verifiable.BJJSignatureProof2021, error) {
	keyID, err := i.getKeyIDFromAuthClaim(ctx, authClaim)
	if err != nil {
		return nil, err
	}

	bjjSigner, err := primitive.NewBJJSigner(i.kms, keyID)
	if err != nil {
		return nil, err
	}
	bjjVerifier := &primitive.BJJVerifier{}

	bbjSuite := babyjubjub.New(suite.WithSigner(bjjSigner),
		suite.WithVerifier(bjjVerifier))

	circuitSigner := signer.New(bbjSuite)

	var issuerMTP verifiable.Iden3SparseMerkleTreeProof
	err = authClaim.MTPProof.AssignTo(&issuerMTP)
	if err != nil {
		log.Error(ctx, "assigning to issuerMTP", "err", err)
		return nil, ErrAssigningMTPProof
	}

	signtureBytes, err := circuitSigner.Sign(ctx, babyjubjub.SignatureType, claimEntry)
	if err != nil {
		return nil, err
	}

	// followed https://w3c-ccg.github.io/ld-proofs/
	var proof verifiable.BJJSignatureProof2021
	proof.Type = babyjubjub.SignatureType
	proof.Signature = hex.EncodeToString(signtureBytes)
	issuerMTP.IssuerData.AuthCoreClaim, err = authClaim.CoreClaim.Get().Hex()
	if err != nil {
		return nil, err
	}

	proof.IssuerData = issuerMTP.IssuerData
	proof.IssuerData.MTP = issuerMTP.MTP

	proof.CoreClaim, err = claimEntry.Hex()
	if err != nil {
		return nil, err
	}

	return &proof, nil
}

func (i *identity) Exists(ctx context.Context, identifier w3c.DID) (bool, error) {
	identity, err := i.identityRepository.GetByID(ctx, i.storage.Pgx, identifier)
	if err != nil {
		return false, err
	}

	return identity != nil, nil
}

// getKeyIDFromAuthClaim finds BJJ KeyID of auth claim
// in registered key providers
func (i *identity) getKeyIDFromAuthClaim(ctx context.Context, authClaim *domain.Claim) (kms.KeyID, error) {
	var keyID kms.KeyID

	if authClaim.Identifier == nil {
		return keyID, errors.New("identifier is empty in auth claim")
	}

	identity, err := w3c.ParseDID(*authClaim.Identifier)
	if err != nil {
		return keyID, err
	}

	entry := authClaim.CoreClaim.Get()
	bjjClaim := entry.RawSlotsAsInts()

	var publicKey babyjub.PublicKey
	publicKey.X, publicKey.Y = bjjClaim[2], bjjClaim[3]

	compPubKey := publicKey.Compress()

	keyIDs, err := i.kms.KeysByIdentity(ctx, *identity)
	if err != nil {
		return keyID, err
	}

	for _, keyID = range keyIDs {
		if keyID.Type != kms.KeyTypeBabyJubJub {
			continue
		}

		pubKeyBytes, err := i.kms.PublicKey(keyID)
		if err != nil {
			return keyID, err
		}
		if bytes.Equal(pubKeyBytes, compPubKey[:]) {
			return keyID, nil
		}
	}

	return keyID, errors.New("private key not found")
}

// Get - returns all the identities
func (i *identity) Get(ctx context.Context) (identities []string, err error) {
	return i.identityRepository.Get(ctx, i.storage.Pgx)
}

// GetLatestStateByID get latest identity state by identifier
func (i *identity) GetLatestStateByID(ctx context.Context, identifier w3c.DID) (*domain.IdentityState, error) {
	// check that identity exists in the db
	state, err := i.identityStateRepository.GetLatestStateByIdentifier(ctx, i.storage.Pgx, &identifier)
	if err != nil {
		return nil, err
	}
	if state == nil {
		return nil, fmt.Errorf("state is not found for identifier: %s",
			identifier.String())
	}
	return state, nil
}

// GetKeyIDFromAuthClaim finds BJJ KeyID of auth claim
// in registered key providers
func (i *identity) GetKeyIDFromAuthClaim(ctx context.Context, authClaim *domain.Claim) (kms.KeyID, error) {
	var keyID kms.KeyID

	if authClaim.Identifier == nil {
		return keyID, errors.New("identifier is empty in auth claim")
	}

	identity, err := w3c.ParseDID(*authClaim.Identifier)
	if err != nil {
		return keyID, err
	}

	entry := authClaim.CoreClaim.Get()
	bjjClaim := entry.RawSlotsAsInts()

	var publicKey babyjub.PublicKey
	publicKey.X = bjjClaim[2]
	publicKey.Y = bjjClaim[3]

	compPubKey := publicKey.Compress()

	keyIDs, err := i.kms.KeysByIdentity(ctx, *identity)
	if err != nil {
		return keyID, err
	}

	for _, keyID = range keyIDs {
		if keyID.Type != kms.KeyTypeBabyJubJub {
			continue
		}

		pubKeyBytes, err := i.kms.PublicKey(keyID)
		if err != nil {
			return keyID, err
		}
		if bytes.Equal(pubKeyBytes, compPubKey[:]) {
			return keyID, nil
		}
	}

	return keyID, errors.New("private key not found")
}

func (i *identity) UpdateState(ctx context.Context, did w3c.DID) (*domain.IdentityState, error) {
	newState := &domain.IdentityState{
		Identifier: did.String(),
		Status:     domain.StatusCreated,
	}

	err := i.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			iTrees, err := i.mtService.GetIdentityMerkleTrees(ctx, tx, &did)
			if err != nil {
				return err
			}

			previousState, err := i.identityStateRepository.GetLatestStateByIdentifier(ctx, tx, &did)
			if err != nil {
				return fmt.Errorf("error getting the identifier last state: %w", err)
			}

			lc, err := i.claimsRepository.GetAllByState(ctx, tx, &did, nil)
			if err != nil {
				return fmt.Errorf("error getting the states: %w", err)
			}

			for i := range lc {
				err = iTrees.AddClaim(ctx, &lc[i])
				if err != nil {
					return err
				}
			}

			err = populateIdentityState(ctx, iTrees, newState, previousState)
			if err != nil {
				return err
			}

			err = i.update(ctx, tx, &did, *newState)
			if err != nil {
				return err
			}

			updatedRevocations, err := i.revocationRepository.UpdateStatus(ctx, tx, &did)
			if err != nil {
				return err
			}

			err = i.identityStateRepository.Save(ctx, tx, *newState)
			if err != nil {
				return fmt.Errorf("error saving new identity state: %w", err)
			}

			rhsPublishers, err := i.rhsFactory.BuildPublishers(ctx, reverse_hash.RHSMode(i.credentialStatusSettings.RHSMode), &kms.KeyID{
				Type: kms.KeyTypeEthereum,
				ID:   i.credentialStatusSettings.OnchainTreeStore.PublishingKeyPath,
			})
			if err != nil {
				log.Error(ctx, "can't create RHS publisher", "err", err)
				return fmt.Errorf("can't create RHS publisher: %w", err)
			}

			var errIn error
			for _, rhsPublisher := range rhsPublishers {
				errIn = rhsPublisher.PushHashesToRHS(ctx, newState, previousState, updatedRevocations, iTrees)
				if errIn != nil {
					log.Error(ctx, "publishing hashes to RHS", "err", errIn)
					if i.ignoreRHSErrors {
						errIn = nil
					} else {
						return errIn
					}
				}
			}
			err = errIn
			return err
		},
	)
	if err != nil {
		return nil, err
	}

	return newState, err
}

func (i *identity) UpdateIdentityState(ctx context.Context, state *domain.IdentityState) error {
	// save identity to store
	err := i.storage.Pgx.BeginFunc(ctx, func(tx pgx.Tx) error {
		affected, err := i.identityStateRepository.UpdateState(ctx, tx, state)
		if err != nil {
			return fmt.Errorf("can't save identity state; %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("identity state hasn't been updated")
		}

		return nil
	})
	return err
}

func (i *identity) Authenticate(ctx context.Context, message string, sessionID uuid.UUID, serverURL string, issuerDID w3c.DID) (*protocol.AuthorizationResponseMessage, error) {
	authReq, err := i.sessionManager.Get(ctx, sessionID.String())
	if err != nil {
		log.Warn(ctx, "authentication session not found")
		return nil, err
	}

	arm, err := i.verifier.FullVerify(ctx, message, authReq, pubsignals.WithAcceptedStateTransitionDelay(transitionDelay))
	if err != nil {
		log.Error(ctx, "authentication failed", "err", err)
		return nil, err
	}

	issuerDoc := newDIDDocument(serverURL, issuerDID)
	bytesIssuerDoc, err := json.Marshal(issuerDoc)
	if err != nil {
		log.Error(ctx, "failed to marshal issuerDoc", "err", err)
		return nil, err
	}

	bytesIssuerDoc = sanitizeIssuerDoc(bytesIssuerDoc)
	userDID, err := w3c.ParseDID(arm.From)
	if err != nil {
		log.Error(ctx, "failed to parse userDID", "err", err)
		return nil, err
	}

	conn := &domain.Connection{
		ID:         uuid.New(),
		IssuerDID:  issuerDID,
		UserDID:    *userDID,
		IssuerDoc:  bytesIssuerDoc,
		UserDoc:    arm.Body.DIDDoc,
		CreatedAt:  time.Now(),
		ModifiedAt: time.Now(),
	}
	var connID uuid.UUID
	if err := i.storage.Pgx.BeginFunc(ctx, func(tx pgx.Tx) error {
		connID, err = i.connectionsRepository.Save(ctx, i.storage.Pgx, conn)
		if err != nil {
			return err
		}

		return i.connectionsRepository.SaveUserAuthentication(ctx, i.storage.Pgx, connID, sessionID, conn.CreatedAt)
	}); err != nil {
		return nil, err
	}

	if connID == conn.ID { // a connection has been created so previously created credentials have to be sent
		err = i.pubsub.Publish(ctx, event.CreateConnectionEvent, &event.CreateConnection{ConnectionID: connID.String(), IssuerID: issuerDID.String()})
		if err != nil {
			log.Error(ctx, "sending connection notification", "err", err.Error(), "connection", connID)
		}
	}

	return arm, nil
}

func (i *identity) CreateAuthenticationQRCode(ctx context.Context, serverURL string, issuerDID w3c.DID) (string, uuid.UUID, error) {
	sessionID := uuid.New()
	reqID := uuid.New().String()

	qrCode := &protocol.AuthorizationRequestMessage{
		From:     issuerDID.String(),
		ID:       reqID,
		ThreadID: reqID,
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.AuthorizationRequestMessageType,
		Body: protocol.AuthorizationRequestMessageBody{
			CallbackURL: fmt.Sprintf("%s/v1/authentication/callback?sessionID=%s", serverURL, sessionID),
			Reason:      authReason,
		},
	}
	if err := i.sessionManager.Set(ctx, sessionID.String(), *qrCode); err != nil {
		return "", uuid.Nil, err
	}

	raw, err := json.Marshal(qrCode)
	if err != nil {
		return "", uuid.Nil, err
	}
	id, err := i.qrService.Store(ctx, raw, DefaultQRBodyTTL)
	if err != nil {
		return "", uuid.Nil, err
	}
	return i.qrService.ToURL(serverURL, id), nil
}

func (i *identity) update(ctx context.Context, conn db.Querier, id *w3c.DID, currentState domain.IdentityState) error {
	claims, err := i.claimsRepository.GetAllByState(ctx, conn, id, nil)
	if err != nil {
		return err
	}

	// do not have claims to process
	if len(claims) == 0 {
		return nil
	}

	for j := range claims {
		var err error
		claims[j].IdentityState = currentState.State

		affected, err := i.claimsRepository.UpdateState(ctx, i.storage.Pgx, &claims[j])
		if err != nil {
			return fmt.Errorf("can't update claim: %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("claim has not been updated %v", claims[j])
		}
	}

	return nil
}

// populate identity state with data needed to do generate new state.
// Get Data from MT and previous state
func populateIdentityState(ctx context.Context, trees *domain.IdentityMerkleTrees, state, previousState *domain.IdentityState) error {
	claimsTree, err := trees.ClaimsTree()
	if err != nil {
		return err
	}

	revTree, err := trees.RevsTree()
	if err != nil {
		return err
	}

	rootsTree, err := trees.RootsTree()
	if err != nil {
		return err
	}

	_, _, _, err = rootsTree.Get(ctx, claimsTree.Root().BigInt())
	if err == merkletree.ErrKeyNotFound {
		err = rootsTree.Add(ctx, claimsTree.Root().BigInt(), big.NewInt(0))
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}

	// calculate identity state
	currentState, err := merkletree.HashElems(claimsTree.Root().BigInt(), revTree.Root().BigInt(), rootsTree.Root().BigInt())
	if err != nil {
		return err
	}

	hex := currentState.Hex()
	state.State = &hex
	claimTreeRootHex := claimsTree.Root().Hex()
	state.ClaimsTreeRoot = &claimTreeRootHex
	revTreeHex := revTree.Root().Hex()
	state.RevocationTreeRoot = &revTreeHex
	rootOfRootsTreeHex := rootsTree.Root().Hex()
	state.RootOfRoots = &rootOfRootsTreeHex

	state.PreviousState = previousState.State

	return nil
}

// createEthIdentity - creates a new eth identity
func (i *identity) createEthIdentity(ctx context.Context, tx db.Querier, hostURL string, didOptions *ports.DIDCreationOptions) (*w3c.DID, *big.Int, error) {
	mts, err := i.mtService.CreateIdentityMerkleTrees(ctx, tx)
	if err != nil {
		log.Error(ctx, "creating identity markle tree", "err", err)
		return nil, nil, err
	}

	var key kms.KeyID
	key, err = i.kms.CreateKey(kms.KeyTypeEthereum, nil)
	if err != nil {
		return nil, nil, err
	}

	identity, did, err := i.createEthIdentityFromKeyID(ctx, mts, &key, didOptions, tx)
	if err != nil {
		return nil, nil, err
	}

	if err = i.identityRepository.Save(ctx, tx, identity); err != nil {
		log.Error(ctx, "saving identity", "err", err)
		return nil, nil, errors.Join(err, errors.New("can't save identity"))
	}

	identity.State.Status = domain.StatusConfirmed
	err = i.identityStateRepository.Save(ctx, tx, identity.State)
	if err != nil {
		log.Error(ctx, "saving identity state", "err", err)
		return nil, nil, errors.Join(err, errors.New("can't save identity state"))
	}

	// add auth claim
	bjjKey, err := i.kms.CreateKey(kms.KeyTypeBabyJubJub, did)
	if err != nil {
		return nil, nil, err
	}

	bjjPubKey, err := bjjPubKey(i.kms, bjjKey)
	if err != nil {
		return nil, nil, err
	}

	authClaim, err := newAuthClaim(bjjPubKey)
	if err != nil {
		return nil, nil, errors.Join(err, errors.New("can't create auth claim"))
	}
	var revNonce uint64 = 0
	authClaim.SetRevocationNonce(revNonce)

	claimsTree, err := mts.ClaimsTree()
	if err != nil {
		return nil, nil, err
	}

	authClaimModel, err := i.authClaimToModel(ctx, did, identity, authClaim, claimsTree, bjjPubKey, hostURL, didOptions.AuthBJJCredentialStatus, false)
	if err != nil {
		log.Error(ctx, "auth claim to model", "err", err)
		return nil, nil, err
	}

	_, err = i.claimsRepository.Save(ctx, tx, authClaimModel)
	if err != nil {
		return nil, nil, errors.Join(err, errors.New("can't save auth claim"))
	}

	return did, identity.State.TreeState().State.BigInt(), nil
}

// createIdentity - creates a new identity
func (i *identity) createIdentity(ctx context.Context, tx db.Querier, hostURL string, didOptions *ports.DIDCreationOptions) (*w3c.DID, *big.Int, error) {
	if didOptions == nil {
		// nolint : it's a right assignment
		didOptions = &ports.DIDCreationOptions{
			Method:                  core.DIDMethodIden3,
			Blockchain:              core.NoChain,
			Network:                 core.NoNetwork,
			KeyType:                 kms.KeyTypeBabyJubJub,
			AuthBJJCredentialStatus: verifiable.SparseMerkleTreeProof,
		}
	}

	mts, err := i.mtService.CreateIdentityMerkleTrees(ctx, tx)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create identity markle tree: %w", err)
	}

	key, err := i.kms.CreateKey(kms.KeyTypeBabyJubJub, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create babyJubJub key: %w", err)
	}

	pubKey, err := bjjPubKey(i.kms, key)
	if err != nil {
		return nil, nil, fmt.Errorf("can't get babyJubJub public key: %w", err)
	}

	authClaim, err := newAuthClaim(pubKey)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create auth claim: %w", err)
	}

	var revNonce uint64 = 0
	authClaim.SetRevocationNonce(revNonce)

	identity, did, err := i.addGenesisClaimsToTree(ctx, mts, &key, authClaim, didOptions, tx)
	if err != nil {
		log.Error(ctx, "adding genesis claims to tree", "err", err)
		return nil, nil, fmt.Errorf("can't add genesis claims to tree: %w", err)
	}

	claimsTree, err := mts.ClaimsTree()
	if err != nil {
		return nil, nil, err
	}

	authClaimModel, err := i.authClaimToModel(ctx, did, identity, authClaim, claimsTree, pubKey, hostURL, didOptions.AuthBJJCredentialStatus, true)
	if err != nil {
		log.Error(ctx, "auth claim to model", "err", err)
		return nil, nil, err
	}

	_, err = i.claimsRepository.Save(ctx, tx, authClaimModel)
	if err != nil {
		return nil, nil, fmt.Errorf("can't save auth claim: %w", err)
	}

	if err = i.identityRepository.Save(ctx, tx, identity); err != nil {
		return nil, nil, fmt.Errorf("can't save identity: %w", err)
	}

	rhsPublishers, err := i.rhsFactory.BuildPublishers(ctx, reverse_hash.RHSMode(i.credentialStatusSettings.RHSMode), &kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   i.credentialStatusSettings.OnchainTreeStore.PublishingKeyPath,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("can't create RHS publisher: %w", err)
	}

	if len(rhsPublishers) > 0 {
		log.Info(ctx, "publishing state to RHS", "publishers", len(rhsPublishers))
		for _, rhsPublisher := range rhsPublishers {
			err := rhsPublisher.PublishNodesToRHS(ctx, []mtproof.Node{
				{
					Hash: identity.State.TreeState().State,
					Children: []*merkletree.Hash{
						claimsTree.Root(),
						&merkletree.HashZero,
						&merkletree.HashZero,
					},
				},
			})
			if err != nil {
				log.Error(ctx, "publishing state to RHS", "err", err)
				return nil, nil, err
			}
		}
	}

	identity.State.Status = domain.StatusConfirmed
	err = i.identityStateRepository.Save(ctx, tx, identity.State)
	if err != nil {
		log.Error(ctx, "saving identity state", "err", err)
		return nil, nil, fmt.Errorf("can't save identity state: %w", err)
	}

	return did, identity.State.TreeState().State.BigInt(), nil
}

func (i *identity) createEthIdentityFromKeyID(ctx context.Context, mts *domain.IdentityMerkleTrees, key *kms.KeyID, didOptions *ports.DIDCreationOptions, tx db.Querier) (*domain.Identity, *w3c.DID, error) {
	pubKey, err := ethPubKey(i.kms, *key)
	if err != nil {
		return nil, nil, err
	}
	address := crypto.PubkeyToAddress(*pubKey)
	var ethAddr [20]byte
	copy(ethAddr[:], address.Bytes())

	currentState := core.GenesisFromEthAddress(ethAddr)
	if err != nil {
		return nil, nil, err
	}

	didType, err := core.BuildDIDType(didOptions.Method, didOptions.Blockchain, didOptions.Network)
	if err != nil {
		return nil, nil, err
	}

	did, err := core.NewDID(didType, currentState)
	if err != nil {
		return nil, nil, err
	}

	err = mts.BindToIdentifier(tx, did)
	if err != nil {
		return nil, nil, err
	}

	for _, mt := range mts.GetMtModels() {
		err := i.imtRepository.UpdateByID(ctx, tx, mt)
		if err != nil {
			return nil, nil, fmt.Errorf("can't update merkle tree: %w", err)
		}
	}

	//nolint:ineffassign,staticcheck // old key ID is invalid after this operation,
	// override it if somebody would try to use it in the future
	// to prevent possible errors.
	_, err = i.kms.LinkToIdentity(ctx, *key, *did)
	if err != nil {
		return nil, nil, err
	}
	// empty genesis state for eth identity
	identity, err := domain.NewIdentityFromIdentifier(did, merkletree.HashZero.Hex())
	if err != nil {
		return nil, nil, err
	}
	return identity, did, nil
}

func (i *identity) getAuthClaimMtpProof(ctx context.Context, claimsTree *merkletree.MerkleTree, currentState *merkletree.Hash, authClaim *core.Claim, did *w3c.DID) (verifiable.Iden3SparseMerkleTreeProof, error) {
	index, err := authClaim.HIndex()
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	proof, _, err := claimsTree.GenerateProof(ctx, index, nil)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	authClaimHex, err := authClaim.Hex()
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, fmt.Errorf("auth claim core hex error: %w", err)
	}

	stateHex := currentState.Hex()
	cltHex := claimsTree.Root().Hex()
	mtpProof := verifiable.Iden3SparseMerkleTreeProof{
		Type: verifiable.Iden3SparseMerkleTreeProofType,
		IssuerData: verifiable.IssuerData{
			ID:            did.String(),
			AuthCoreClaim: authClaimHex,
			State: verifiable.State{
				ClaimsTreeRoot: &cltHex,
				Value:          &stateHex,
			},
		},
		CoreClaim: authClaimHex,
		MTP:       proof,
	}
	return mtpProof, nil
}

func (i *identity) GetTransactedStates(ctx context.Context) ([]domain.IdentityState, error) {
	var states []domain.IdentityState
	var err error
	err = i.storage.Pgx.BeginFunc(ctx, func(tx pgx.Tx) error {
		states, err = i.identityStateRepository.GetStatesByStatus(ctx, tx, domain.StatusTransacted)

		return err
	})
	if err != nil {
		return nil, err
	}

	return states, nil
}

func (i *identity) GetStates(ctx context.Context, issuerDID w3c.DID) ([]domain.IdentityState, error) {
	return i.identityStateRepository.GetStates(ctx, i.storage.Pgx, issuerDID)
}

func (i *identity) GetUnprocessedIssuersIDs(ctx context.Context) ([]*w3c.DID, error) {
	return i.identityRepository.GetUnprocessedIssuersIDs(ctx, i.storage.Pgx)
}

func (i *identity) HasUnprocessedStatesByID(ctx context.Context, identifier w3c.DID) (bool, error) {
	return i.identityRepository.HasUnprocessedStatesByID(ctx, i.storage.Pgx, &identifier)
}

func (i *identity) HasUnprocessedAndFailedStatesByID(ctx context.Context, identifier w3c.DID) (bool, error) {
	return i.identityRepository.HasUnprocessedAndFailedStatesByID(ctx, i.storage.Pgx, &identifier)
}

func (i *identity) GetNonTransactedStates(ctx context.Context) ([]domain.IdentityState, error) {
	states, err := i.identityStateRepository.GetStatesByStatus(ctx, i.storage.Pgx, domain.StatusCreated)
	if err != nil {
		return nil, fmt.Errorf("error getting non transacted states: %w", err)
	}

	return states, nil
}

func (i *identity) GetFailedState(ctx context.Context, identifier w3c.DID) (*domain.IdentityState, error) {
	states, err := i.identityStateRepository.GetStatesByStatusAndIssuerID(ctx, i.storage.Pgx, domain.StatusFailed, identifier)
	if err != nil {
		return nil, fmt.Errorf("error getting failed state: %w", err)
	}
	if len(states) > 0 {
		return &states[0], nil
	}
	return nil, nil
}

func (i *identity) PublishGenesisStateToRHS(ctx context.Context, did *w3c.DID) error {
	identity, err := i.identityRepository.GetByID(ctx, i.storage.Pgx, *did)
	if err != nil {
		log.Error(ctx, "can't get identity", "err", err)
		return err
	}

	if kms.KeyType(identity.KeyType) == kms.KeyTypeEthereum {
		return errors.New("can't publish genesis state for the identity based on the ethereum key")
	}

	publishers, err := i.rhsFactory.BuildPublishers(ctx, reverse_hash.RHSMode(i.credentialStatusSettings.RHSMode), &kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   i.credentialStatusSettings.OnchainTreeStore.PublishingKeyPath,
	})
	if err != nil {
		log.Error(ctx, "can't create RHS client", "err", err)
		return err
	}

	if len(publishers) == 0 {
		log.Error(ctx, "rhs client is not initialized")
		return errors.New("rhs client is not initialized")
	}

	genesisState, err := i.identityStateRepository.GetGenesisState(ctx, i.storage.Pgx, did.String())
	if err != nil {
		log.Error(ctx, "can't get genesis state", "err", err, "did", did.String())
		return err
	}

	if genesisState == nil {
		log.Error(ctx, "genesis state is not found")
		return errors.New("genesis state is not found")
	}

	for _, publisher := range publishers {
		err = publisher.PublishNodesToRHS(ctx, []mtproof.Node{
			{
				Hash: genesisState.TreeState().State,
				Children: []*merkletree.Hash{
					genesisState.TreeState().ClaimsRoot,
					genesisState.TreeState().RevocationRoot,
					genesisState.TreeState().RootOfRoots,
				},
			},
		})
		if err != nil {
			log.Error(ctx, "publishing genesis state to RHS", "err", err)
			return err
		}
	}

	return nil
}

func (i *identity) addGenesisClaimsToTree(ctx context.Context,
	mts *domain.IdentityMerkleTrees,
	key *kms.KeyID,
	authClaim *core.Claim,
	didOptions *ports.DIDCreationOptions,
	tx db.Querier,
) (*domain.Identity, *w3c.DID, error) {
	entry := common.TreeEntryFromCoreClaim(*authClaim)
	err := mts.AddEntry(ctx, &entry)
	if err != nil {
		return nil, nil, fmt.Errorf("can't add entry to merkle tree: %w", err)
	}

	claimsTree, err := mts.ClaimsTree()
	if err != nil {
		return nil, nil, fmt.Errorf("can't get claims tree: %w", err)
	}

	currentState, err := merkletree.HashElems(claimsTree.Root().BigInt(), merkletree.HashZero.BigInt(), merkletree.HashZero.BigInt())
	if err != nil {
		return nil, nil, fmt.Errorf("can't add get current state from merkle tree: %w", err)
	}

	// TODO: add config options for blockchain and network
	didType, err := core.BuildDIDType(didOptions.Method, didOptions.Blockchain, didOptions.Network)
	if err != nil {
		return nil, nil, ErrWrongDIDMetada
	}

	did, err := core.NewDIDFromIdenState(didType, currentState.BigInt())
	if err != nil {
		return nil, nil, fmt.Errorf("can't genesis from state: %w", err)
	}

	err = mts.BindToIdentifier(tx, did)
	if err != nil {
		return nil, nil, fmt.Errorf("can't bind identifier to merkle tree: %w", err)
	}

	for _, mt := range mts.GetMtModels() {
		err := i.imtRepository.UpdateByID(ctx, tx, mt)
		if err != nil {
			return nil, nil, fmt.Errorf("can't update merkle tree: %w", err)
		}
	}

	_, err = i.kms.LinkToIdentity(ctx, *key, *did)
	if err != nil {
		return nil, nil, fmt.Errorf("can't link to identity: %w", err)
	}

	identity, err := domain.NewIdentityFromIdentifier(did, currentState.Hex())
	if err != nil {
		log.Error(ctx, "can't create identity from identifier", "err", err)
		return nil, nil, err
	}
	claimsTreeHex := claimsTree.Root().Hex()
	identity.State.ClaimsTreeRoot = &claimsTreeHex

	return identity, did, nil
}

func (i *identity) authClaimToModel(ctx context.Context, did *w3c.DID, identity *domain.Identity, authClaim *core.Claim, claimsTree *merkletree.MerkleTree, pubKey *babyjub.PublicKey, hostURL string, status verifiable.CredentialStatusType, isAuthInGenesis bool) (*domain.Claim, error) {
	authClaimData := make(map[string]interface{})
	authClaimData["x"] = pubKey.X.String()
	authClaimData["y"] = pubKey.Y.String()

	authMarshalledClaimData, err := json.Marshal(authClaimData)
	if err != nil {
		return nil, err
	}

	revNonce := authClaim.GetRevocationNonce()

	authCredReq := common.CredentialRequest{
		CredentialSchema:  domain.AuthBJJCredentialJSONSchemaURL,
		Type:              domain.AuthBJJCredential,
		CredentialSubject: authMarshalledClaimData,
		Version:           0,
		RevNonce:          &revNonce,
		LDContext:         domain.AuthBJJCredentialJSONLDContext,
	}
	exp, ok := authClaim.GetExpirationDate()
	if !ok {
		authCredReq.Expiration = 0
	} else {
		authCredReq.Expiration = exp.Unix()
	}

	authClaimID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	authCred, err := common.CreateCredential(did, authCredReq)
	if err != nil {
		return nil, err
	}

	authCred.ID = fmt.Sprintf("%s/api/v1/claim/%s", strings.TrimSuffix(hostURL, "/"), authClaimID)
	cs, err := i.revocationStatusResolver.GetCredentialRevocationStatus(ctx, *did, revNonce, *identity.State.State, status)
	if err != nil {
		log.Error(ctx, "get credential status", "err", err)
		return nil, err
	}

	authCred.CredentialStatus = cs

	authMarshaledCredential, err := json.Marshal(authCred)
	if err != nil {
		log.Error(ctx, "marshal auth credential", "err", err)
		return nil, err
	}

	authClaimModel, err := domain.NewClaimModel(domain.AuthBJJCredentialJSONSchemaURL, domain.AuthBJJCredentialTypeID, *authClaim, nil)
	if err != nil {
		log.Error(ctx, "can't create auth claim model", "err", err)
		return nil, err
	}

	if isAuthInGenesis {
		authMtpProof, err := i.getAuthClaimMtpProof(ctx, claimsTree, identity.State.TreeState().State, authClaim, did)
		if err != nil {
			return nil, err
		}

		authJSONProof, err := json.Marshal(authMtpProof)
		if err != nil {
			return nil, errors.Join(err, errors.New("can't marshal proof"))
		}

		err = authClaimModel.MTPProof.Set(authJSONProof)
		if err != nil {
			return nil, errors.Join(err, errors.New("failed set mtp proof to auth claim"))
		}
	}

	err = authClaimModel.Data.Set(authMarshaledCredential)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed set auth claim data"))
	}

	err = authClaimModel.CredentialStatus.Set(cs)
	if err != nil {
		return nil, errors.Join(err, errors.New("failed set auth revocation status"))
	}

	authClaimModel.Issuer = did.String()
	if isAuthInGenesis {
		authClaimModel.IdentityState = identity.State.State
	}

	authClaimModel.Identifier = &identity.Identifier
	authClaimModel.MtProof = true
	return authClaimModel, nil
}

// newAuthClaim generate BabyJubKeyTypeAuthorizeKSign claimL
func newAuthClaim(key *babyjub.PublicKey) (*core.Claim, error) {
	revNonce, err := common.RandInt64()
	if err != nil {
		return nil, fmt.Errorf("can't create revocation nonce: %w", err)
	}
	return core.NewClaim(core.AuthSchemaHash,
		core.WithIndexDataInts(key.X, key.Y),
		core.WithRevocationNonce(revNonce))
}

func bjjPubKey(keyMS kms.KMSType, keyID kms.KeyID) (*babyjub.PublicKey, error) {
	keyBytes, err := keyMS.PublicKey(keyID)
	if err != nil {
		return nil, fmt.Errorf("can't get bytes from public key: %w", err)
	}

	return kms.DecodeBJJPubKey(keyBytes)
}

func newDIDDocument(serverURL string, issuerDID w3c.DID) verifiable.DIDDocument {
	return verifiable.DIDDocument{
		Context: []string{serviceContext},
		ID:      issuerDID.String(),
		Service: []interface{}{
			verifiable.Service{
				ID:              fmt.Sprintf("%s#%s", issuerDID, verifiable.Iden3CommServiceType),
				Type:            verifiable.Iden3CommServiceType,
				ServiceEndpoint: fmt.Sprintf("%s/v1/agent", serverURL),
			},
		},
	}
}

func sanitizeIssuerDoc(issDoc []byte) []byte {
	str := strings.Replace(string(issDoc), "\\u0000", "", -1)
	return []byte(str)
}

func ethPubKey(keyMS kms.KMSType, keyID kms.KeyID) (*ecdsa.PublicKey, error) {
	keyBytes, err := keyMS.PublicKey(keyID)
	if err != nil {
		return nil, err
	}
	return kms.DecodeETHPubKey(keyBytes)
}
