package repositories

import (
	"context"

	"github.com/google/uuid"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type link struct {
	conn db.Storage
}

// NewLink returns a new connections repository
func NewLink(conn db.Storage) ports.LinkRepository {
	return &link{
		conn,
	}
}

func (l link) Save(ctx context.Context, link *domain.Link) (uuid.UUID, error) {
	var id uuid.UUID
	sql := `INSERT INTO links (issuer_id, max_issuance, valid_until, schema_id, credential_expiration, credential_signature_proof, credential_mtp_proof, credential_attributes, active)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)
			RETURNING id`
	err := l.conn.Pgx.QueryRow(ctx, sql, link.IssuerCoreDID().String(), link.MaxIssuance, link.ValidUntil, link.SchemaID, link.CredentialExpiration, link.CredentialSignatureProof,
		link.CredentialMTPProof, link.CredentialAttributesString(), link.Active).Scan(&id)

	return id, err
}

func (l link) GetByID(ctx context.Context, id uuid.UUID) (*domain.Link, error) {
	link := domain.Link{}
	var schemasAttributtes string
	sql := `SELECT * FROM links 
			WHERE id = $1`
	err := l.conn.Pgx.QueryRow(ctx, sql, id).
		Scan(&link.ID, &link.IssuerDID, &link.CreatedAt, &link.MaxIssuance, &link.ValidUntil, &link.SchemaID, &link.CredentialExpiration,
			&link.CredentialSignatureProof, &link.CredentialMTPProof, &schemasAttributtes, &link.Active)

	link.StrToCredentialAttributes(schemasAttributtes)
	return &link, err
}
