package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// CredentialOfferMessageType - TODO
const CredentialOfferMessageType string = "https://iden3-communication.io/credentials/1.0/offer"

// CredentialLink is structure to fetch credential
type CredentialLink struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// CredentialsLinkMessageBody is struct the represents offer message
type CredentialsLinkMessageBody struct {
	URL         string           `json:"url"`
	Credentials []CredentialLink `json:"credentials"`
}

// LinkQRCodeMessage - TODO
type LinkQRCodeMessage struct {
	ID       string                     `json:"id"`
	Typ      string                     `json:"typ,omitempty"`
	Type     string                     `json:"type"`
	ThreadID string                     `json:"thid,omitempty"`
	Body     CredentialsLinkMessageBody `json:"body,omitempty"`
	From     string                     `json:"from,omitempty"`
	To       string                     `json:"to,omitempty"`
}

// LinkService - the interface that defines the available methods
type LinkService interface {
	Save(ctx context.Context, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID,
		credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error)
	Activate(ctx context.Context, linkID uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID, did core.DID) error
	CreateQRCode(ctx context.Context, issuerDID core.DID, linkID uuid.UUID, serverURL string) (*protocol.AuthorizationRequestMessage, string, error)
	IssueClaim(ctx context.Context, sessionID string, issuerDID core.DID, userDID core.DID, linkID uuid.UUID, hostURL string) error
}
