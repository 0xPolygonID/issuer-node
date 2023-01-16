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
	"github.com/iden3/go-schema-processor/processor"
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/rand"
)

var (
	ErrJSONLdContext = errors.New("jsonLdContext must be a string") // ErrJSONLdContext Field jsonLdContext must be a string
	ErrProcessSchema = errors.New("cannot process schema")          // ErrProcessSchema Cannot process schema
	ErrClaimNotFound = errors.New("claim not found")                // ErrClaimNotFound Cannot retrieve the given claim 	// ErrProcessSchema Cannot process schema
)

// ClaimCfg claim service configuration
type ClaimCfg struct {
	RHSEnabled bool // ReverseHash Enabled
	RHSUrl     string
	Host       string
}

type claim struct {
	cfg         ClaimCfg
	icRepo      ports.ClaimsRepository
	schemaSrv   ports.SchemaService
	identitySrv ports.IndentityService
	mtService   ports.MtService
	storage     *db.Storage
}

// NewClaim creates a new claim service
func NewClaim(repo ports.ClaimsRepository, schemaSrv ports.SchemaService, idenSrv ports.IndentityService, mtService ports.MtService, storage *db.Storage, cfg ClaimCfg) ports.ClaimsService {
	s := &claim{
		cfg: ClaimCfg{
			RHSEnabled: cfg.RHSEnabled,
			RHSUrl:     cfg.RHSUrl,
			Host:       cfg.Host,
		},
		icRepo:      repo,
		schemaSrv:   schemaSrv,
		identitySrv: idenSrv,
		mtService:   mtService,
		storage:     storage,
	}
	return s
}

func (c *claim) CreateClaim(ctx context.Context, req *ports.ClaimRequest) (*domain.Claim, error) {
	nonce, err := rand.Int64()
	if err != nil {
		log.Error(ctx, "create a nonce", err)
		return nil, err
	}

	schema, err := c.schemaSrv.LoadSchema(ctx, req.Schema)
	if err != nil {
		log.Error(ctx, "loading schema", err)
		return nil, err
	}

	jsonLdContext, ok := schema.Metadata.Uris["jsonLdContext"].(string)
	if !ok {
		log.Error(ctx, "invalid jsonLdContext", ErrJSONLdContext)
		return nil, ErrJSONLdContext
	}

	vc, err := c.createVC(req, jsonLdContext, nonce)
	if err != nil {
		log.Error(ctx, "creating verifiable credential", err)
		return nil, err
	}

	credentialType := fmt.Sprintf("%s#%s", jsonLdContext, req.Type)
	mtRootPostion := common.DefineMerklizedRootPosition(schema.Metadata, req.MerklizedRootPosition)

	coreClaim, err := c.schemaSrv.Process(ctx, req.CredentialSchema, credentialType, vc, &processor.CoreClaimOptions{
		RevNonce:              nonce,
		MerklizedRootPosition: mtRootPostion,
		Version:               req.Version,
		SubjectPosition:       req.SubjectPos,
		Updatable:             false,
	})
	if err != nil {
		log.Error(ctx, "Can not process the schema", err)
		return nil, ErrProcessSchema
	}

	claim, err := domain.FromClaimer(coreClaim, req.CredentialSchema, credentialType)
	if err != nil {
		log.Error(ctx, "Can not obtain the claim from claimer", err)
		return nil, err
	}

	authClaim, err := c.getAuthClaim(ctx, req.DID)
	if err != nil {
		log.Error(ctx, "Can not retrieve the auth claim", err)
		return nil, err
	}

	proof, err := c.identitySrv.SignClaimEntry(ctx, authClaim,
		coreClaim)
	if err != nil {
		log.Error(ctx, "Can not sign claim entry", err)
		return nil, err
	}

	issuerDIDString := req.DID.String()
	claim.Identifier = &issuerDIDString
	claim.Issuer = issuerDIDString

	proof.IssuerData.CredentialStatus = c.getRevocationSource(issuerDIDString, uint64(authClaim.RevNonce))

	jsonSignatureProof, err := json.Marshal(proof)
	if err != nil {
		log.Error(ctx, "Can not encode the json signature proof", err)
		return nil, err
	}
	err = claim.SignatureProof.Set(jsonSignatureProof)
	if err != nil {
		log.Error(ctx, "Can not set the json signature proof", err)
		return nil, err
	}

	err = claim.Data.Set(vc)
	if err != nil {
		log.Error(ctx, "Can not set the credential", err)
		return nil, err
	}

	err = claim.CredentialStatus.Set(vc.CredentialStatus)
	if err != nil {
		log.Error(ctx, "Can not set the credential status", err)
		return nil, err
	}

	claimResp, err := c.save(ctx, claim)
	if err != nil {
		log.Error(ctx, "Can not save the claim", err)
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
		return fmt.Errorf("error gettting merkles trees: %w", err)
	}

	err = identityTrees.RevokeClaim(ctx, rID)
	if err != nil {
		return fmt.Errorf("error revoking the claim: %w", err)
	}

	var claim *domain.Claim
	claim, err = c.icRepo.GetByRevocationNonce(ctx, c.storage.Pgx, did, domain.RevNonceUint64(nonce))

	switch {
	case errors.Is(err, repositories.ErrClaimDoesNotExist):
		// Claim does not exist. No need to update it.
		return err
	case err != nil:
		return fmt.Errorf("error getting the claim by revocation nonce: %w", err)
	default:
		claim.Revoked = true
		_, err = c.icRepo.Save(ctx, c.storage.Pgx, claim)
		if err != nil {
			return fmt.Errorf("error saving the claim: %w", err)
		}
	}

	return c.icRepo.RevokeNonce(ctx, c.storage.Pgx, &revocation)
}

func (c *claim) GetByID(ctx context.Context, issID *core.DID, id uuid.UUID) (*verifiable.W3CCredential, error) {
	claim, err := c.icRepo.GetByID(ctx, c.storage.Pgx, issID, id)
	switch {
	case errors.Is(err, repositories.ErrClaimDoesNotExist):
		return nil, ErrClaimNotFound
	case err != nil:
		return nil, err
	}

	return c.schemaSrv.FromClaimModelToW3CCredential(*claim)
}

func (c *claim) createVC(claimReq *ports.ClaimRequest, jsonLdContext string, nonce uint64) (verifiable.W3CCredential, error) {
	if err := claimReq.Validate(); err != nil {
		return verifiable.W3CCredential{}, err
	}

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

func (c *claim) getAuthClaim(ctx context.Context, did *core.DID) (*domain.Claim, error) {
	authHash, err := core.AuthSchemaHash.MarshalText()
	if err != nil {
		return nil, err
	}
	return c.icRepo.FindOneClaimBySchemaHash(ctx, c.storage.Pgx, did, string(authHash))
}

func (c *claim) newVerifiableCredential(claimReq *ports.ClaimRequest, jsonLdContext string, nonce uint64) (verifiable.W3CCredential, error) {
	credentialCtx := []string{verifiable.JSONLDSchemaW3CCredential2018, verifiable.JSONLDSchemaIden3Credential, jsonLdContext}
	credentialType := []string{verifiable.TypeW3CVerifiableCredential, claimReq.Type}

	var credentialSubject map[string]interface{}

	if err := json.Unmarshal(claimReq.CredentialSubject, &credentialSubject); err != nil {
		return verifiable.W3CCredential{}, err
	}

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

	cs := c.getRevocationSource(claimReq.DID.String(), nonce)

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
			ID:   claimReq.CredentialSchema,
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
	return fmt.Sprintf("%s/api/v1/identities/%s/claims/revocation/status/%d",
		host, url.QueryEscape(issuerDID), nonce)
}
