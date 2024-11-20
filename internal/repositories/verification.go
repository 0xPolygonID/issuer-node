package repositories

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const foreignKeyViolationErrorCode = "23503"

var (
	// VerificationQueryNotFoundError is returned when a verification query is not found
	VerificationQueryNotFoundError = errors.New("verification query not found")
	// VerificationScopeNotFoundError is returned when a verification scope is not found
	VerificationScopeNotFoundError = errors.New("verification scope not found")
)

// VerificationRepository is a repository for verification queries
type VerificationRepository struct {
	conn db.Storage
}

// NewVerification creates a new VerificationRepository
func NewVerification(conn db.Storage) *VerificationRepository {
	return &VerificationRepository{conn: conn}
}

// Save stores a verification query in the database
func (r *VerificationRepository) Save(ctx context.Context, issuerID w3c.DID, query domain.VerificationQuery) (uuid.UUID, error) {
	sql := `INSERT INTO verification_queries (id, issuer_id, chain_id, skip_check_revocation)
			VALUES($1, $2, $3, $4) ON CONFLICT (id) DO
			UPDATE SET issuer_id=$2, chain_id=$3, skip_check_revocation=$4
			RETURNING id`

	var queryID uuid.UUID
	tx, err := r.conn.Pgx.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	err = tx.QueryRow(ctx, sql, query.ID, issuerID.String(), query.ChainID, query.SkipCheckRevocation).Scan(&queryID)
	if err != nil {
		errIn := tx.Rollback(ctx)
		if errIn != nil {
			return uuid.Nil, errIn
		}
		return uuid.Nil, err
	}

	for _, scope := range query.Scopes {
		sql = `INSERT INTO verification_scopes (id, verification_query_id, scope_id, circuit_id, context, allowed_issuers, credential_type, credential_subject)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (id) DO
					UPDATE SET circuit_id=$4, context=$5, allowed_issuers=$6, credential_type=$7, credential_subject=$8`
		_, err = tx.Exec(ctx, sql, scope.ID, queryID, scope.ScopeID, scope.CircuitID, scope.Context, scope.AllowedIssuers, scope.CredentialType, scope.CredentialSubject)
		if err != nil {
			errIn := tx.Rollback(ctx)
			if errIn != nil {
				return uuid.Nil, errIn
			}
			return uuid.Nil, err
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	return queryID, err
}

// Get returns a verification query by issuer and id
func (r *VerificationRepository) Get(ctx context.Context, issuerID w3c.DID, id uuid.UUID) (*domain.VerificationQuery, error) {
	sql := `SELECT id, issuer_id, chain_id, skip_check_revocation, created_at
			FROM verification_queries
			WHERE issuer_id = $1 and id = $2`

	var query domain.VerificationQuery
	err := r.conn.Pgx.QueryRow(ctx, sql, issuerID.String(), id).Scan(&query.ID, &query.IssuerDID, &query.ChainID, &query.SkipCheckRevocation, &query.CreatedAt)
	if err != nil {

		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, VerificationQueryNotFoundError
		}
		return nil, err
	}

	sql = `SELECT id, scope_id, circuit_id, context, allowed_issuers, credential_type, credential_subject, created_at
			FROM verification_scopes
			WHERE verification_query_id = $1`

	rows, err := r.conn.Pgx.Query(ctx, sql, query.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var scope domain.VerificationScope
		err = rows.Scan(&scope.ID, &scope.ScopeID, &scope.CircuitID, &scope.Context, &scope.AllowedIssuers, &scope.CredentialType, &scope.CredentialSubject, &scope.CreatedAt)
		if err != nil {
			return nil, err
		}
		query.Scopes = append(query.Scopes, scope)
	}

	return &query, nil
}

// GetAll returns all verification queries for a given issuer
func (r *VerificationRepository) GetAll(ctx context.Context, issuerID w3c.DID) ([]domain.VerificationQuery, error) {
	sql := `SELECT id, issuer_id, chain_id, skip_check_revocation, created_at
			FROM verification_queries
			WHERE issuer_id = $1`

	rows, err := r.conn.Pgx.Query(ctx, sql, issuerID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []domain.VerificationQuery
	for rows.Next() {
		var query domain.VerificationQuery
		err = rows.Scan(&query.ID, &query.IssuerDID, &query.ChainID, &query.SkipCheckRevocation, &query.CreatedAt)
		if err != nil {
			return nil, err
		}

		sql = `SELECT id, scope_id, circuit_id, context, allowed_issuers, credential_type, credential_subject, created_at
				FROM verification_scopes
				WHERE verification_query_id = $1`

		scopeRows, err := r.conn.Pgx.Query(ctx, sql, query.ID)
		if err != nil {
			return nil, err
		}
		defer scopeRows.Close()

		for scopeRows.Next() {
			var scope domain.VerificationScope
			err = scopeRows.Scan(&scope.ID, &scope.ScopeID, &scope.CircuitID, &scope.Context, &scope.AllowedIssuers, &scope.CredentialType, &scope.CredentialSubject, &scope.CreatedAt)
			if err != nil {
				return nil, err
			}
			query.Scopes = append(query.Scopes, scope)
		}
		queries = append(queries, query)
	}
	return queries, nil
}

// AddResponse stores a verification response in the database
func (r *VerificationRepository) AddResponse(ctx context.Context, scopeID uuid.UUID, response domain.VerificationResponse) (uuid.UUID, error) {
	sql := `INSERT INTO verification_responses (id, verification_scope_id, user_did, response, pass)
			VALUES($1, $2, $3, $4, $5) ON CONFLICT (id) DO
			UPDATE SET user_did=$3, response=$4, pass=$5
			RETURNING id`

	var responseID uuid.UUID
	err := r.conn.Pgx.QueryRow(ctx, sql, response.ID, scopeID, response.UserDID, response.Response, response.Pass).Scan(&responseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationErrorCode {
			return uuid.Nil, VerificationScopeNotFoundError
		}
		return uuid.Nil, err
	}
	return responseID, nil
}
