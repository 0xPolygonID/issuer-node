package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

var errorShemaNotFound = errors.New("schema id not found")

type link struct {
	conn db.Storage
}

// NewLink returns a new connections repository
func NewLink(conn db.Storage) ports.LinkRepository {
	return &link{
		conn,
	}
}

func (l link) Save(ctx context.Context, link *domain.Link) (*uuid.UUID, error) {
	pgAttrs := pgtype.JSONB{}
	if err := pgAttrs.Set(link.CredentialAttributes); err != nil {
		return nil, fmt.Errorf("cannot set schema attributes values: %w", err)
	}

	var id uuid.UUID
	sql := `INSERT INTO links (id, issuer_id, max_issuance, valid_until, schema_id, credential_expiration, credential_signature_proof, credential_mtp_proof, credential_attributes, active)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (id) DO
			UPDATE SET issuer_id=$2, max_issuance=$3, valid_until=$4, schema_id=$5, credential_expiration=$6, credential_signature_proof=$7, credential_mtp_proof=$8, credential_attributes=$9, active=$10 
			RETURNING id`
	err := l.conn.Pgx.QueryRow(ctx, sql, link.ID, link.IssuerCoreDID().String(), link.MaxIssuance, link.ValidUntil, link.SchemaID, link.CredentialExpiration, link.CredentialSignatureProof,
		link.CredentialMTPProof, pgAttrs, link.Active).Scan(&id)

	if err != nil && strings.Contains(err.Error(), `table "links" violates foreign key constraint "links_schemas_id_key"`) {
		return nil, errorShemaNotFound
	}
	return &id, err
}

func (l link) GetByID(ctx context.Context, id uuid.UUID) (*domain.Link, error) {
	link := domain.Link{}
	var credentialAttributtes pgtype.JSONB
	sql := `SELECT * FROM links 
			WHERE id = $1`
	err := l.conn.Pgx.QueryRow(ctx, sql, id).
		Scan(&link.ID, &link.IssuerDID, &link.CreatedAt, &link.MaxIssuance, &link.ValidUntil, &link.SchemaID, &link.CredentialExpiration,
			&link.CredentialSignatureProof, &link.CredentialMTPProof, &credentialAttributtes, &link.Active)

	if err := credentialAttributtes.AssignTo(&link.CredentialAttributes); err != nil {
		return nil, fmt.Errorf("parsing credential attributes: %w", err)
	}
	return &link, err
}
