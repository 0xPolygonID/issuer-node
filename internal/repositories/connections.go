package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
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

// SaveUserAuthentication creates a new entry in the user_authentications table
func (c *connections) SaveUserAuthentication(ctx context.Context, conn db.Querier, connID uuid.UUID, sessID uuid.UUID, mTime time.Time) error {
	sql := `INSERT INTO user_authentications (connection_id,session_id,created_at) VALUES($1, $2,$3)`
	_, err := conn.Exec(ctx, sql, connID.String(), sessID.String(), mTime)

	return err
}

func (c *connections) Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID w3c.DID) error {
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

func (c *connections) DeleteCredentials(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID w3c.DID) error {
	sql := `DELETE FROM claims USING connections WHERE claims.issuer = connections.issuer_id AND claims.other_identifier = connections.user_id AND connections.id = $1 AND connections.issuer_id = $2`
	_, err := conn.Exec(ctx, sql, id.String(), issuerID.String())

	return err
}

func (c *connections) GetByIDAndIssuerID(ctx context.Context, conn db.Querier, id uuid.UUID, issuerID w3c.DID) (*domain.Connection, error) {
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

func (c *connections) GetByUserSessionID(ctx context.Context, conn db.Querier, sessionID uuid.UUID) (*domain.Connection, error) {
	connection := dbConnection{}
	err := conn.QueryRow(ctx,
		`SELECT connections.id, connections.issuer_id,connections.user_id,connections.issuer_doc,connections.user_doc,connections.created_at,connections.modified_at 
				FROM connections 
				JOIN user_authentications ON connections.id = user_authentications.connection_id
				WHERE user_authentications.session_id = $1`, sessionID.String()).Scan(
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

func (c *connections) GetByUserID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, userDID w3c.DID) (*domain.Connection, error) {
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

func (c *connections) GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, filter *ports.NewGetAllConnectionsRequest) ([]*domain.Connection, uint, error) {
	fields := []string{"id", "issuer_id", "user_id", "issuer_doc", "user_doc", "created_at", "modified_at"}
	all := `SELECT ##QUERYFIELDS## 
	FROM connections 
	WHERE connections.issuer_id = $1`

	if filter.Query != "" {
		dids := tokenizeQuery(filter.Query)
		if len(dids) > 0 {
			all += " AND (" + buildPartialQueryDidLikes("connections.user_id", dids, "OR") + ")"
		}
	}

	var count uint
	if filter.Pagination != nil {
		countQuery := strings.Replace(all, "##QUERYFIELDS##", "COUNT(*)", 1)
		if err := conn.QueryRow(ctx, countQuery, issuerDID.String()).Scan(&count); err != nil {
			return nil, 0, err
		}

		all += fmt.Sprintf(" OFFSET %d LIMIT %d;", filter.Pagination.GetOffset(), filter.Pagination.GetLimit())
	}

	all = strings.Replace(all, "##QUERYFIELDS##", strings.Join(fields, ","), 1)
	rows, err := conn.Query(ctx, all, issuerDID.String())
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	domainConns := make([]*domain.Connection, 0)
	dbConn := dbConnection{}
	for rows.Next() {
		if err := rows.Scan(&dbConn.ID, &dbConn.IssuerDID, &dbConn.UserDID, &dbConn.IssuerDoc, &dbConn.UserDoc, &dbConn.CreatedAt, &dbConn.ModifiedAt); err != nil {
			return nil, 0, err
		}
		domainConn, err := toConnectionDomain(&dbConn)
		if err != nil {
			return nil, 0, err
		}
		domainConns = append(domainConns, domainConn)
	}

	if filter.Pagination == nil {
		count = uint(len(domainConns))
	}

	return domainConns, count, nil
}

func (c *connections) GetAllWithCredentialsByIssuerID(ctx context.Context, conn db.Querier, issuerDID w3c.DID, filter *ports.NewGetAllConnectionsRequest) ([]*domain.Connection, uint, error) {
	sqlQuery, countQuery, filters := buildGetAllWithCredentialsQueryAndFilters(issuerDID, filter)

	var count uint
	if filter.Pagination != nil {
		if err := conn.QueryRow(ctx, countQuery, filters...).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	rows, err := conn.Query(ctx, sqlQuery, filters...)
	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	conns, err := toConnectionsWithCredentials(rows)
	if filter.Pagination == nil {
		count = uint(len(conns))
	}

	return conns, count, err
}

func buildGetAllWithCredentialsQueryAndFilters(issuerDID w3c.DID, filter *ports.NewGetAllConnectionsRequest) (string, string, []interface{}) {
	fields := []string{
		"connections.id",
		"connections.issuer_id",
		"connections.user_id",
		"connections.issuer_doc",
		"connections.user_doc",
		"connections.created_at",
		"connections.modified_at",
		"claims.id",
		"claims.issuer",
		"claims.schema_hash",
		"claims.schema_url",
		"claims.schema_type",
		"claims.other_identifier",
		"claims.expiration",
		"claims.version",
		"claims.rev_nonce",
		"claims.updatable",
		"claims.signature_proof",
		"claims.mtp_proof",
		"claims.data",
		"claims.identifier",
		"claims.identity_state",
		"identity_states.status",
		"claims.credential_status",
		"claims.core_claim",
		"claims.mtp",
		"claims.created_at",
	}

	sqlQuery := `SELECT ##QUERYFIELDS##
	FROM connections 
	LEFT JOIN claims
	ON connections.issuer_id = claims.issuer AND connections.user_id = claims.other_identifier
	LEFT JOIN identity_states  ON claims.identity_state = identity_states.state`

	if filter.Query != "" {
		sqlQuery = fmt.Sprintf("%s LEFT JOIN schemas ON claims.schema_hash=schemas.hash AND claims.issuer=schemas.issuer_id ", sqlQuery)
	}

	sqlArgs := []interface{}{issuerDID.String()}

	sqlQuery = fmt.Sprintf("%s WHERE connections.issuer_id = $%d", sqlQuery, len(sqlArgs))
	if filter.Query != "" {
		terms := tokenizeQuery(filter.Query)
		if len(terms) > 0 {
			ftsConds := buildPartialQueryLikes("schemas.words", "OR", len(sqlArgs)+1, len(terms)) + " OR " + buildPartialQueryDidLikes("connections.user_id", terms, "OR")
			sqlQuery += fmt.Sprintf(" AND (%s) ", ftsConds)
			for _, term := range terms {
				sqlArgs = append(sqlArgs, term)
			}
		}
	}

	countQuery := strings.Replace(sqlQuery, "##QUERYFIELDS##", "COUNT(*)", 1)
	sqlQuery = strings.Replace(sqlQuery, "##QUERYFIELDS##", strings.Join(fields, ","), 1)

	sqlQuery += " ORDER BY claims.created_at DESC NULLS LAST, connections.created_at DESC"
	if filter.Pagination != nil {
		sqlQuery += fmt.Sprintf(" OFFSET %d LIMIT %d;", filter.Pagination.GetOffset(), filter.Pagination.GetLimit())
	}

	return sqlQuery, countQuery, sqlArgs
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
	issID, err := w3c.ParseDID(c.IssuerDID)
	if err != nil {
		return nil, fmt.Errorf("parsing issuer DID from connection: %w", err)
	}

	usrDID, err := w3c.ParseDID(c.UserDID)
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
