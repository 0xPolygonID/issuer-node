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
	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

type claim struct {
	RHSEnabled bool
	RHSUrl     string
	Host       string
	icRepo     ports.ClaimsRepository
	storage    *db.Storage
	mtService  ports.MtService
}

// NewClaim creates a new claim service
func NewClaim(rhsEnabled bool, rhsUrl string, host string, repo ports.ClaimsRepository, storage *db.Storage, mtService ports.MtService) ports.ClaimsService {
	return &claim{
		RHSEnabled: rhsEnabled,
		RHSUrl:     rhsUrl,
		Host:       host,
		icRepo:     repo,
		storage:    storage,
		mtService:  mtService,
	}
}

func (c *claim) CreateVC(ctx context.Context, claimReq *ports.ClaimRequest, nonce uint64) (verifiable.W3CCredential, error) {
	if err := claimReq.Validate(); err != nil {
		return verifiable.W3CCredential{}, err
	}

	vCredential, err := c.newVerifiableCredential(claimReq, nonce) // create vc credential
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	return vCredential, nil
}

func (c *claim) Save(ctx context.Context, claim *domain.Claim) (*domain.Claim, error) {
	id, err := c.icRepo.Save(ctx, c.storage.Pgx, claim)
	if err != nil {
		return nil, err
	}

	claim.ID = id

	return claim, nil
}

func (c *claim) SendClaimOfferPushNotification(ctx context.Context, claim *domain.Claim) error {
	return nil
}

func (c *claim) GetAuthClaim(ctx context.Context, did *core.DID) (*domain.Claim, error) {
	authHash, err := core.AuthSchemaHash.MarshalText()
	if err != nil {
		return nil, err
	}
	return c.icRepo.FindOneClaimBySchemaHash(ctx, c.storage.Pgx, did, string(authHash))
}

func (c *claim) newVerifiableCredential(claimReq *ports.ClaimRequest, nonce uint64) (verifiable.W3CCredential, error) {
	jsonLdContext, ok := claimReq.Schema.Metadata.Uris["jsonLdContext"].(string)
	if !ok {
		return verifiable.W3CCredential{}, fmt.Errorf("invalid jsonLdContext type, expected string")
	}
	credentialCtx := []string{verifiable.JSONLDSchemaW3CCredential2018, verifiable.JSONLDSchemaIden3Credential, jsonLdContext}
	credentialType := []string{verifiable.TypeW3CVerifiableCredential, claimReq.Type}

	var expirationTime *time.Time
	if claimReq.Expiration != nil {
		expirationTime = common.ToPointer(time.Unix(*claimReq.Expiration, 0))
	}

	issuanceDate := time.Now()

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

	cs := c.GetRevocationSource(claimReq.DID.String(), nonce)

	return verifiable.W3CCredential{
		ID:                fmt.Sprintf("%s/api/v1/claim/%s", strings.TrimSuffix(c.Host, "/"), vcID),
		Context:           credentialCtx,
		Type:              credentialType,
		Expiration:        expirationTime,
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

func (c *claim) GetRevocationSource(issuerDID string, nonce uint64) interface{} {
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

func buildRevocationURL(host, issuerDID string, nonce uint64) string {
	return fmt.Sprintf("%s/api/v1/identities/%s/claims/revocation/status/%d",
		host, url.QueryEscape(issuerDID), nonce)
}
