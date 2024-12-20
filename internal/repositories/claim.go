package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/gommon/log"
	"github.com/lib/pq"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const duplicateViolationErrorCode = "23505"

var (
	// ErrClaimDuplication claim duplication error
	ErrClaimDuplication = errors.New("claim duplication error")
	// ErrClaimDoesNotExist claim does not exist
	ErrClaimDoesNotExist = errors.New("claim does not exist")
)

type claim struct{}

type dbClaim struct {
	ID               *uuid.UUID
	Identifier       sql.NullString
	Issuer           sql.NullString
	SchemaHash       sql.NullString
	SchemaURL        sql.NullString
	SchemaType       sql.NullString
	OtherIdentifier  sql.NullString
	Expiration       *int64
	Updatable        sql.NullBool
	Version          *int64
	RevNonce         *uint64
	Revoked          sql.NullBool
	Data             *pgtype.JSONB
	CoreClaim        *domain.CoreClaim
	MTPProof         *pgtype.JSONB
	SignatureProof   *pgtype.JSONB
	IdentityState    sql.NullString
	Status           sql.NullString
	CredentialStatus *pgtype.JSONB
	HIndex           sql.NullString
	CreatedAt        *time.Time

	MtProof sql.NullBool
}

// NewClaim returns a new claim repository
func NewClaim() ports.ClaimRepository {
	return &claim{}
}

// GetRevoked returns all the revoked claims from the given state
func (c *claim) GetRevoked(ctx context.Context, conn db.Querier, currentState string) ([]*domain.Claim, error) {
	query := `SELECT claims.id,
		issuer,
		schema_hash,
		schema_url,
		schema_type,
		other_identifier,
		expiration,
		updatable,
		claims.version,
		rev_nonce,
		signature_proof,
		mtp_proof,
		data,
		claims.identifier,
		identity_state,
		claims.metadata,
		credential_status,
		core_claim,
		revoked,
		mtp,
		claims.created_at
	FROM claims
	INNER JOIN revocation ON claims.rev_nonce = revocation.nonce AND claims.issuer = revocation.identifier
	WHERE claims.identity_state = $1`

	rows, err := conn.Query(ctx, query, currentState)
	if err != nil {
		return nil, err
	}

	claims, err := processClaims(rows)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (c *claim) Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error) {
	var err error
	id := claim.ID

	if claim.MTPProof.Status == pgtype.Undefined {
		claim.MTPProof.Status = pgtype.Null
	}
	if claim.Data.Status == pgtype.Undefined {
		claim.Data.Status = pgtype.Null
	}
	if claim.SignatureProof.Status == pgtype.Undefined {
		claim.SignatureProof.Status = pgtype.Null
	}
	if claim.CredentialStatus.Status == pgtype.Undefined {
		claim.CredentialStatus.Status = pgtype.Null
	}

	if id == uuid.Nil {
		s := `INSERT INTO claims (identifier,
                    other_identifier,
                    expiration,
                    updatable,
                    version,
					rev_nonce,
                    signature_proof,
                    issuer,
                    mtp_proof,
                    data,
                    identity_state,
                    schema_hash,
                    schema_url,
                    schema_type,
          			credential_status,
					revoked,
                    core_claim,
                    index_hash,
					mtp, 
					link_id,
                    created_at)
		VALUES ($1,  $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id`

		err = conn.QueryRow(ctx, s,
			claim.Identifier,
			claim.OtherIdentifier,
			claim.Expiration,
			claim.Updatable,
			claim.Version,
			claim.RevNonce,
			claim.SignatureProof,
			claim.Issuer,
			claim.MTPProof,
			claim.Data,
			claim.IdentityState,
			claim.SchemaHash,
			claim.SchemaURL,
			claim.SchemaType,
			claim.CredentialStatus,
			claim.Revoked,
			claim.CoreClaim,
			claim.HIndex,
			claim.MtProof,
			claim.LinkID,
			claim.CreatedAt).Scan(&id)
	} else {
		s := `INSERT INTO claims (
					id,
                    identifier,
                    other_identifier,
                    expiration,
                    updatable,
                    version,
					rev_nonce,
                    signature_proof,
                    issuer,
                    mtp_proof,
                    data,
                    identity_state,
					schema_hash,
                    schema_url,
                    schema_type,
                    credential_status,
                    revoked,
                    core_claim,
                    index_hash,
					mtp,
					link_id,
                    created_at
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
		)
		ON CONFLICT ON CONSTRAINT claims_pkey 
		DO UPDATE SET 
			( expiration, updatable, version, rev_nonce, signature_proof, mtp_proof, data, identity_state, 
			other_identifier, schema_hash, schema_url, schema_type, issuer, credential_status, revoked, core_claim, mtp, link_id, created_at)
			= (EXCLUDED.expiration, EXCLUDED.updatable, EXCLUDED.version, EXCLUDED.rev_nonce, EXCLUDED.signature_proof,
		EXCLUDED.mtp_proof, EXCLUDED.data, EXCLUDED.identity_state, EXCLUDED.other_identifier, EXCLUDED.schema_hash, 
		EXCLUDED.schema_url, EXCLUDED.schema_type, EXCLUDED.issuer, EXCLUDED.credential_status, EXCLUDED.revoked, EXCLUDED.core_claim, EXCLUDED.mtp, EXCLUDED.link_id, EXCLUDED.created_at)
			RETURNING id`
		err = conn.QueryRow(ctx, s,
			claim.ID,
			claim.Identifier,
			claim.OtherIdentifier,
			claim.Expiration,
			claim.Updatable,
			claim.Version,
			claim.RevNonce,
			claim.SignatureProof,
			claim.Issuer,
			claim.MTPProof,
			claim.Data,
			claim.IdentityState,
			claim.SchemaHash,
			claim.SchemaURL,
			claim.SchemaType,
			claim.CredentialStatus,
			claim.Revoked,
			claim.CoreClaim,
			claim.HIndex,
			claim.MtProof,
			claim.LinkID,
			claim.CreatedAt).Scan(&id)
	}

	if err == nil {
		return id, nil
	}

	pqErr, ok := err.(*pq.Error)
	if ok {
		if pqErr.Code == duplicateViolationErrorCode {
			return uuid.Nil, ErrClaimDuplication
		}
	}

	log.Errorf("error saving the claim: %v", "err", err.Error())
	return uuid.Nil, fmt.Errorf("error saving the claim: %w", err)
}

func (c *claim) Revoke(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error {
	_, err := conn.Exec(ctx, `INSERT INTO revocation (identifier, nonce, version, status, description) VALUES($1, $2, $3, $4, $5)`,
		revocation.Identifier,
		revocation.Nonce,
		revocation.Version,
		revocation.Status,
		revocation.Description)
	if err != nil {
		return fmt.Errorf("error revoking the claim: %w", err)
	}

	return nil
}

func (c *claim) Delete(ctx context.Context, conn db.Querier, id uuid.UUID) error {
	sql := `DELETE FROM claims WHERE id = $1`
	cmd, err := conn.Exec(ctx, sql, id.String())
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrClaimDoesNotExist
	}

	return nil
}

func (c *claim) GetByRevocationNonce(ctx context.Context, conn db.Querier, identifier *w3c.DID, revocationNonce domain.RevNonceUint64) ([]*domain.Claim, error) {
	rows, err := conn.Query(
		ctx,
		`SELECT id,
				   issuer,
				   schema_hash,
				   schema_type,
				   schema_url,
				   other_identifier,
				   expiration,
				   updatable,
				   version,
				   rev_nonce,
				   signature_proof,
				   mtp_proof,
				   data,
				   claims.identifier,
				   identity_state,
				   credential_status,
				   core_claim,
				   mtp
			FROM claims
			LEFT JOIN identity_states ON claims.identity_state = identity_states.state
			WHERE claims.identifier = $1
			  AND claims.rev_nonce = $2`, identifier.String(), revocationNonce)

	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	if err != nil && err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	if err != nil {
		return nil, fmt.Errorf("error getting claims by nonce: %w", err)
	}

	defer rows.Close()

	claims := make([]*domain.Claim, 0)
	for rows.Next() {
		claim := domain.Claim{}
		err = rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaURL,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.SignatureProof,
			&claim.MTPProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.CoreClaim,
			&claim.MtProof)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	if len(claims) == 0 {
		return nil, ErrClaimDoesNotExist
	}
	return claims, nil
}

// FindOneClaimBySchemaHash returns a claim by schema hash
// The claim must have MTP proof and not be revoked. This means the claim is published.
func (c *claim) FindOneClaimBySchemaHash(ctx context.Context, conn db.Querier, subject *w3c.DID, schemaHash string) (*domain.Claim, error) {
	var claim domain.Claim

	row := conn.QueryRow(ctx,
		`SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
		   revoked,
		   core_claim
		FROM claims
		WHERE claims.identifier=$1  
				AND ( claims.other_identifier = $1 or claims.other_identifier = '') 
				AND claims.schema_hash = $2 
				AND claims.revoked = false 
				AND claims.mtp_proof IS NOT NULL `, subject.String(), schemaHash)

	err := row.Scan(&claim.ID,
		&claim.Issuer,
		&claim.SchemaHash,
		&claim.SchemaType,
		&claim.SchemaURL,
		&claim.OtherIdentifier,
		&claim.Expiration,
		&claim.Updatable,
		&claim.Version,
		&claim.RevNonce,
		&claim.MTPProof,
		&claim.SignatureProof,
		&claim.Data,
		&claim.Identifier,
		&claim.IdentityState,
		&claim.CredentialStatus,
		&claim.Revoked,
		&claim.CoreClaim)

	if err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	return &claim, err
}

// FindClaimsBySchemaHash returns all claims by schema hash
// The claim must have MTP proof and not be revoked.
func (c *claim) FindClaimsBySchemaHash(ctx context.Context, conn db.Querier, subject *w3c.DID, schemaHash string) ([]*domain.Claim, error) {
	rows, err := conn.Query(ctx,
		`SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
       	   core_claim,
       	   revoked,
		   mtp,
		   claims.created_at
		FROM claims
		WHERE claims.identifier=$1  
				AND ( claims.other_identifier = $1 or claims.other_identifier = '') 
				AND claims.schema_hash = $2 
				AND claims.revoked = false 
				AND claims.mtp_proof IS NOT NULL `, subject.String(), schemaHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	credentials := make([]*domain.Claim, 0)
	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaURL,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.MTPProof,
			&claim.SignatureProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.CoreClaim,
			&claim.Revoked,
			&claim.MtProof,
			&claim.CreatedAt)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, &claim)
	}

	return credentials, nil
}

func (c *claim) RevokeNonce(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error {
	_, err := conn.Exec(ctx,
		`	INSERT INTO revocation (identifier, nonce, version, status, description) 
				VALUES($1, $2, $3, $4, $5)`,
		revocation.Identifier,
		revocation.Nonce,
		revocation.Version,
		revocation.Status,
		revocation.Description)
	return err
}

// GetByIdAndIssuer get claim by id
func (c *claim) GetByIdAndIssuer(ctx context.Context, conn db.Querier, identifier *w3c.DID, claimID uuid.UUID) (*domain.Claim, error) {
	claim := domain.Claim{}
	err := conn.QueryRow(ctx,
		`SELECT id,
       				issuer,
       				schema_hash,
       				schema_type,
       				schema_url,
       				other_identifier,
       				expiration,
       				updatable,
       				version,
        			rev_nonce,
       				signature_proof,
       				mtp_proof,
       				data,
       				claims.identifier,
        			identity_state,
       				credential_status,
       				core_claim,
					mtp,
					revoked,
					link_id
        FROM claims
        WHERE claims.identifier = $1 AND claims.id = $2`, identifier.String(), claimID).Scan(
		&claim.ID,
		&claim.Issuer,
		&claim.SchemaHash,
		&claim.SchemaType,
		&claim.SchemaURL,
		&claim.OtherIdentifier,
		&claim.Expiration,
		&claim.Updatable,
		&claim.Version,
		&claim.RevNonce,
		&claim.SignatureProof,
		&claim.MTPProof,
		&claim.Data,
		&claim.Identifier,
		&claim.IdentityState,
		&claim.CredentialStatus,
		&claim.CoreClaim,
		&claim.MtProof,
		&claim.Revoked,
		&claim.LinkID)

	if err != nil && err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	return &claim, err
}

// GetAllByIssuerID returns all the claims of the given issuer
func (c *claim) GetAllByIssuerID(ctx context.Context, conn db.Querier, issuerID w3c.DID, filter *ports.ClaimsFilter) (claims []*domain.Claim, count uint, err error) {
	query, countQuery, args := buildGetAllQueryAndFilters(issuerID, filter)

	// Let's count all results, only if we are paginating
	if filter.Page != nil {
		if err := conn.QueryRow(ctx, countQuery, args...).Scan(&count); err != nil {
			return nil, 0, err
		}
	}

	// Let's do the real query
	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, 0, ErrClaimDoesNotExist
		}

		return nil, 0, err
	}
	defer rows.Close()
	claims, err = processClaims(rows)

	if filter.Page == nil {
		count = uint(len(claims))
	}

	return claims, count, err
}

func (c *claim) GetNonRevokedByConnectionAndIssuerID(ctx context.Context, conn db.Querier, connID uuid.UUID, issuerID w3c.DID) ([]*domain.Claim, error) {
	query := `SELECT claims.id,
				   issuer,
				   schema_hash,
				   schema_url,
				   schema_type,
				   other_identifier,
				   expiration,
				   updatable,
				   claims.version,
				   rev_nonce,
				   signature_proof,
				   mtp_proof,
				   data,
				   claims.identifier,
				   identity_state,
				   identity_states.status,
				   credential_status,
				   core_claim,
				   revoked,
				   mtp,
				   claims.created_at
			FROM claims
			JOIN connections ON connections.issuer_id = claims.issuer AND connections.user_id = claims.other_identifier
			LEFT JOIN identity_states  ON claims.identity_state = identity_states.state
			WHERE connections.id = $1 AND claims.issuer = $2 AND  claims.revoked = false
			`

	rows, err := conn.Query(ctx, query, connID.String(), issuerID.String())

	if err != nil && err != pgx.ErrNoRows {
		return nil, err
	}

	defer rows.Close()

	return processClaims(rows)
}

func (c *claim) GetAllByState(ctx context.Context, conn db.Querier, did *w3c.DID, state *merkletree.Hash) (claims []domain.Claim, err error) {
	claims = make([]domain.Claim, 0)
	var rows pgx.Rows
	if state == nil {
		rows, err = conn.Query(ctx,
			`
		SELECT id,
			issuer,
			schema_hash,
			schema_url,
			schema_type,
			other_identifier,
			expiration,
			updatable,
			version,
			rev_nonce,
			signature_proof,
			mtp_proof,
			data,
			identifier,
			identity_state,
			NULL AS status,
			credential_status,
			core_claim 
		FROM claims
		WHERE issuer = $1 AND identity_state IS NULL AND identifier = issuer AND mtp = true
		`, did.String())
	} else {
		rows, err = conn.Query(ctx, `
		SELECT
			id,
			issuer,
			schema_hash,
			schema_url,
			schema_type,
			other_identifier,
			expiration,
			updatable,
			version,
			rev_nonce,
			signature_proof,
			mtp_proof,
			data,
			claims.identifier,
			identity_state,
			status,
			credential_status,
			core_claim 
		FROM claims
		  LEFT OUTER JOIN identity_states ON claims.identity_state = identity_states.state
		WHERE issuer = $1 AND ((identity_state IS NULL AND (mtp = true OR revoked = true) OR (identity_state = $2 AND mtp = true)))
		AND claims.identifier = issuer 
		`, did.String(), state.Hex())
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaURL,
			&claim.SchemaType,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.SignatureProof,
			&claim.MTPProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.Status,
			&claim.CredentialStatus,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, err
}

func (c *claim) GetAllByStateWithMTProof(ctx context.Context, conn db.Querier, did *w3c.DID, state *merkletree.Hash) (claims []domain.Claim, err error) {
	claims = make([]domain.Claim, 0)
	var rows pgx.Rows
	if state == nil {
		rows, err = conn.Query(ctx,
			`
		SELECT id,
			issuer,
			schema_hash,
			schema_url,
			schema_type,
			other_identifier,
			expiration,
			updatable,
			version,
			rev_nonce,
			signature_proof,
			mtp_proof,
			data,
			identifier,
			identity_state,
			NULL AS status,
			credential_status,
			core_claim 
		FROM claims
		WHERE issuer = $1 AND identity_state IS NULL AND identifier = issuer AND mtp = true
		`, did.String())
	} else {
		rows, err = conn.Query(ctx, `
		SELECT
			id,
			issuer,
			schema_hash,
			schema_url,
			schema_type,
			other_identifier,
			expiration,
			updatable,
			version,
			rev_nonce,
			signature_proof,
			mtp_proof,
			data,
			claims.identifier,
			identity_state,
			status,
			credential_status,
			core_claim 
		FROM claims
		  LEFT OUTER JOIN identity_states ON claims.identity_state = identity_states.state
		WHERE issuer = $1 AND identity_state = $2 AND claims.identifier = issuer AND mtp = true
		`, did.String(), state.Hex())
	}

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaURL,
			&claim.SchemaType,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.SignatureProof,
			&claim.MTPProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.Status,
			&claim.CredentialStatus,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	return claims, err
}

func (c *claim) UpdateState(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error) {
	query := "UPDATE claims SET identity_state = $1 WHERE id = $2 AND identifier = $3"
	res, err := conn.Exec(ctx, query, *claim.IdentityState, claim.ID, claim.Identifier)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

func processClaims(rows pgx.Rows) ([]*domain.Claim, error) {
	claims := make([]*domain.Claim, 0)

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaURL,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.SignatureProof,
			&claim.MTPProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.Status,
			&claim.CredentialStatus,
			&claim.CoreClaim,
			&claim.Revoked,
			&claim.MtProof,
			&claim.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, rows.Err()
}

func buildGetAllQueryAndFilters(issuerID w3c.DID, filter *ports.ClaimsFilter) (query string, countQuery string, filters []interface{}) {
	fields := []string{
		"claims.id",
		"issuer",
		"schema_hash",
		"schema_url",
		"schema_type",
		"other_identifier",
		"expiration",
		"updatable",
		"claims.version",
		"rev_nonce",
		"signature_proof",
		"mtp_proof",
		"data",
		"claims.identifier",
		"identity_state",
		"identity_states.status",
		"credential_status",
		"core_claim",
		"revoked",
		"mtp",
		"claims.created_at",
	}
	query = `SELECT ##QUERYFIELDS## FROM claims
			LEFT JOIN identity_states ON claims.identity_state = identity_states.state 
			`
	if filter.FTSQuery != "" {
		query = fmt.Sprintf("%s LEFT JOIN schemas ON claims.schema_hash=schemas.hash AND claims.issuer=schemas.issuer_id ", query)
	}

	filters = []interface{}{issuerID.String()}
	query = fmt.Sprintf("%s WHERE claims.identifier = $%d ", query, len(filters))

	query = fmt.Sprintf("%s AND claims.schema_type <> '%s' ", query, domain.AuthBJJCredentialSchemaType)

	if filter.Self != nil && *filter.Self {
		query = fmt.Sprintf("%s and other_identifier = '' ", query)
	}
	if filter.Subject != "" {
		filters = append(filters, filter.Subject)
		query = fmt.Sprintf("%s and other_identifier = $%d ", query, len(filters))
	}
	if filter.SchemaHash != "" {
		filters = append(filters, fmt.Sprintf("%s%%", filter.SchemaHash))
		query = fmt.Sprintf("%s and schema_hash like $%d", query, len(filters))
	}
	if filter.SchemaType != "" {
		filters = append(filters, fmt.Sprintf("%%%s%%", filter.SchemaType))
		query = fmt.Sprintf("%s and schema_type like $%d", query, len(filters))
	}
	if filter.Revoked != nil {
		filters = append(filters, *filter.Revoked)
		query = fmt.Sprintf("%s and claims.revoked = $%d", query, len(filters))
	}
	if filter.QueryField != "" {
		filters = append(filters, filter.QueryField, filter.QueryFieldValue)
		query = fmt.Sprintf("%s and data -> 'credentialSubject'  ->>$%d = $%d ", query, len(filters)-1, len(filters))
	}
	if filter.ExpiredOn != nil {
		t := *filter.ExpiredOn
		filters = append(filters, t.Unix())
		query = fmt.Sprintf("%s AND claims.expiration>0 AND claims.expiration<$%d", query, len(filters))
	}
	if len(filter.Proofs) > 0 {
		for _, proof := range filter.Proofs {
			switch proof {
			case verifiable.BJJSignatureProofType:
				query = fmt.Sprintf("%s AND claims.signature_proof IS NOT NULL", query)
			case verifiable.Iden3SparseMerkleTreeProofType:
				query = fmt.Sprintf("%s AND claims.mtp_proof IS NOT NULL", query)
			case domain.AnyProofType:
				query = fmt.Sprintf("%s AND ((claims.mtp = true AND claims.mtp_proof IS NOT NULL) OR claims.signature_proof IS NOT NULL)", query)
			}
		}
	}
	if filter.FTSQuery != "" {
		cond := "OR"
		if filter.FTSAndCond {
			cond = "AND"
		}
		terms := tokenizeQuery(filter.FTSQuery)
		ftsConds := buildPartialQueryLikes("schemas.words", cond, 1+len(filters), len(terms))
		for _, term := range terms {
			filters = append(filters, term)
		}
		if filter.Subject == "" {
			ftsConds += fmt.Sprintf(" %s %s", cond, buildPartialQueryDidLikes("claims.other_identifier", tokenizeQuery(filter.FTSQuery), cond))
		}
		query = fmt.Sprintf("%s AND (%s) ", query, ftsConds)
	}

	countQuery = strings.Replace(query, "##QUERYFIELDS##", "count(*)", 1)
	query = strings.Replace(query, "##QUERYFIELDS##", strings.Join(fields, ","), 1)

	_ = filter.OrderBy.Add(ports.CredentialCreatedAt, true)
	query += " ORDER BY " + filter.OrderBy.String()

	if filter.Page != nil {
		query += fmt.Sprintf(" OFFSET %d LIMIT %d;", (*filter.Page-1)*filter.MaxResults, filter.MaxResults)
	}

	return query, countQuery, filters
}

func (c *claim) UpdateClaimMTP(ctx context.Context, conn db.Querier, claim *domain.Claim) (int64, error) {
	query := "UPDATE claims SET mtp_proof = $1 WHERE id = $2 AND identifier = $3"
	res, err := conn.Exec(ctx, query, claim.MTPProof, claim.ID, claim.Identifier)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}

// GetAuthClaimsForPublishing of all claims for identity
func (c *claim) GetAuthClaimsForPublishing(ctx context.Context, conn db.Querier, identifier *w3c.DID, publishingState string, schemaHash string) ([]*domain.Claim, error) {
	var err error
	query := `SELECT claims.id,
		issuer,
       	schema_hash,
       	schema_type,
       	schema_url,
       	other_identifier,
       	expiration,
       	updatable,
       	claims.version,
		rev_nonce,
       	signature_proof,
       	mtp_proof,
       	data,
       	claims.identifier,
		identity_state,
		identity_states.status,
       	credential_status,
       	core_claim,
       	revoked,
		mtp,
		claims.created_at
	FROM claims
	LEFT JOIN identity_states  ON claims.identity_state = identity_states.state
	LEFT JOIN revocation  ON claims.rev_nonce = revocation.nonce AND claims.issuer = revocation.identifier
	WHERE claims.identifier = $1 
			AND state != $2
			AND claims.schema_hash = $3
			AND revocation.nonce IS NULL `

	rows, err := conn.Query(ctx, query, identifier.String(), publishingState, schemaHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims, err := processClaims(rows)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (c *claim) GetClaimsOfAConnection(ctx context.Context, conn db.Querier, identifier w3c.DID, userDID w3c.DID) ([]*domain.Claim, error) {
	query := `SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
		   revoked,
		   core_claim
		FROM claims
		WHERE claims.identifier = $1 
		AND claims.other_identifier = $2
	`
	rows, err := conn.Query(ctx, query, identifier.String(), userDID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]*domain.Claim, 0)

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaHash,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.MTPProof,
			&claim.SignatureProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.Revoked,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, nil
}

func (c *claim) GetClaimsIssuedForUser(ctx context.Context, conn db.Querier, identifier w3c.DID, userDID w3c.DID, linkID uuid.UUID) ([]*domain.Claim, error) {
	query := `SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
		   revoked,
		   core_claim
		FROM claims
		WHERE claims.identifier = $1 
		AND claims.other_identifier = $2
		AND claims.link_id = $3
	`
	rows, err := conn.Query(ctx, query, identifier.String(), userDID.String(), linkID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]*domain.Claim, 0)

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaHash,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.MTPProof,
			&claim.SignatureProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.Revoked,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}
	return claims, nil
}

func (c *claim) GetByStateIDWithMTPProof(ctx context.Context, conn db.Querier, did *w3c.DID, state string) ([]*domain.Claim, error) {
	query := `SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
		   revoked,
		   core_claim
		FROM claims
		WHERE claims.identifier = $1
		AND identity_state = $2
		AND mtp_proof IS NOT NULL
	`
	rows, err := conn.Query(ctx, query, did.String(), state)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]*domain.Claim, 0)

	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaHash,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.MTPProof,
			&claim.SignatureProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.Revoked,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}

	return claims, nil
}

// GetAuthCoreClaims returns all the core claims for the given identifier and schema hash
// The auth core claims returned may not be published onchain.
func (c *claim) GetAuthCoreClaims(ctx context.Context, conn db.Querier, identifier *w3c.DID, schemaHash string) ([]*domain.Claim, error) {
	rows, err := conn.Query(ctx,
		`SELECT claims.id,
		   issuer,
		   schema_hash,
		   schema_type,
		   schema_url,
		   other_identifier,
		   expiration,
		   updatable,
		   claims.version,
		   rev_nonce,
		   mtp_proof,
		   signature_proof,
		   data,
		   claims.identifier,
		   identity_state,
		   credential_status,
		   revoked,
		   core_claim
		FROM claims
		WHERE claims.identifier=$1  
				AND ( claims.other_identifier = $1 or claims.other_identifier = '') 
				AND claims.schema_hash = $2`, identifier.String(), schemaHash)
	if err != nil {
		return nil, err
	}

	claims := make([]*domain.Claim, 0)
	for rows.Next() {
		var claim domain.Claim
		err := rows.Scan(&claim.ID,
			&claim.Issuer,
			&claim.SchemaHash,
			&claim.SchemaType,
			&claim.SchemaURL,
			&claim.OtherIdentifier,
			&claim.Expiration,
			&claim.Updatable,
			&claim.Version,
			&claim.RevNonce,
			&claim.MTPProof,
			&claim.SignatureProof,
			&claim.Data,
			&claim.Identifier,
			&claim.IdentityState,
			&claim.CredentialStatus,
			&claim.Revoked,
			&claim.CoreClaim)
		if err != nil {
			return nil, err
		}
		claims = append(claims, &claim)
	}
	return claims, nil
}
