package services

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/utils"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/schema"
	"github.com/polygonid/sh-id-platform/pkg/rand"
)

var (
	ErrClaimNotFound = errors.New("claim not found")                // ErrClaimNotFound Cannot retrieve the given claim 	// ErrProcessSchema Cannot process schema
	ErrJSONLdContext = errors.New("jsonLdContext must be a string") // ErrJSONLdContext Field jsonLdContext must be a string
	ErrLoadingSchema = errors.New("cannot load schema")             // ErrLoadingSchema means the system cannot load the schema file
	ErrMalformedURL  = errors.New("malformed url")                  // ErrMalformedURL The schema url is wrong
	ErrProcessSchema = errors.New("cannot process schema")          // ErrProcessSchema Cannot process schema
)

// ClaimCfg claim service configuration
type ClaimCfg struct {
	RHSEnabled bool // ReverseHash Enabled
	RHSUrl     string
	Host       string
}

type claim struct {
	cfg                     ClaimCfg
	icRepo                  ports.ClaimsRepository
	identitySrv             ports.IdentityService
	mtService               ports.MtService
	identityStateRepository ports.IdentityStateRepository
	storage                 *db.Storage
}

// NewClaim creates a new claim service
func NewClaim(repo ports.ClaimsRepository, idenSrv ports.IdentityService, mtService ports.MtService, identityStateRepository ports.IdentityStateRepository, storage *db.Storage, cfg ClaimCfg) ports.ClaimsService {
	s := &claim{
		cfg: ClaimCfg{
			RHSEnabled: cfg.RHSEnabled,
			RHSUrl:     cfg.RHSUrl,
			Host:       cfg.Host,
		},
		icRepo:                  repo,
		identitySrv:             idenSrv,
		mtService:               mtService,
		identityStateRepository: identityStateRepository,
		storage:                 storage,
	}
	return s
}

// CreateClaim creates a new claim
// 1.- Creates document
// 2.- Signature proof
// 3.- MerkelTree proof
func (c *claim) CreateClaim(ctx context.Context, req *ports.CreateClaimRequest) (*domain.Claim, error) {
	if err := c.guardCreateClaimRequest(req); err != nil {
		log.Warn(ctx, "validating create claim request", "req", req)
		return nil, err
	}

	nonce, err := rand.Int64()
	if err != nil {
		log.Error(ctx, "create a nonce", err)
		return nil, err
	}

	loader := schema.FactoryLoader(req.SchemaURL)
	sch, err := schema.LoadSchema(ctx, loader)
	if err != nil {
		log.Error(ctx, "loading schemaSrv", err, "schemaSrv", req.SchemaURL)
		return nil, ErrLoadingSchema
	}

	jsonLdContext, ok := sch.Metadata.Uris["jsonLdContext"].(string)
	if !ok {
		log.Error(ctx, "invalid jsonLdContext", ErrJSONLdContext)
		return nil, ErrJSONLdContext
	}

	vc, err := c.createVC(req, jsonLdContext, domain.RevNonceUint64(nonce))
	if err != nil {
		log.Error(ctx, "creating verifiable credential", err)
		return nil, err
	}

	credentialType := fmt.Sprintf("%s#%s", jsonLdContext, req.Type)
	coreClaim, err := schema.Process(ctx, loader, credentialType, vc, &processor.CoreClaimOptions{
		RevNonce:              nonce,
		MerklizedRootPosition: defineMerklizedRootPosition(sch.Metadata, req.MerklizedRootPosition),
		Version:               req.Version,
		SubjectPosition:       req.SubjectPos,
		Updatable:             false,
	})
	if err != nil {
		log.Error(ctx, "cannot process the schemaSrv", err)
		return nil, ErrProcessSchema
	}

	authClaim, err := c.getAuthClaim(ctx, req.DID)
	if err != nil {
		log.Error(ctx, "cannot retrieve the auth claim", err)
		return nil, err
	}

	proof, err := c.identitySrv.SignClaimEntry(ctx, authClaim, coreClaim)
	if err != nil {
		log.Error(ctx, "cannot sign claim entry", err)
		return nil, err
	}
	proof.IssuerData.CredentialStatus = c.getRevocationSource(req.DID, authClaim.RevNonce)

	claim, err := domain.FromClaimer(coreClaim, req.SchemaURL, credentialType)
	if err != nil {
		log.Error(ctx, "cannot obtain the claim from claimer", err)
		return nil, err
	}
	issuerDIDString := req.DID.String()
	claim.Identifier = &issuerDIDString
	claim.Issuer = issuerDIDString
	err = claim.SignatureProof.Set(proof)
	if err != nil {
		log.Error(ctx, "cannot set the json signature proof", err)
		return nil, err
	}

	err = claim.Data.Set(vc)
	if err != nil {
		log.Error(ctx, "cannot set the credential", err)
		return nil, err
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	if err != nil {
		log.Error(ctx, "cannot set the credential status", err)
		return nil, err
	}

	claimResp, err := c.save(ctx, claim)
	if err != nil {
		log.Error(ctx, "cannot save the claim", err)
		return nil, err
	}
	return claimResp, err
}

func (c *claim) Revoke(ctx context.Context, id string, nonce uint64, description string) error {
	did, err := core.ParseDID(id)
	if err != nil {
		return fmt.Errorf("error parsing did: %w", err)
	}

	rID := new(big.Int).SetUint64(nonce)
	revocation := domain.Revocation{
		Identifier:  id,
		Nonce:       domain.RevNonceUint64(nonce),
		Version:     0,
		Status:      0,
		Description: description,
	}

	identityTrees, err := c.mtService.GetIdentityMerkleTrees(ctx, c.storage.Pgx, did)
	if err != nil {
		return fmt.Errorf("error getting merkle trees: %w", err)
	}

	err = identityTrees.RevokeClaim(ctx, rID)
	if err != nil {
		return fmt.Errorf("error revoking the claim: %w", err)
	}

	var claim *domain.Claim
	claim, err = c.icRepo.GetByRevocationNonce(ctx, c.storage.Pgx, did, domain.RevNonceUint64(nonce))

	if err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return err
		}
		return fmt.Errorf("error getting the claim by revocation nonce: %w", err)
	}

	claim.Revoked = true
	_, err = c.icRepo.Save(ctx, c.storage.Pgx, claim)
	if err != nil {
		return fmt.Errorf("error saving the claim: %w", err)
	}

	return c.icRepo.RevokeNonce(ctx, c.storage.Pgx, &revocation)
}

func (c *claim) GetByID(ctx context.Context, issID *core.DID, id uuid.UUID) (*domain.Claim, error) {
	claim, err := c.icRepo.GetByIdAndIssuer(ctx, c.storage.Pgx, issID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return nil, ErrClaimNotFound
		}
		return nil, err
	}

	return claim, nil
}

func (c *claim) Agent(ctx context.Context, req *ports.AgentRequest) (interface{}, error) {
	exists, err := c.identitySrv.Exists(ctx, req.IssuerDID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, fmt.Errorf("can not proceed with this identity, not found")
	}
	if req.Type == protocol.RevocationStatusRequestMessageType {
		return c.getAgentRevocationStatus(ctx, req)
	}

	return c.getAgentCredential(ctx, req) // at this point the type is already validated
}

func (c *claim) GetAuthClaim(ctx context.Context, did *core.DID) (*domain.Claim, error) {
	authHash, err := core.AuthSchemaHash.MarshalText()
	if err != nil {
		return nil, err
	}
	return c.icRepo.FindOneClaimBySchemaHash(ctx, c.storage.Pgx, did, string(authHash))
}

func (c *claim) getAgentRevocationStatus(ctx context.Context, basicMessage *ports.AgentRequest) (*protocol.RevocationStatusResponseMessage, error) {
	revData := &protocol.RevocationStatusRequestMessageBody{}
	err := json.Unmarshal(basicMessage.Body, revData)
	if err != nil {
		return nil, fmt.Errorf("invalid revocation request body: %w", err)
	}

	revStatus, err := c.getRevocationNonceMTP(ctx, basicMessage.IssuerDID, revData.RevocationNonce)
	if err != nil {
		return nil, fmt.Errorf("failed get revocation nonce: %w", err)
	}

	return &protocol.RevocationStatusResponseMessage{
		ID:       uuid.NewString(),
		Type:     protocol.RevocationStatusResponseMessageType,
		ThreadID: basicMessage.ThreadID,
		Body:     protocol.RevocationStatusResponseMessageBody{RevocationStatus: *revStatus},
		From:     basicMessage.UserDID.String(),
		To:       basicMessage.IssuerDID.String(),
	}, nil
}

func (c *claim) getAgentCredential(ctx context.Context, basicMessage *ports.AgentRequest) (*protocol.CredentialIssuanceMessage, error) {
	fetchRequestBody := &protocol.CredentialFetchRequestMessageBody{}
	err := json.Unmarshal(basicMessage.Body, fetchRequestBody)
	if err != nil {
		return nil, fmt.Errorf("invalid credential fetch request body: %w", err)
	}

	claim, err := c.icRepo.GetByIdAndIssuer(ctx, c.storage.Pgx, basicMessage.IssuerDID, basicMessage.ClaimID)
	if err != nil {
		return nil, fmt.Errorf("failed get claim by claimID: %w", err)
	}

	if claim.OtherIdentifier != basicMessage.UserDID.String() {
		return nil, fmt.Errorf("claim doesn't relate to sender")
	}

	vc, err := c.schemaSrv.FromClaimModelToW3CCredential(*claim)
	if err != nil {
		return nil, fmt.Errorf("failed to convert claim to  w3cCredential: %w", err)
	}

	return &protocol.CredentialIssuanceMessage{
		ID:       uuid.NewString(),
		Type:     protocol.CredentialIssuanceResponseMessageType,
		ThreadID: basicMessage.ThreadID,
		Body:     protocol.IssuanceMessageBody{Credential: *vc},
		From:     basicMessage.UserDID.String(),
		To:       basicMessage.IssuerDID.String(),
	}, err
}

// getRevocationNonceMTP generates MTP proof for given nonce
func (c *claim) getRevocationNonceMTP(ctx context.Context, did *core.DID, nonce uint64) (*verifiable.RevocationStatus, error) {
	rID := new(big.Int).SetUint64(nonce)
	revocationStatus := &verifiable.RevocationStatus{}

	// current state of identity / the latest published on chain
	state, err := c.identityStateRepository.GetLatestStateByIdentifier(ctx, c.storage.Pgx, did)
	if err != nil {
		return nil, err
	}

	revocationStatus.Issuer.State = state.State
	revocationStatus.Issuer.ClaimsTreeRoot = state.ClaimsTreeRoot
	revocationStatus.Issuer.RevocationTreeRoot = state.RevocationTreeRoot
	revocationStatus.Issuer.RootOfRoots = state.RootOfRoots

	if state.RevocationTreeRoot == nil {

		var mtp *merkletree.Proof
		mtp, err = merkletree.NewProofFromData(false, nil, nil)
		if err != nil {
			return nil, err
		}
		revocationStatus.MTP = *mtp
		return revocationStatus, nil
	}

	revocationTreeHash, err := merkletree.NewHashFromHex(*state.RevocationTreeRoot)
	if err != nil {
		return nil, err
	}
	identityTrees, err := c.mtService.GetIdentityMerkleTrees(ctx, c.storage.Pgx, did)
	if err != nil {
		return nil, err
	}

	// revocation / non revocation MTP for the latest identity state
	proof, err := identityTrees.GenerateRevocationProof(ctx, rID, revocationTreeHash)
	if err != nil {
		return nil, err
	}

	revocationStatus.MTP = *proof

	return revocationStatus, nil
}

func (c *claim) GetAll(ctx context.Context, did *core.DID) ([]*verifiable.W3CCredential, error) {
	claims, err := c.icRepo.GetAllByIssuerID(ctx, c.storage.Pgx, did)
	if err != nil {
		return nil, err
	}

	w3Credentials := make([]*verifiable.W3CCredential, 0)
	for _, cred := range claims {
		w3Cred, err := schema.FromClaimModelToW3CCredential(*cred)
		if err != nil {
			log.Warn(ctx, "could not convert claim model to W3CCredential", err)
			continue
		}

		w3Credentials = append(w3Credentials, w3Cred)
	}

	return w3Credentials, nil
}

func (c *claim) GetRevocationStatus(ctx context.Context, id string, nonce uint64) (*verifiable.RevocationStatus, error) {
	did, err := core.ParseDID(id)
	if err != nil {
		return nil, fmt.Errorf("error parsing did: %w", err)
	}

	rID := new(big.Int).SetUint64(nonce)
	revocationStatus := &verifiable.RevocationStatus{}

	state, err := c.identityStateRepository.GetLatestStateByIdentifier(ctx, c.storage.Pgx, did)
	if err != nil {
		return nil, err
	}

	revocationStatus.Issuer.State = state.State
	revocationStatus.Issuer.ClaimsTreeRoot = state.ClaimsTreeRoot
	revocationStatus.Issuer.RevocationTreeRoot = state.RevocationTreeRoot
	revocationStatus.Issuer.RootOfRoots = state.RootOfRoots

	if state.RevocationTreeRoot == nil {
		var mtp *merkletree.Proof
		mtp, err = merkletree.NewProofFromData(false, nil, nil)
		if err != nil {
			return nil, err
		}
		revocationStatus.MTP = *mtp
		return revocationStatus, nil
	}

	revocationTreeHash, err := merkletree.NewHashFromHex(*state.RevocationTreeRoot)
	if err != nil {
		return nil, err
	}
	identityTrees, err := c.mtService.GetIdentityMerkleTrees(ctx, c.storage.Pgx, did)
	if err != nil {
		return nil, err
	}

	// revocation / non revocation MTP for the latest identity state
	proof, err := identityTrees.GenerateRevocationProof(ctx, rID, revocationTreeHash)
	if err != nil {
		return nil, err
	}

	revocationStatus.MTP = *proof

	return revocationStatus, nil
}

func (c *claim) createVC(claimReq *ports.CreateClaimRequest, jsonLdContext string, nonce domain.RevNonceUint64) (verifiable.W3CCredential, error) {
	vCredential, err := c.newVerifiableCredential(claimReq, jsonLdContext, nonce) // create vc credential
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	return vCredential, nil
}

func (c *claim) save(ctx context.Context, claim *domain.Claim) (*domain.Claim, error) {
	id, err := c.icRepo.Save(ctx, c.storage.Pgx, claim)
	if err != nil {
		return nil, err
	}
	claim.ID = id
	return claim, nil
}

func (c *claim) getAuthClaim(ctx context.Context, did core.DID) (*domain.Claim, error) {
	authHash, err := core.AuthSchemaHash.MarshalText()
	if err != nil {
		return nil, err
	}
	return c.icRepo.FindOneClaimBySchemaHash(ctx, c.storage.Pgx, did, string(authHash))
}

func (c *claim) guardCreateClaimRequest(req *ports.CreateClaimRequest) error {
	if _, err := url.ParseRequestURI(req.SchemaURL); err != nil {
		return ErrMalformedURL
	}
	return nil
}

func (c *claim) newVerifiableCredential(claimReq *ports.CreateClaimRequest, jsonLdContext string, nonce domain.RevNonceUint64) (verifiable.W3CCredential, error) {
	credentialCtx := []string{verifiable.JSONLDSchemaW3CCredential2018, verifiable.JSONLDSchemaIden3Credential, jsonLdContext}
	credentialType := []string{verifiable.TypeW3CVerifiableCredential, claimReq.Type}

	credentialSubject := claimReq.CredentialSubject

	if idSubject, ok := credentialSubject["id"].(string); ok {
		did, err := core.ParseDID(idSubject)
		if err != nil {
			return verifiable.W3CCredential{}, err
		}
		credentialSubject["id"] = did.String()
	}

	credentialSubject["type"] = claimReq.Type

	vcID, err := uuid.NewUUID()
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	cs := c.getRevocationSource(claimReq.DID, nonce)

	issuanceDate := time.Now()
	return verifiable.W3CCredential{
		ID:                fmt.Sprintf("%s/api/v1/claim/%s", strings.TrimSuffix(c.cfg.Host, "/"), vcID),
		Context:           credentialCtx,
		Type:              credentialType,
		Expiration:        claimReq.Expiration,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: credentialSubject,
		Issuer:            claimReq.DID.String(),
		CredentialSchema: verifiable.CredentialSchema{
			ID:   claimReq.SchemaURL,
			Type: verifiable.JSONSchemaValidator2018,
		},
		CredentialStatus: cs,
	}, nil
}

func (c *claim) getRevocationSource(issuerDID core.DID, nonce domain.RevNonceUint64) interface{} {
	if c.cfg.RHSEnabled {
		return &verifiable.RHSCredentialStatus{
			ID:              fmt.Sprintf("%s/node", strings.TrimSuffix(c.cfg.RHSUrl, "/")),
			Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
			RevocationNonce: uint64(nonce),
			StatusIssuer: &verifiable.CredentialStatus{
				ID:              buildRevocationURL(c.cfg.Host, issuerDID.String(), nonce),
				Type:            verifiable.SparseMerkleTreeProof,
				RevocationNonce: uint64(nonce),
			},
		}
	}
	return &verifiable.CredentialStatus{
		ID:              buildRevocationURL(c.cfg.Host, issuerDID.String(), nonce),
		Type:            verifiable.SparseMerkleTreeProof,
		RevocationNonce: uint64(nonce),
	}
}

func buildRevocationURL(host, issuerDID string, nonce domain.RevNonceUint64) string {
	return fmt.Sprintf("%s/api/v1/identities/%s/claims/revocation/status/%d",
		host, url.QueryEscape(issuerDID), nonce)
}

// defineMerklizedRootPosition define merkle root position for claim
// If Serialization is available in metadata of schema, position is empty, claim should not be merklized
// If metadata is empty:
// default merklized position is `index`
// otherwise value from `position`
func defineMerklizedRootPosition(metadata *jsonSuite.SchemaMetadata, position string) string {
	if metadata != nil && metadata.Serialization != nil {
		return ""
	}

	if position != "" {
		return position
	}

	return utils.MerklizedRootPositionIndex
}
