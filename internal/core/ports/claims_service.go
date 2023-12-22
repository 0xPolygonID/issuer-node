package ports

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/sqltools"
)

// CreateClaimRequest struct
type CreateClaimRequest struct {
	DID                   *w3c.DID
	Schema                string
	CredentialSubject     map[string]any
	Expiration            *time.Time
	Type                  string
	Version               uint32
	SubjectPos            string
	MerklizedRootPosition string
	SignatureProof        bool
	MTProof               bool
	LinkID                *uuid.UUID
	SingleIssuer          bool
	CredentialStatusType  verifiable.CredentialStatusType
	SchemaTypeDescription string
	RefreshService        *verifiable.RefreshService
	RevNonce              *uint64
}

// AgentRequest struct
type AgentRequest struct {
	Body      json.RawMessage
	ThreadID  string
	IssuerDID *w3c.DID
	UserDID   *w3c.DID
	ClaimID   uuid.UUID
	Typ       comm.MediaType
	Type      comm.ProtocolMessage
}

// Defines values for GetCredentialsParamsStatus.
// TIP: Use the sql field name in these constants. A little bit coupled but easy to construct the ORDER BY clause later
const (
	CredentialSchemaType sqltools.SQLFieldName = "claims.schema_type"
	CredentialCreatedAt  sqltools.SQLFieldName = "claims.created_at"
	CredentialExpiresAt  sqltools.SQLFieldName = "claims.expiration"
	CredentialRevoked    sqltools.SQLFieldName = "claims.revoked"
)

// ClaimsFilter struct
type ClaimsFilter struct {
	Self            *bool
	Revoked         *bool
	ExpiredOn       *time.Time
	SchemaHash      string
	SchemaType      string
	Subject         string
	QueryField      string
	QueryFieldValue string
	FTSQuery        string
	FTSAndCond      bool
	Proofs          []verifiable.ProofType
	MaxResults      int  // Max number of results to return on each call.
	Page            *int // Page number to return. First is 1. if nul, then there is no limit in the number to return
	OrderBy         sqltools.OrderByFilters
}

// NewClaimsFilter returns a valid claims filter
func NewClaimsFilter(schemaHash, schemaType, subject, queryField, queryValue *string, self, revoked *bool) (*ClaimsFilter, error) {
	var filter ClaimsFilter

	if self != nil && *self {
		if subject != nil && *subject != "" {
			return nil, fmt.Errorf("self and subject filter cannot be used together")
		}
		filter.Self = self
	}
	if schemaHash != nil {
		filter.SchemaHash = *schemaHash
	}
	if schemaType != nil {
		filter.SchemaType = *schemaType
	}
	if revoked != nil {
		filter.Revoked = revoked
	}
	if subject != nil {
		filter.Subject = *subject
	}
	if queryField != nil {
		filter.QueryField = *queryField
	}
	if queryValue != nil {
		filter.QueryFieldValue = *queryValue
	}

	return &filter, nil
}

// NewCreateClaimRequest returns a new claim object with the given parameters
func NewCreateClaimRequest(did *w3c.DID, credentialSchema string, credentialSubject map[string]any, expiration *time.Time, typ string, cVersion *uint32, subjectPos *string, merklizedRootPosition *string, sigProof *bool, mtProof *bool, linkID *uuid.UUID, singleIssuer bool, credentialStatusType verifiable.CredentialStatusType, refreshService *verifiable.RefreshService, revNonce *uint64) *CreateClaimRequest {
	if sigProof == nil {
		sigProof = common.ToPointer(false)
	}

	if mtProof == nil {
		mtProof = common.ToPointer(false)
	}

	req := &CreateClaimRequest{
		DID:               did,
		Schema:            credentialSchema,
		CredentialSubject: credentialSubject,
		Type:              typ,
		SignatureProof:    *sigProof,
		MTProof:           *mtProof,
		RefreshService:    refreshService,
	}
	if expiration != nil {
		req.Expiration = expiration
	}
	if cVersion != nil {
		req.Version = *cVersion
	}
	if subjectPos != nil {
		req.SubjectPos = *subjectPos
	}
	if merklizedRootPosition != nil {
		req.MerklizedRootPosition = *merklizedRootPosition
	}

	req.RevNonce = revNonce
	req.LinkID = linkID
	req.SingleIssuer = singleIssuer
	req.CredentialStatusType = credentialStatusType

	return req
}

// NewAgentRequest validates the inputs and returns a new AgentRequest
func NewAgentRequest(basicMessage *comm.BasicMessage) (*AgentRequest, error) {
	if basicMessage.To == "" {
		return nil, fmt.Errorf("'to' field cannot be empty")
	}

	toDID, err := w3c.ParseDID(basicMessage.To)
	if err != nil {
		return nil, err
	}

	if basicMessage.From == "" {
		return nil, fmt.Errorf("'from' field cannot be empty")
	}

	fromDID, err := w3c.ParseDID(basicMessage.From)
	if err != nil {
		return nil, err
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field cannot be empty")
	}

	claimID, err := uuid.Parse(basicMessage.ID)
	if err != nil {
		return nil, err
	}

	if basicMessage.Type != protocol.CredentialFetchRequestMessageType && basicMessage.Type != protocol.RevocationStatusRequestMessageType {
		return nil, fmt.Errorf("invalid type")
	}

	if basicMessage.ID == "" {
		return nil, fmt.Errorf("'id' field cannot be empty")
	}

	return &AgentRequest{
		Body:      basicMessage.Body,
		UserDID:   fromDID,
		IssuerDID: toDID,
		ThreadID:  basicMessage.ThreadID,
		ClaimID:   claimID,
		Typ:       basicMessage.Typ,
		Type:      basicMessage.Type,
	}, nil
}

// GetCredentialQrCodeResponse is the response of the GetCredentialQrCode method
type GetCredentialQrCodeResponse struct {
	QrCodeURL  string
	SchemaType string
	QrID       uuid.UUID
}

// ClaimsService is the interface implemented by the claim service
type ClaimsService interface {
	Save(ctx context.Context, claimReq *CreateClaimRequest) (*domain.Claim, error)
	CreateCredential(ctx context.Context, req *CreateClaimRequest) (*domain.Claim, error)
	Revoke(ctx context.Context, id w3c.DID, nonce uint64, description string) error
	GetAll(ctx context.Context, did w3c.DID, filter *ClaimsFilter) ([]*domain.Claim, int, error)
	RevokeAllFromConnection(ctx context.Context, connID uuid.UUID, issuerID w3c.DID) error
	GetRevocationStatus(ctx context.Context, issuerDID w3c.DID, nonce uint64) (*verifiable.RevocationStatus, error)
	GetByID(ctx context.Context, issID *w3c.DID, id uuid.UUID) (*domain.Claim, error)
	GetCredentialQrCode(ctx context.Context, issID *w3c.DID, id uuid.UUID, hostURL string) (*GetCredentialQrCodeResponse, error)
	Agent(ctx context.Context, req *AgentRequest) (*domain.Agent, error)
	GetAuthClaim(ctx context.Context, did *w3c.DID) (*domain.Claim, error)
	GetAuthClaimForPublishing(ctx context.Context, did *w3c.DID, state string) (*domain.Claim, error)
	UpdateClaimsMTPAndState(ctx context.Context, currentState *domain.IdentityState) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByStateIDWithMTPProof(ctx context.Context, did *w3c.DID, state string) ([]*domain.Claim, error)
}
