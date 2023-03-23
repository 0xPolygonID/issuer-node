package ports

import (
	"context"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

// LinkService - the interface that defines the available methods
type LinkService interface {
	Save(ctx context.Context, ID uuid.UUID, did core.DID, maxIssuance *int, validUntil *time.Time, schemaID uuid.UUID,
		credentialExpiration *time.Time, credentialSignatureProof bool, credentialMTPProof bool, credentialAttributes []domain.CredentialAttributes) (*domain.Link, error)
}
