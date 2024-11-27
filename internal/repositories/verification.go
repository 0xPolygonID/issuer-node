package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/jackc/pgconn"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const foreignKeyViolationErrorCode = "23503"

var (
	// VerificationQueryNotFoundError is returned when a verification query is not found
	VerificationQueryNotFoundError    = errors.New("verification query not found")
	VerificationResponseNotFoundError = errors.New("verification response not found")
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
	sql := `INSERT INTO verification_queries (id, issuer_id, chain_id, scope, skip_check_revocation)
			VALUES($1, $2, $3, $4, $5) ON CONFLICT (id) DO
			UPDATE SET issuer_id=$2, chain_id=$3, scope=$4, skip_check_revocation=$5
			RETURNING id`

	var queryID uuid.UUID
	if err := r.conn.Pgx.QueryRow(ctx, sql, query.ID, issuerID.String(), query.ChainID, query.Scope, query.SkipCheckRevocation).Scan(&queryID); err != nil {
		return uuid.Nil, err
	}
	return queryID, nil
}

// Get returns a verification query by issuer and id
func (r *VerificationRepository) Get(ctx context.Context, issuerID w3c.DID, id uuid.UUID) (*domain.VerificationQuery, error) {
	sql := `SELECT id, issuer_id, chain_id, scope, skip_check_revocation, created_at
			FROM verification_queries
			WHERE issuer_id = $1 and id = $2`

	var query domain.VerificationQuery
	err := r.conn.Pgx.QueryRow(ctx, sql, issuerID.String(), id).Scan(&query.ID, &query.IssuerDID, &query.ChainID, &query.Scope, &query.SkipCheckRevocation, &query.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, VerificationQueryNotFoundError
		}
		return nil, err
	}
	return &query, nil
}

// GetAll returns all verification queries for a given issuer
func (r *VerificationRepository) GetAll(ctx context.Context, issuerID w3c.DID) ([]domain.VerificationQuery, error) {
	sql := `SELECT id, issuer_id, chain_id, scope, skip_check_revocation, created_at
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
		err = rows.Scan(&query.ID, &query.IssuerDID, &query.ChainID, &query.Scope, &query.SkipCheckRevocation, &query.CreatedAt)
		if err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}
	return queries, nil
}

// AddResponse stores a verification response in the database
func (r *VerificationRepository) AddResponse(ctx context.Context, queryID uuid.UUID, response domain.VerificationResponse) (uuid.UUID, error) {
	sql := `INSERT INTO verification_responses (id, verification_query_id, user_did, response, pass)
			VALUES($1, $2, $3, $4, $5) ON CONFLICT (id) DO
			UPDATE SET user_did=$3, response=$4, pass=$5
			RETURNING id`

	var responseID uuid.UUID
	err := r.conn.Pgx.QueryRow(ctx, sql, response.ID, queryID, response.UserDID, response.Response, response.Pass).Scan(&responseID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolationErrorCode {
			return uuid.Nil, VerificationQueryNotFoundError
		}
		return uuid.Nil, err
	}
	return responseID, nil
}

// GetVerificationResponse returns a verification response by scopeID and userDID
func (r *VerificationRepository) GetVerificationResponse(ctx context.Context, queryID uuid.UUID, userDID string) (*domain.VerificationResponse, error) {
	sql := `SELECT id, verification_query_id, user_did, response, pass, created_at
            FROM verification_responses
            WHERE verification_query_id = $1 AND user_did = $2`

	var response domain.VerificationResponse
	err := r.conn.Pgx.QueryRow(ctx, sql, queryID, userDID).Scan(
		&response.ID,
		&response.VerificationQueryID,
		&response.UserDID,
		&response.Response,
		&response.Pass,
		&response.CreatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, VerificationResponseNotFoundError
		}
		return nil, err
	}

	return &response, nil
}
