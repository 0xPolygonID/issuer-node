package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	linkState "github.com/polygonid/sh-id-platform/pkg/link"
)

type CreateQRCodeResponse struct {
	Link      *domain.Link
	QrCode    *protocol.AuthorizationRequestMessage
	SessionID string
}

// LinkService - the interface that defines the available methods
type LinkService interface {
	Save(ctx context.Context, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID, credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttrsRequest) (*domain.Link, error)
	Activate(ctx context.Context, issuerID core.DID, linkID uuid.UUID, active bool) error
	Delete(ctx context.Context, id uuid.UUID, did core.DID) error
	GetByID(ctx context.Context, issuerID core.DID, id uuid.UUID) (*domain.Link, error)
	CreateQRCode(ctx context.Context, issuerDID core.DID, linkID uuid.UUID, serverURL string) (*CreateQRCodeResponse, error)
	IssueClaim(ctx context.Context, sessionID string, issuerDID core.DID, userDID core.DID, linkID uuid.UUID, hostURL string) error
	GetQRCode(ctx context.Context, sessionID uuid.UUID, linkID uuid.UUID) (*linkState.State, error)
}
