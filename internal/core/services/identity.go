package services

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth"
	"github.com/iden3/go-iden3-auth/pubsignals"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree-sql/v2"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/session"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/circuit/signer"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/suite"
	"github.com/polygonid/sh-id-platform/pkg/credentials/signature/suite/babyjubjub"
	"github.com/polygonid/sh-id-platform/pkg/primitive"
	"github.com/polygonid/sh-id-platform/pkg/reverse_hash"
)

const (
	transitionDelay = time.Minute * 5
	serviceContext  = "https://www.w3.org/ns/did/v1"
	authReason      = "authentication"
	randSessions    = 1000000
)

// ErrWrongDIDMetada - represents an error in the identity metadata
var (
	ErrWrongDIDMetada = errors.New("wrong DID Metadata")
)

type identity struct {
	identityRepository      ports.IndentityRepository
	imtRepository           ports.IdentityMerkleTreeRepository
	identityStateRepository ports.IdentityStateRepository
	claimsRepository        ports.ClaimsRepository
	revocationRepository    ports.RevocationRepository
	connectionsRepository   ports.ConnectionsRepository
	storage                 *db.Storage
	mtService               ports.MtService
	kms                     kms.KMSType
	verifier                *auth.Verifier
	sessionManager          session.Manager

	ignoreRHSErrors bool
	rhsPublisher    reverse_hash.RhsPublisher
}

// NewIdentity creates a new identity
func NewIdentity(kms kms.KMSType, identityRepository ports.IndentityRepository, imtRepository ports.IdentityMerkleTreeRepository, identityStateRepository ports.IdentityStateRepository, mtservice ports.MtService, claimsRepository ports.ClaimsRepository, revocationRepository ports.RevocationRepository, connectionsRepository ports.ConnectionsRepository, storage *db.Storage, rhsPublisher reverse_hash.RhsPublisher, verifier *auth.Verifier, sessionManager session.Manager) ports.IdentityService {
	return &identity{
		identityRepository:      identityRepository,
		imtRepository:           imtRepository,
		identityStateRepository: identityStateRepository,
		claimsRepository:        claimsRepository,
		revocationRepository:    revocationRepository,
		connectionsRepository:   connectionsRepository,
		storage:                 storage,
		mtService:               mtservice,
		kms:                     kms,
		ignoreRHSErrors:         false,
		rhsPublisher:            rhsPublisher,
		verifier:                verifier,
		sessionManager:          sessionManager,
	}
}

func (i *identity) Create(ctx context.Context, DIDMethod string, blockchain, networkID, hostURL string) (*domain.Identity, error) {
	var identifier *core.DID
	var err error
	err = i.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			identifier, _, err = i.createIdentity(ctx, tx, DIDMethod, blockchain, networkID, hostURL)
			if err != nil {
				if errors.Is(err, ErrWrongDIDMetada) {
					return err
				}
				return fmt.Errorf("can't create identity: %w", err)
			}

			return nil
		})

	if err != nil {
		log.Error(ctx, "creating identity", err, "id", identifier)
		return nil, fmt.Errorf("cannot create identity: %w", err)
	}

	identityDB, err := i.identityRepository.GetByID(ctx, i.storage.Pgx, identifier)
	if err != nil {
		log.Error(ctx, "loading identity", err, "id", identifier)
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

	var issuerMTP verifiable.Iden3SparseMerkleProof
	err = authClaim.MTPProof.AssignTo(&issuerMTP)
	if err != nil {
		return nil, err
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

func (i *identity) Exists(ctx context.Context, identifier *core.DID) (bool, error) {
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

	identity, err := core.ParseDID(*authClaim.Identifier)
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
func (i *identity) GetLatestStateByID(ctx context.Context, identifier *core.DID) (*domain.IdentityState, error) {
	// check that identity exists in the db
	state, err := i.identityStateRepository.GetLatestStateByIdentifier(ctx, i.storage.Pgx, identifier)
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

	identity, err := core.ParseDID(*authClaim.Identifier)
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

func (i *identity) UpdateState(ctx context.Context, did *core.DID) (*domain.IdentityState, error) {
	newState := &domain.IdentityState{
		Identifier: did.String(),
		Status:     domain.StatusCreated,
	}

	err := i.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			iTrees, err := i.mtService.GetIdentityMerkleTrees(ctx, tx, did)
			if err != nil {
				return err
			}

			previousState, err := i.identityStateRepository.GetLatestStateByIdentifier(ctx, tx, did)
			if err != nil {
				return fmt.Errorf("error getting the identifier last state: %w", err)
			}

			lc, err := i.claimsRepository.GetAllByState(ctx, tx, did, nil)
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

			err = i.update(ctx, tx, did, *newState)
			if err != nil {
				return err
			}

			updatedRevocations, err := i.revocationRepository.UpdateStatus(ctx, tx, did)
			if err != nil {
				return err
			}

			err = i.identityStateRepository.Save(ctx, tx, *newState)
			if err != nil {
				return fmt.Errorf("error saving new identity state: %w", err)
			}

			err = i.rhsPublisher.PushHashesToRHS(ctx, newState, previousState, updatedRevocations, iTrees)
			if err != nil {
				log.Error(ctx, "publishing hashes to RHS", err)
				if i.ignoreRHSErrors {
					err = nil
				} else {
					return err
				}
			}

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

func (i *identity) Authenticate(ctx context.Context, message, sessionID, serverURL string, issuerDID core.DID) error {
	authReq, err := i.sessionManager.Get(ctx, sessionID)
	if err != nil {
		log.Warn(ctx, "authentication session not found")
		return err
	}

	arm, err := i.verifier.FullVerify(ctx, message, authReq, pubsignals.WithAcceptedStateTransitionDelay(transitionDelay))
	if err != nil {
		log.Error(ctx, "authentication failed", err)
		return err
	}

	issuerDoc := newDIDDocument(serverURL, issuerDID)
	bytesIssuerDoc, err := json.Marshal(issuerDoc)
	if err != nil {
		log.Error(ctx, "failed to marshal issuerDoc", err)
		return err
	}

	userDID, err := core.ParseDID(arm.From)
	if err != nil {
		log.Error(ctx, "failed to parse userDID", err)
		return err
	}

	conn := &domain.Connection{
		IssuerDID: issuerDID,
		UserDID:   *userDID,
		IssuerDoc: bytesIssuerDoc,
		UserDoc:   arm.Body.DIDDoc,
	}

	return i.connectionsRepository.Save(ctx, i.storage.Pgx, conn)
}

func (i *identity) CreateAuthenticationQRCode(ctx context.Context, serverURL string, issuerDID core.DID) (*protocol.AuthorizationRequestMessage, error) {
	sessionID := rand.Intn(randSessions)

	id, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	qrCode := &protocol.AuthorizationRequestMessage{
		From:     issuerDID.String(),
		ID:       id.String(),
		ThreadID: id.String(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.AuthorizationRequestMessageType,
		Body: protocol.AuthorizationRequestMessageBody{
			CallbackURL: fmt.Sprintf("%s/v1/authentication/callback?sessionID=%d", serverURL, sessionID),
			Reason:      authReason,
		},
	}

	err = i.sessionManager.Set(ctx, strconv.Itoa(sessionID), *qrCode)

	return qrCode, err
}

func (i *identity) update(ctx context.Context, conn db.Querier, id *core.DID, currentState domain.IdentityState) error {
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

func (i *identity) createIdentity(ctx context.Context, tx db.Querier, DIDMethod string, blockchain, networkID, hostURL string) (*core.DID, *big.Int, error) {
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
	entry := common.TreeEntryFromCoreClaim(*authClaim)
	err = mts.AddEntry(ctx, &entry)
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
	// didType, err := core.BuildDIDType(core.DIDMethodPolygonID, core.Polygon, core.Mumbai)
	didType, err := core.BuildDIDType(core.DIDMethod(DIDMethod), core.Blockchain(blockchain), core.NetworkID(networkID))
	if err != nil {
		return nil, nil, ErrWrongDIDMetada
	}

	identifier, err := core.IdGenesisFromIdenState(didType, currentState.BigInt())
	if err != nil {
		return nil, nil, fmt.Errorf("can't genesis from state: %w", err)
	}

	did, err := core.ParseDIDFromID(*identifier)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse did: %w", err)
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

	_, err = i.kms.LinkToIdentity(ctx, key, *did)
	if err != nil {
		return nil, nil, fmt.Errorf("can't link to identity: %w", err)
	}

	identity := domain.NewIdentityFromIdentifier(did, currentState.Hex())
	claimsTreeHex := claimsTree.Root().Hex()
	identity.State.ClaimsTreeRoot = &claimsTreeHex

	claimData := make(map[string]interface{})
	claimData["x"] = pubKey.X.String()
	claimData["y"] = pubKey.Y.String()

	marshalledClaimData, err := json.Marshal(claimData)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal claim data: %w", err)
	}

	cr := common.CredentialRequest{
		CredentialSchema:  domain.AuthBJJCredentialJSONSchemaURL,
		Type:              domain.AuthBJJCredential,
		CredentialSubject: marshalledClaimData,
		Version:           0,
		RevNonce:          &revNonce,
	}

	exp, ok := authClaim.GetExpirationDate()
	if !ok {
		cr.Expiration = 0
	} else {
		cr.Expiration = exp.Unix()
	}

	var schema jsonSuite.Schema
	err = json.Unmarshal([]byte(domain.AuthBJJCredentialSchemaJSON), &schema)
	if err != nil {
		return nil, nil, fmt.Errorf("can't unmarshal the shema: %w", err)
	}

	var jsonLdContext string
	if jsonLdContext, ok = schema.Metadata.Uris["jsonLdContext"].(string); !ok {
		return nil, nil, fmt.Errorf("invalid: %w", err)
	}

	credentialType := fmt.Sprintf("%s#%s", jsonLdContext, domain.AuthBJJCredential)
	claimID, err := uuid.NewUUID()
	if err != nil {
		return nil, nil, fmt.Errorf("can't crate uuid: %w", err)
	}

	cred, err := common.CreateCredential(did, cr, schema)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create credential: %w", err)
	}

	cred.ID = fmt.Sprintf("%s/v1/%s/claims/%s", strings.TrimSuffix(hostURL, "/"), identifier, claimID)
	cs := &verifiable.CredentialStatus{
		ID: fmt.Sprintf("%s/v1/%s/claims/revocation/status/%d",
			hostURL, url.QueryEscape(did.String()), revNonce),
		RevocationNonce: revNonce,
		Type:            verifiable.SparseMerkleTreeProof,
	}

	cred.CredentialStatus = cs

	marshaledCredential, err := json.Marshal(cred)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal credential: %w", err)
	}

	authClaimModel, err := domain.FromClaimer(authClaim, domain.AuthBJJCredentialJSONSchemaURL, credentialType)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create authClaimModel: %w", err)
	}

	mtpProof, err := i.getAuthClaimMtpProof(ctx, claimsTree, currentState, authClaim, did)
	if err != nil {
		return nil, nil, fmt.Errorf("can't add get current state from merkle tree: %w", err)
	}

	err = authClaimModel.Data.Set(marshaledCredential)
	if err != nil {
		return nil, nil, fmt.Errorf("can't set data to auth claim: %w", err)
	}

	err = authClaimModel.CredentialStatus.Set(cs)
	if err != nil {
		return nil, nil, fmt.Errorf("can't set credential status to auth claim: %w", err)
	}

	jsonProof, err := json.Marshal(mtpProof)
	if err != nil {
		return nil, nil, fmt.Errorf("can't marshal proof: %w", err)
	}

	err = authClaimModel.MTPProof.Set(jsonProof)
	if err != nil {
		return nil, nil, fmt.Errorf("can't set mtp proof to auth claim: %w", err)
	}

	authClaimModel.Issuer = did.String()

	if err = i.identityRepository.Save(ctx, tx, identity); err != nil {
		return nil, nil, fmt.Errorf("can't save identity: %w", err)
	}

	// mark genesis state like `confirmed` state.
	identity.State.Status = domain.StatusConfirmed
	err = i.identityStateRepository.Save(ctx, tx, identity.State)
	if err != nil {
		return nil, nil, fmt.Errorf("can't save identity state: %w", err)
	}

	authClaimModel.IdentityState = identity.State.State
	authClaimModel.Identifier = &identity.Identifier
	_, err = i.claimsRepository.Save(ctx, tx, authClaimModel)
	if err != nil {
		return nil, nil, fmt.Errorf("can't save auth claim: %w", err)
	}

	return did, currentState.BigInt(), nil
}

func (i *identity) getAuthClaimMtpProof(ctx context.Context, claimsTree *merkletree.MerkleTree, currentState *merkletree.Hash, authClaim *core.Claim, did *core.DID) (verifiable.Iden3SparseMerkleProof, error) {
	index, err := authClaim.HIndex()
	if err != nil {
		return verifiable.Iden3SparseMerkleProof{}, err
	}

	proof, _, err := claimsTree.GenerateProof(ctx, index, nil)
	if err != nil {
		return verifiable.Iden3SparseMerkleProof{}, err
	}

	authClaimHex, err := authClaim.Hex()
	if err != nil {
		return verifiable.Iden3SparseMerkleProof{}, fmt.Errorf("auth claim core hex error: %w", err)
	}

	stateHex := currentState.Hex()
	cltHex := claimsTree.Root().Hex()
	mtpProof := verifiable.Iden3SparseMerkleProof{
		Type: verifiable.Iden3SparseMerkleProofType,
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

func (i *identity) GetUnprocessedIssuersIDs(ctx context.Context) ([]*core.DID, error) {
	return i.identityRepository.GetUnprocessedIssuersIDs(ctx, i.storage.Pgx)
}

func (i *identity) HasUnprocessedStatesByID(ctx context.Context, identifier *core.DID) (bool, error) {
	return i.identityRepository.HasUnprocessedStatesByID(ctx, i.storage.Pgx, identifier)
}

func (i *identity) GetNonTransactedStates(ctx context.Context) ([]domain.IdentityState, error) {
	states, err := i.identityStateRepository.GetStatesByStatus(ctx, i.storage.Pgx, domain.StatusCreated)
	if err != nil {
		return nil, fmt.Errorf("error getting non transacted states: %w", err)
	}

	return states, nil
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

func newDIDDocument(serverURL string, issuerDID core.DID) verifiable.DIDDocument {
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
