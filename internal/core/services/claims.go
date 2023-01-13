package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	"github.com/polygonid/sh-id-platform/pkg/rand"
)

var (
	// ErrJSONLdContext Field jsonLdContext must be a string
	ErrJSONLdContext = errors.New("jsonLdContext must be a string")
	ErrProcessSchema = errors.New("cannot process schema") // ErrProcessSchema Cannot process schema
)

type claim struct {
	RHSEnabled  bool
	RHSUrl      string
	Host        string
	icRepo      ports.ClaimsRepository
	schemaSrv   ports.SchemaService
	identitySrv ports.IndentityService
	storage     *db.Storage
}

// NewClaim creates a new claim service
func NewClaim(rhsEnabled bool, rhsUrl string, host string, repo ports.ClaimsRepository, schemaSrv ports.SchemaService, idenSrv ports.IndentityService, storage *db.Storage) ports.ClaimsService {
	return &claim{
		RHSEnabled:  rhsEnabled,
		RHSUrl:      rhsUrl,
		Host:        host,
		icRepo:      repo,
		schemaSrv:   schemaSrv,
		identitySrv: idenSrv,
		storage:     storage,
	}
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
		ID:                fmt.Sprintf("%s/api/v1/claim/%s", strings.TrimSuffix(c.Host, "/"), vcID),
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
	if c.RHSEnabled {
		return &verifiable.RHSCredentialStatus{
			ID:              fmt.Sprintf("%s/node", strings.TrimSuffix(c.RHSUrl, "/")),
			Type:            verifiable.Iden3ReverseSparseMerkleTreeProof,
			RevocationNonce: nonce,
			StatusIssuer: &verifiable.CredentialStatus{
				ID:              buildRevocationURL(c.Host, issuerDID, nonce),
				Type:            verifiable.SparseMerkleTreeProof,
				RevocationNonce: nonce,
			},
		}
	}
	return &verifiable.CredentialStatus{
		ID:              buildRevocationURL(c.Host, issuerDID, nonce),
		Type:            verifiable.SparseMerkleTreeProof,
		RevocationNonce: nonce,
	}
}

func buildRevocationURL(host, issuerDID string, nonce uint64) string {
	return fmt.Sprintf("%s/api/v1/identities/%s/claims/revocation/status/%d",
		host, url.QueryEscape(issuerDID), nonce)
}
