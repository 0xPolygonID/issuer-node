package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-iden3-crypto/babyjub"
	"github.com/iden3/go-merkletree-sql/v2"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/jackc/pgx/v4"
	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

type identity struct {
	identityRepository      ports.IndentityRepository
	imtRepository           ports.IdentityMerkleTreeRepository
	identityStateRepository ports.IdentityStateRepository
	claimsRepository        ports.ClaimsRepository
	storage                 *db.Storage
	mtservice               ports.MtService
	kms                     kms.KMSType
}

func NewIdentity(kms kms.KMSType, identityRepository ports.IndentityRepository, imtRepository ports.IdentityMerkleTreeRepository, identityStateRepository ports.IdentityStateRepository, mtservice ports.MtService, claimsRepository ports.ClaimsRepository, storage *db.Storage) ports.IndentityService {
	return &identity{
		identityRepository:      identityRepository,
		imtRepository:           imtRepository,
		identityStateRepository: identityStateRepository,
		claimsRepository:        claimsRepository,
		mtservice:               mtservice,
		storage:                 storage,
		kms:                     kms,
	}
}

func (i *identity) Create(ctx context.Context, hostURL string) (*domain.Identity, error) {
	var identifier *core.DID
	var err error
	err = i.storage.Pgx.BeginFunc(ctx,
		func(tx pgx.Tx) error {
			identifier, _, err = i.createIdentity(ctx, tx, hostURL)
			if err != nil {
				return fmt.Errorf("can't create identity: %w", err)
			}

			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("can't create identity: %w", err)
	}

	identityDB, err := i.identityRepository.GetByID(ctx, i.storage.Pgx, identifier)
	if err != nil {
		return nil, fmt.Errorf("can't get identity: %w", err)
	}
	return identityDB, nil
}

func (i *identity) createIdentity(ctx context.Context, tx db.Querier, hostURL string) (*core.DID, *big.Int, error) {
	mts, err := i.mtservice.CreateIdentityMerkleTrees(ctx, tx)
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

	// TODO: add config options for Blockchain and network
	didType, err := core.BuildDIDType(core.DIDMethodPolygonID, core.Polygon, core.Mumbai)
	if err != nil {
		return nil, nil, fmt.Errorf("can't build didtype: %w", err)
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

	credentialType := fmt.Sprintf("%s#%s", schema.Metadata.Uris["jsonLdContext"].(string), domain.AuthBJJCredential)
	claimID, err := uuid.NewUUID()
	if err != nil {
		return nil, nil, fmt.Errorf("can't crate uuid: %w", err)
	}

	cred, err := common.CreateCredential(did, cr, schema)
	if err != nil {
		return nil, nil, fmt.Errorf("can't create credential: %w", err)
	}

	cred.ID = fmt.Sprintf("%s/api/v1/claim/%s", strings.TrimSuffix(hostURL, "/"), claimID)
	cs := &verifiable.CredentialStatus{
		ID: fmt.Sprintf("%s/api/v1/identities/%s/claims/revocation/status/%d",
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
