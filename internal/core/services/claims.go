package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/loader"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/rand"
	schemaPkg "github.com/polygonid/sh-id-platform/pkg/schema"
)

var (
	ErrClaimNotFound  = errors.New("claim not found")                // ErrClaimNotFound Cannot retrieve the given claim
	ErrSchemaNotFound = errors.New("claim not found")                // ErrSchemaNotFound Cannot retrieve the given schema from DB
	ErrJSONLdContext  = errors.New("jsonLdContext must be a string") // ErrJSONLdContext Field jsonLdContext must be a string
	ErrLoadingSchema  = errors.New("cannot load schema")             // ErrLoadingSchema means the system cannot load the schema file
	ErrMalformedURL   = errors.New("malformed url")                  // ErrMalformedURL The schema url is wrong
	ErrProcessSchema  = errors.New("cannot process schema")          // ErrProcessSchema Cannot process schema
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
	loaderFactory           loader.Factory
}

// NewClaim creates a new claim service
func NewClaim(repo ports.ClaimsRepository, idenSrv ports.IdentityService, mtService ports.MtService, identityStateRepository ports.IdentityStateRepository, ld loader.Factory, storage *db.Storage, cfg ClaimCfg) ports.ClaimsService {
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
		loaderFactory:           ld,
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
		log.Error(ctx, "create a nonce", "err", err)
		return nil, err
	}

	schema, err := schemaPkg.LoadSchema(ctx, c.loaderFactory(req.Schema))
	if err != nil {
		log.Error(ctx, "loading schema", "err", err, "schema", req.Schema)
		return nil, ErrLoadingSchema
	}

	jsonLdContext, ok := schema.Metadata.Uris["jsonLdContext"].(string)
	if !ok {
		log.Error(ctx, "invalid jsonLdContext", "err", ErrJSONLdContext)
		return nil, ErrJSONLdContext
	}

	vcID, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	vc, err := c.createVC(req, vcID, jsonLdContext, nonce)
	if err != nil {
		log.Error(ctx, "creating verifiable credential", "err", err)
		return nil, err
	}

	credentialType := fmt.Sprintf("%s#%s", jsonLdContext, req.Type)
	mtRootPostion := common.DefineMerklizedRootPosition(schema.Metadata, req.MerklizedRootPosition)

	coreClaim, err := schemaPkg.Process(ctx, c.loaderFactory(req.Schema), credentialType, vc, &processor.CoreClaimOptions{
		RevNonce:              nonce,
		MerklizedRootPosition: mtRootPostion,
		Version:               req.Version,
		SubjectPosition:       req.SubjectPos,
		Updatable:             false,
	})
	if err != nil {
		log.Error(ctx, "Can not process the schema", "err", err)
		return nil, ErrProcessSchema
	}

	claim, err := domain.FromClaimer(coreClaim, req.Schema, credentialType)
	if err != nil {
		log.Error(ctx, "Can not obtain the claim from claimer", "err", err)
		return nil, err
	}

	issuerDIDString := req.DID.String()
	claim.Identifier = &issuerDIDString
	claim.Issuer = issuerDIDString
	claim.ID = vcID

	if req.SignatureProof {
		authClaim, err := c.GetAuthClaim(ctx, req.DID)
		if err != nil {
			log.Error(ctx, "Can not retrieve the auth claim", "err", err)
			return nil, err
		}

		proof, err := c.identitySrv.SignClaimEntry(ctx, authClaim, coreClaim)
		if err != nil {
			log.Error(ctx, "Can not sign claim entry", "err", err)
			return nil, err
		}

		proof.IssuerData.CredentialStatus = c.getRevocationSource(issuerDIDString, uint64(authClaim.RevNonce))

		jsonSignatureProof, err := json.Marshal(proof)
		if err != nil {
			log.Error(ctx, "Can not encode the json signature proof", "err", err)
			return nil, err
		}
		err = claim.SignatureProof.Set(jsonSignatureProof)
		if err != nil {
			log.Error(ctx, "Can not set the json signature proof", "err", err)
			return nil, err
		}
	}

	err = claim.Data.Set(vc)
	if err != nil {
		log.Error(ctx, "Can not set the credential", "err", err)
		return nil, err
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	if err != nil {
		log.Error(ctx, "Can not set the credential status", "err", err)
		return nil, err
	}

	claim.MtProof = req.MTProof
	claimResp, err := c.save(ctx, claim)
	if err != nil {
		log.Error(ctx, "Can not save the claim", "err", err)
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

func (c *claim) Delete(ctx context.Context, id uuid.UUID) error {
	err := c.icRepo.Delete(ctx, c.storage.Pgx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return ErrClaimNotFound
		}
		return err
	}

	return nil
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

func (c *claim) Agent(ctx context.Context, req *ports.AgentRequest) (*domain.Agent, error) {
	exists, err := c.identitySrv.Exists(ctx, *req.IssuerDID)
	if err != nil {
		log.Error(ctx, "loading issuer identity", "err", err, "issuerDID", req.IssuerDID)
		return nil, err
	}

	if !exists {
		log.Warn(ctx, "issuer not found", "issuerDID", req.IssuerDID)
		return nil, fmt.Errorf("cannot proceed with this identity, not found")
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

func (c *claim) GetAll(ctx context.Context, did core.DID, filter *ports.ClaimsFilter) ([]*domain.Claim, error) {
	claims, err := c.icRepo.GetAllByIssuerID(ctx, c.storage.Pgx, did, filter)
	if err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return nil, ErrClaimNotFound
		}
		return nil, err
	}

	return claims, nil
}

func (c *claim) getAgentCredential(ctx context.Context, basicMessage *ports.AgentRequest) (*domain.Agent, error) {
	fetchRequestBody := &protocol.CredentialFetchRequestMessageBody{}
	err := json.Unmarshal(basicMessage.Body, fetchRequestBody)
	if err != nil {
		log.Error(ctx, "unmarshalling agent body", err)
		return nil, fmt.Errorf("invalid credential fetch request body: %w", err)
	}

	claimID, err := uuid.Parse(fetchRequestBody.ID)
	if err != nil {
		log.Error(ctx, "wrong claimID in agent request body", err)
		return nil, fmt.Errorf("invalid claim ID")
	}

	claim, err := c.icRepo.GetByIdAndIssuer(ctx, c.storage.Pgx, basicMessage.IssuerDID, claimID)
	if err != nil {
		log.Error(ctx, "loading claim", err, "claimID", claim.ID)
		return nil, fmt.Errorf("failed get claim by claimID: %w", err)
	}
	if claim.OtherIdentifier != basicMessage.UserDID.String() {
		err := fmt.Errorf("claim doesn't relate to sender")
		log.Error(ctx, "claim doesn't relate to sender", err, "claimID", claim.ID)
		return nil, err
	}

	vc, err := schemaPkg.FromClaimModelToW3CCredential(*claim)
	if err != nil {
		log.Error(ctx, "creating W3 credential", err)
		return nil, fmt.Errorf("failed to convert claim to  w3cCredential: %w", err)
	}

	return &domain.Agent{
		ID:       uuid.NewString(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialIssuanceResponseMessageType,
		ThreadID: basicMessage.ThreadID,
		Body:     protocol.IssuanceMessageBody{Credential: *vc},
		From:     basicMessage.IssuerDID.String(),
		To:       basicMessage.UserDID.String(),
	}, err
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

func (c *claim) createVC(claimReq *ports.CreateClaimRequest, vcID uuid.UUID, jsonLdContext string, nonce uint64) (verifiable.W3CCredential, error) {
	vCredential, err := c.newVerifiableCredential(claimReq, vcID, jsonLdContext, nonce) // create vc credential
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

func (c *claim) guardCreateClaimRequest(req *ports.CreateClaimRequest) error {
	if _, err := url.ParseRequestURI(req.Schema); err != nil {
		return ErrMalformedURL
	}
	return nil
}

func (c *claim) newVerifiableCredential(claimReq *ports.CreateClaimRequest, vcID uuid.UUID, jsonLdContext string, nonce uint64) (verifiable.W3CCredential, error) {
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

	cs := c.getRevocationSource(claimReq.DID.String(), nonce)

	issuanceDate := time.Now()
	return verifiable.W3CCredential{
		ID:                fmt.Sprintf("%s/v1/%s/claims/%s", strings.TrimSuffix(c.cfg.Host, "/"), claimReq.DID.String(), vcID),
		Context:           credentialCtx,
		Type:              credentialType,
		Expiration:        claimReq.Expiration,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: credentialSubject,
		Issuer:            claimReq.DID.String(),
		CredentialSchema: verifiable.CredentialSchema{
			ID:   claimReq.Schema,
			Type: verifiable.JSONSchemaValidator2018,
		},
		CredentialStatus: cs,
	}, nil
}

func (c *claim) getRevocationSource(issuerDID string, nonce uint64) interface{} {
	if c.cfg.RHSEnabled {
		return &verifiable.RHSCredentialStatus{
			ID:              fmt.Sprintf("%s/node", strings.TrimSuffix(c.cfg.RHSUrl, "/")),
			Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
			RevocationNonce: nonce,
			StatusIssuer: &verifiable.CredentialStatus{
				ID:              buildRevocationURL(c.cfg.Host, issuerDID, nonce),
				Type:            verifiable.SparseMerkleTreeProof,
				RevocationNonce: nonce,
			},
		}
	}
	return &verifiable.CredentialStatus{
		ID:              buildRevocationURL(c.cfg.Host, issuerDID, nonce),
		Type:            verifiable.SparseMerkleTreeProof,
		RevocationNonce: nonce,
	}
}

func buildRevocationURL(host, issuerDID string, nonce uint64) string {
	return fmt.Sprintf("%s/v1/%s/claims/revocation/status/%d",
		host, url.QueryEscape(issuerDID), nonce)
}

func (c *claim) GetAuthClaimForPublishing(ctx context.Context, did *core.DID, state string) (*domain.Claim, error) {
	authHash, err := core.AuthSchemaHash.MarshalText()
	if err != nil {
		return nil, err
	}

	validAuthClaims, err := c.icRepo.GetAuthClaimsForPublishing(ctx, c.storage.Pgx, did, state, string(authHash))
	if err != nil {
		return nil, err
	}
	if len(validAuthClaims) == 0 {
		return nil, errors.New("no auth claims for publishing")
	}

	return validAuthClaims[0], nil
}

// UpdateClaimsMTPAndState update identity status and claim MTP
func (c *claim) UpdateClaimsMTPAndState(ctx context.Context, currentState *domain.IdentityState) error {
	did, err := core.ParseDID(currentState.Identifier)
	if err != nil {
		return err
	}

	iTrees, err := c.mtService.GetIdentityMerkleTrees(ctx, c.storage.Pgx, did)
	if err != nil {
		return err
	}

	claimsTree, err := iTrees.ClaimsTree()
	if err != nil {
		return err
	}

	currState, err := merkletree.NewHashFromHex(*currentState.State)
	if err != nil {
		return err
	}

	claims, err := c.icRepo.GetAllByStateWithMTProof(ctx, c.storage.Pgx, did, currState)
	if err != nil {
		return err
	}

	for i := range claims {
		var index *big.Int
		var coreClaimHex string
		coreClaim := claims[i].CoreClaim.Get()
		index, err = coreClaim.HIndex()
		if err != nil {
			return err
		}
		var proof *merkletree.Proof
		proof, _, err = claimsTree.GenerateProof(ctx, index, claimsTree.Root())
		if err != nil {
			return err
		}
		coreClaimHex, err = coreClaim.Hex()
		if err != nil {
			return err
		}
		mtpProof := verifiable.Iden3SparseMerkleTreeProof{
			Type: verifiable.Iden3SparseMerkleTreeProofType,
			IssuerData: verifiable.IssuerData{
				ID: did.String(),
				State: verifiable.State{
					RootOfRoots:        currentState.RootOfRoots,
					ClaimsTreeRoot:     currentState.ClaimsTreeRoot,
					RevocationTreeRoot: currentState.RevocationTreeRoot,
					Value:              currentState.State,
					BlockTimestamp:     currentState.BlockTimestamp,
					TxID:               currentState.TxID,
					BlockNumber:        currentState.BlockNumber,
				},
			},
			CoreClaim: coreClaimHex,
			MTP:       proof,
		}

		var jsonProof []byte
		jsonProof, err = json.Marshal(mtpProof)
		if err != nil {
			return fmt.Errorf("can't marshal proof: %w", err)
		}

		var affected int64
		err = claims[i].MTPProof.Set(jsonProof)
		if err != nil {
			return fmt.Errorf("failed set mtp proof: %w", err)
		}
		affected, err = c.icRepo.UpdateClaimMTP(ctx, c.storage.Pgx, &claims[i])

		if err != nil {
			return fmt.Errorf("can't update claim mtp:  %w", err)
		}
		if affected == 0 {
			return fmt.Errorf("claim has not been updated %v", claims[i])
		}
	}
	_, err = c.identityStateRepository.UpdateState(ctx, c.storage.Pgx, currentState)
	if err != nil {
		return fmt.Errorf("can't update identity state: %w", err)
	}

	return nil
}
