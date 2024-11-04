package repositories

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

var (
	errorShemaNotFound = errors.New("schema id not found")

	// ErrLinkDoesNotExist link does not exist
	ErrLinkDoesNotExist = errors.New("link does not exist")

	// ErrorLinkWithClaims cannot delete link with associated claims
	ErrorLinkWithClaims = errors.New("cannot delete link with associated claims")
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

func (l link) Save(ctx context.Context, conn db.Querier, link *domain.Link) (*uuid.UUID, error) {
	pgAttrs := pgtype.JSONB{}
	if err := pgAttrs.Set(link.CredentialSubject); err != nil {
		return nil, fmt.Errorf("cannot set credential subject values: %w", err)
	}

	var id uuid.UUID
	sql := `INSERT INTO links (id, issuer_id, max_issuance, valid_until, schema_id, credential_expiration, credential_signature_proof, credential_mtp_proof, credential_attributes, active, refresh_service, display_method)
			VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) ON CONFLICT (id) DO
			UPDATE SET issuer_id=$2, max_issuance=$3, valid_until=$4, schema_id=$5, credential_expiration=$6, credential_signature_proof=$7, credential_mtp_proof=$8, credential_attributes=$9, active=$10 
			RETURNING id`
	err := conn.QueryRow(ctx, sql, link.ID, link.IssuerCoreDID().String(), link.MaxIssuance, link.ValidUntil, link.SchemaID, link.CredentialExpiration, link.CredentialSignatureProof,
		link.CredentialMTPProof, pgAttrs, link.Active, link.RefreshService, link.DisplayMethod).Scan(&id)

	if err != nil && strings.Contains(err.Error(), `table "links" violates foreign key constraint "links_schemas_id_key"`) {
		return nil, errorShemaNotFound
	}
	return &id, err
}

func (l link) GetByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) (*domain.Link, error) {
	const sql = `
SELECT links.id, 
       links.issuer_id, 
       links.created_at, 
       links.max_issuance, 
       links.valid_until, 
       links.schema_id, 
       links.credential_expiration, 
       links.credential_signature_proof,
       links.credential_mtp_proof, 
       links.credential_attributes, 
       links.active,
	   links.refresh_service,
	   links.display_method,
       count(claims.id) as issued_claims,
       links.authorization_request_message,
       schemas.id as schema_id,
       schemas.issuer_id as schema_issuer_id,
       schemas.url,
       schemas.type,
       schemas.hash,
       schemas.words, 
       schemas.created_at
FROM links
LEFT JOIN schemas ON schemas.id = links.schema_id AND schemas.issuer_id = links.issuer_id
LEFT JOIN claims ON claims.link_id = links.id AND claims.identifier = links.issuer_id
WHERE links.id = $1 AND links.issuer_id = $2
GROUP BY links.id, schemas.id 
`
	link := domain.Link{}
	s := dbSchema{}
	var credentialSubject pgtype.JSONB
	err := l.conn.Pgx.QueryRow(ctx, sql, id, issuerDID.String()).Scan(
		&link.ID,
		&link.IssuerDID,
		&link.CreatedAt,
		&link.MaxIssuance,
		&link.ValidUntil,
		&link.SchemaID,
		&link.CredentialExpiration,
		&link.CredentialSignatureProof,
		&link.CredentialMTPProof,
		&credentialSubject,
		&link.Active,
		&link.RefreshService,
		&link.DisplayMethod,
		&link.IssuedClaims,
		&link.AuthorizationRequestMessage,
		&s.ID,
		&s.IssuerID,
		&s.URL,
		&s.Type,
		&s.Hash,
		&s.Words,
		&s.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, ErrLinkDoesNotExist
	}
	if err != nil {
		return nil, err
	}

	d := json.NewDecoder(bytes.NewReader(credentialSubject.Bytes))
	d.UseNumber()
	if err := d.Decode(&link.CredentialSubject); err != nil {
		return nil, fmt.Errorf("parsing credential attributes: %w", err)
	}
	link.Schema, err = toSchemaDomain(&s)
	if err != nil {
		return nil, fmt.Errorf("parsing link schema: %w", err)
	}
	return &link, err
}

func (l link) GetAll(ctx context.Context, issuerDID w3c.DID, filter ports.LinksFilter) ([]*domain.Link, uint, error) {
	fields := []string{
		"links.id",
		"links.issuer_id",
		"links.created_at",
		"links.max_issuance",
		"links.valid_until",
		"links.schema_id",
		"links.credential_expiration",
		"links.credential_signature_proof",
		"links.credential_mtp_proof",
		"links.credential_attributes",
		"links.active",
		"links.refresh_service",
		"links.display_method",
		"links.authorization_request_message",
		"count(claims.id) as issued_claims",
		"schemas.id as schema_id",
		"schemas.issuer_id as schema_issuer_id",
		"schemas.url",
		"schemas.type",
		"schemas.hash",
		"schemas.words",
		"schemas.created_at",
	}

	//sql := `
	//SELECT links.id,
	//      links.issuer_id,
	//      links.created_at,
	//      links.max_issuance,
	//      links.valid_until,
	//      links.schema_id,
	//      links.credential_expiration,
	//      links.credential_signature_proof,
	//      links.credential_mtp_proof,
	//      links.credential_attributes,
	//      links.active,
	//	   links.refresh_service,
	//	   links.display_method,
	//	   links.authorization_request_message,
	//      count(claims.id) as issued_claims,
	//      schemas.id as schema_id,
	//      schemas.issuer_id as schema_issuer_id,
	//      schemas.url,
	//      schemas.type,
	//      schemas.hash,
	//      schemas.words,
	//      schemas.created_at
	//FROM links
	//LEFT JOIN schemas ON schemas.id = links.schema_id
	//LEFT JOIN claims ON claims.link_id = links.id AND claims.identifier = links.issuer_id
	//WHERE links.issuer_id = $1
	//`

	sql := `
	SELECT  ##QUERYFIELDS##
	FROM links
	LEFT JOIN schemas ON schemas.id = links.schema_id
	LEFT JOIN claims ON claims.link_id = links.id AND claims.identifier = links.issuer_id
	WHERE links.issuer_id = $1
	`

	sqlArgs := make([]interface{}, 0)
	sqlArgs = append(sqlArgs, issuerDID.String(), time.Now())

	switch filter.Status {
	case ports.LinkActive:
		sql += " AND links.active AND coalesce(links.valid_until > $2, true) AND coalesce(links.max_issuance>(SELECT count(claims.id) FROM claims where claims.link_id = links.id), true)"
	case ports.LinkInactive:
		sql += " AND NOT links.active"
	case ports.LinkExceeded:
		sql += " AND " +
			"(links.valid_until IS NOT NULL AND links.valid_until<= $2) " +
			"OR " +
			"(links.max_issuance IS NOT NULL AND links.max_issuance <= (SELECT count(claims.id) FROM claims where claims.link_id = links.id))"
	}
	if filter.Query != nil && *filter.Query != "" {
		terms := tokenizeQuery(*filter.Query)
		sql += " AND (" + buildPartialQueryLikes("schemas.words", "OR", 1+len(sqlArgs), len(terms)) + ")"
		for _, term := range terms {
			sqlArgs = append(sqlArgs, term)
		}
	}

	// Dummy condition to include time in the query although not always used
	sql += " AND (true OR $1::text IS NULL OR $2::text IS NULl)"
	sql += " GROUP BY links.id, schemas.id"
	sql += " ORDER BY links.created_at DESC"

	countInnerQuery := strings.Replace(sql, "##QUERYFIELDS##", "links.id", 1)
	countQuery := `SELECT count(*) FROM (` + countInnerQuery + `) as count`

	var count uint
	if err := l.conn.Pgx.QueryRow(ctx, countQuery, sqlArgs...).Scan(&count); err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	if filter.Page != nil {
		sql += fmt.Sprintf(" OFFSET %d LIMIT %d;", (*filter.Page-1)*filter.MaxResults, filter.MaxResults)
	}

	query := strings.Replace(sql, "##QUERYFIELDS##", strings.Join(fields, ","), 1)
	rows, err := l.conn.Pgx.Query(ctx, query, sqlArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	schema := dbSchema{}
	var link *domain.Link
	links := make([]*domain.Link, 0)
	var credentialAttributes pgtype.JSONB
	for rows.Next() {
		link = &domain.Link{}
		if err := rows.Scan(
			&link.ID,
			&link.IssuerDID,
			&link.CreatedAt,
			&link.MaxIssuance,
			&link.ValidUntil,
			&link.SchemaID,
			&link.CredentialExpiration,
			&link.CredentialSignatureProof,
			&link.CredentialMTPProof, &credentialAttributes,
			&link.Active,
			&link.RefreshService,
			&link.DisplayMethod,
			&link.AuthorizationRequestMessage,
			&link.IssuedClaims,
			&schema.ID,
			&schema.IssuerID,
			&schema.URL,
			&schema.Type,
			&schema.Hash,
			&schema.Words,
			&schema.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if err := credentialAttributes.AssignTo(&link.CredentialSubject); err != nil {
			return nil, 0, fmt.Errorf("parsing credential attributes: %w", err)
		}

		link.Schema, err = toSchemaDomain(&schema)
		if err != nil {
			return nil, 0, fmt.Errorf("parsing link schema: %w", err)
		}

		links = append(links, link)
	}

	return links, count, nil
}

func (l link) Delete(ctx context.Context, id uuid.UUID, issuerDID w3c.DID) error {
	tx, err := l.conn.Pgx.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	const updateClaimsSql = `UPDATE claims SET link_id = NULL WHERE link_id = $1 AND identifier = $2`
	_, err = tx.Exec(ctx, updateClaimsSql, id.String(), issuerDID.String())
	if err != nil {
		return err
	}
	const sql = `DELETE FROM links WHERE id = $1 AND issuer_id =$2`
	cmd, err := tx.Exec(ctx, sql, id.String(), issuerDID.String())
	if err != nil {
		if err := tx.Rollback(ctx); err != nil {
			return err
		}
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrLinkDoesNotExist
	}
	return nil
}

func (l link) AddAuthorizationRequest(ctx context.Context, linkID uuid.UUID, issuerDID w3c.DID, authorizationRequest *protocol.AuthorizationRequestMessage) error {
	const sql = `UPDATE links SET authorization_request_message = $1 WHERE id = $2 AND issuer_id = $3`
	_, err := l.conn.Pgx.Exec(ctx, sql, authorizationRequest, linkID, issuerDID.String())
	return err
}
