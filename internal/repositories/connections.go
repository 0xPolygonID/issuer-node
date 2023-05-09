package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrConnectionDoesNotExist connection does not exist
var ErrConnectionDoesNotExist = errors.New("connection does not exist")

type dbConnection struct {
	ID         uuid.UUID
	IssuerDID  string
	UserDID    string
	IssuerDoc  pgtype.JSONB
	UserDoc    pgtype.JSONB
	CreatedAt  time.Time
	ModifiedAt time.Time
}

type dbConnectionWithCredentials struct {
	dbConnection
	dbClaim
}

type connections struct{}

// NewConnections returns a new connections repository
func NewConnections() ports.ConnectionsRepository {
	return &connections{}
}

// Save stores in the database the given connection and updates the modified at in case already exists
func (c *connections) Save(ctx context.Context, conn db.Querier, connection *domain.Connection) (uuid.UUID, error) {
	var id uuid.UUID
	sql := `INSERT INTO connections (id,issuer_id, user_id, issuer_doc, user_doc,created_at,modified_at)
			VALUES($1, $2, $3, $4,$5,$6,$7) ON CONFLICT ON CONSTRAINT connections_issuer_user_key DO
			UPDATE SET issuer_id=$2, user_id=$3, issuer_doc=$4, user_doc=$5, modified_at = $7
			RETURNING id`
	err := conn.QueryRow(ctx, sql, connection.ID, connection.IssuerDID.String(), connection.UserDID.String(), connection.IssuerDoc, connection.UserDoc, connection.CreatedAt, connection.ModifiedAt).Scan(&id)

	return id, err
}

func (c *connections) Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID core.DID) error {
	sql := `DELETE FROM connections WHERE id = $1 AND issuer_id = $2`
	cmd, err := conn.Exec(ctx, sql, id.String(), issuerDID.String())
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrConnectionDoesNotExist
	}

	return nil
}

func (c *connections) DeleteCredentials(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID core.DID) error {
	sql := `DELETE FROM claims USING connections WHERE claims.issuer = connections.issuer_id AND claims.other_identifier = connections.user_id AND connections.id = $1 AND connections.issuer_id = $2`
	_, err := conn.Exec(ctx, sql, id.String(), issuerID.String())

	return err
}

func (c *connections) GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID core.DID) (*domain.Connection, error) {
	connection := dbConnection{}
	err := conn.QueryRow(ctx,
		`SELECT id, issuer_id,user_id,issuer_doc,user_doc,created_at,modified_at 
				FROM connections 
				WHERE connections.id = $1 AND connections.issuer_id = $2`, id.String(), issuerID.String()).Scan(
		&connection.ID,
		&connection.IssuerDID,
		&connection.UserDID,
		&connection.IssuerDoc,
		&connection.UserDoc,
		&connection.CreatedAt,
		&connection.ModifiedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return toConnectionDomain(&connection)
}

func (c *connections) GetByUserID(ctx context.Context, conn db.Querier, issuerDID core.DID, userDID core.DID) (*domain.Connection, error) {
	connection := dbConnection{}
	err := conn.QueryRow(ctx,
		`SELECT id, issuer_id,user_id,issuer_doc,user_doc,created_at,modified_at 
				FROM connections 
				WHERE   connections.issuer_id = $1 AND  connections.user_id = $2`, issuerDID.String(), userDID.String()).Scan(
		&connection.ID,
		&connection.IssuerDID,
		&connection.UserDID,
		&connection.IssuerDoc,
		&connection.UserDoc,
		&connection.CreatedAt,
		&connection.ModifiedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrConnectionDoesNotExist
		}
		return nil, err
	}

	return toConnectionDomain(&connection)
}

func (c *connections) GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerDID core.DID, query string) ([]*domain.Connection, error) {
	all := `SELECT id, issuer_id,user_id,issuer_doc,user_doc,created_at,modified_at 
FROM connections 
WHERE connections.issuer_id = $1`

	if query != "" {
		dids := tokenizeQuery(query)
		if len(dids) > 0 {
			all += " AND (" + buildPartialQueryDidLikes("connections.user_id", dids, "OR") + ")"
		}
	}

	rows, err := conn.Query(ctx, all, issuerDID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	domainConns := make([]*domain.Connection, 0)
	dbConn := dbConnection{}
	for rows.Next() {
		if err := rows.Scan(&dbConn.ID, &dbConn.IssuerDID, &dbConn.UserDID, &dbConn.IssuerDoc, &dbConn.UserDoc, &dbConn.CreatedAt, &dbConn.ModifiedAt); err != nil {
			return nil, err
		}
		domainConn, err := toConnectionDomain(&dbConn)
		if err != nil {
			return nil, err
		}
		domainConns = append(domainConns, domainConn)
	}

	return domainConns, nil
}

func (c *connections) GetAllWithCredentialsByIssuerID(ctx context.Context, conn db.Querier, issuerDID core.DID, query string) ([]*domain.Connection, error) {
	sqlQuery, filters := buildGetAllWithCredentialsQueryAndFilters(issuerDID, query)
	rows, err := conn.Query(ctx, sqlQuery, filters...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	return toConnectionsWithCredentials(rows)
}

func buildGetAllWithCredentialsQueryAndFilters(issuerDID core.DID, query string) (string, []interface{}) {
	sqlQuery := `SELECT connections.id, 
       			   connections.issuer_id,
       			   connections.user_id,
       			   connections.issuer_doc,
       			   connections.user_doc,
       			   connections.created_at,
       			   connections.modified_at,
				   claims.id,
				   claims.issuer,
				   claims.schema_hash,
				   claims.schema_url,
				   claims.schema_type,
				   claims.other_identifier,
				   claims.expiration,
				   claims.version,
				   claims.rev_nonce,
				   claims.updatable,
				   claims.signature_proof,
				   claims.mtp_proof,
				   claims.data,
				   claims.identifier,
				   claims.identity_state,
				   identity_states.status,
				   claims.credential_status,
				   claims.core_claim,
				   claims.mtp,
				   claims.created_at
	FROM connections 
	LEFT JOIN claims
	ON connections.issuer_id = claims.issuer AND connections.user_id = claims.other_identifier
	LEFT JOIN identity_states  ON claims.identity_state = identity_states.state`

	if query != "" {
		sqlQuery = fmt.Sprintf("%s LEFT JOIN schemas ON claims.schema_hash=schemas.hash AND claims.issuer=schemas.issuer_id ", sqlQuery)
	}

	filters := []interface{}{issuerDID.String()}

	sqlQuery = fmt.Sprintf("%s WHERE connections.issuer_id = $%d", sqlQuery, len(filters))
	if query != "" {
		filters = append(filters, fullTextSearchQuery(query, " | "))
		ftsConds := fmt.Sprintf("(schemas.ts_words @@ to_tsquery($%d))", len(filters))

		dids := tokenizeQuery(query)
		if len(dids) > 0 {
			ftsConds += " OR " + buildPartialQueryDidLikes("connections.user_id", dids, "OR")
		}
		sqlQuery += fmt.Sprintf(" AND (%s) ", ftsConds)
	}

	sqlQuery += " ORDER BY claims.created_at DESC NULLS LAST"

	return sqlQuery, filters
}

func toConnectionsWithCredentials(rows pgx.Rows) ([]*domain.Connection, error) {
	orderedConns := make([]uuid.UUID, 0)
	dbConns := make(map[uuid.UUID]*domain.Connection, 0)

	for rows.Next() {
		var dbConn dbConnectionWithCredentials
		err := rows.Scan(
			&dbConn.dbConnection.ID,
			&dbConn.IssuerDID,
			&dbConn.UserDID,
			&dbConn.IssuerDoc,
			&dbConn.UserDoc,
			&dbConn.dbConnection.CreatedAt,
			&dbConn.ModifiedAt,
			&dbConn.dbClaim.ID,
			&dbConn.Issuer,
			&dbConn.SchemaHash,
			&dbConn.SchemaURL,
			&dbConn.SchemaType,
			&dbConn.OtherIdentifier,
			&dbConn.Expiration,
			&dbConn.Version,
			&dbConn.RevNonce,
			&dbConn.Updatable,
			&dbConn.SignatureProof,
			&dbConn.MTPProof,
			&dbConn.Data,
			&dbConn.Identifier,
			&dbConn.IdentityState,
			&dbConn.Status,
			&dbConn.CredentialStatus,
			&dbConn.CoreClaim,
			&dbConn.MtProof,
			&dbConn.dbClaim.CreatedAt)
		if err != nil {
			return nil, err
		}

		if conn, ok := dbConns[dbConn.dbConnection.ID]; !ok {
			orderedConns = append(orderedConns, dbConn.dbConnection.ID)
			domainConn, err := toConnectionWithCredentialsDomain(dbConn)
			if err != nil {
				return nil, err
			}
			dbConns[dbConn.dbConnection.ID] = domainConn
		} else {
			*conn.Credentials = append(*conn.Credentials, toCredentialDomain(&dbConn.dbClaim))
			dbConns[dbConn.dbConnection.ID] = conn
		}
	}

	resp := make([]*domain.Connection, len(orderedConns))
	for i, conn := range orderedConns {
		resp[i] = dbConns[conn]
	}

	return resp, nil
}

func toConnectionWithCredentialsDomain(dbConn dbConnectionWithCredentials) (*domain.Connection, error) {
	domainConn, err := toConnectionDomain(&dbConn.dbConnection)
	if err != nil {
		return nil, err
	}

	creds := make(domain.Credentials, 0)
	cred := toCredentialDomain(&dbConn.dbClaim)
	if cred != nil {
		creds = append(creds, cred)
	}

	domainConn.Credentials = &creds

	return domainConn, err
}

func toConnectionDomain(c *dbConnection) (*domain.Connection, error) {
	issID, err := core.ParseDID(c.IssuerDID)
	if err != nil {
		return nil, fmt.Errorf("parsing issuer DID from connection: %w", err)
	}

	usrDID, err := core.ParseDID(c.UserDID)
	if err != nil {
		return nil, fmt.Errorf("parsing user DID from connection: %w", err)
	}

	conn := &domain.Connection{
		ID:         c.ID,
		IssuerDID:  *issID,
		UserDID:    *usrDID,
		CreatedAt:  c.CreatedAt,
		ModifiedAt: c.ModifiedAt,
	}

	if err := c.UserDoc.AssignTo(&conn.UserDoc); err != nil {
		return nil, fmt.Errorf("parsing user UserDoc from connection: %w", err)
	}

	if err := c.IssuerDoc.AssignTo(&conn.IssuerDoc); err != nil {
		return nil, fmt.Errorf("parsing user IssuerDoc from connection: %w", err)
	}

	return conn, nil
}
