package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
	"github.com/labstack/gommon/log"
	"github.com/lib/pq"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

const duplicateViolationErrorCode = "23505"

// ErrClaimDuplication claim duplication error
var (
	ErrClaimDuplication = errors.New("claim duplication error")
	// ErrClaimDoesNotExist claim does not exist
	ErrClaimDoesNotExist = errors.New("claim does not exist")
)

type claims struct{}

// NewClaims returns a new claim repository
func NewClaims() ports.ClaimsRepository {
	return &claims{}
}

func (c *claims) Save(ctx context.Context, conn db.Querier, claim *domain.Claim) (uuid.UUID, error) {
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
                    index_hash)
		VALUES ($1,  $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
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
			claim.HIndex).Scan(&id)
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
                    index_hash
		)
		VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
		ON CONFLICT ON CONSTRAINT claims_pkey 
		DO UPDATE SET 
			( expiration, updatable, version, rev_nonce, signature_proof, mtp_proof, data, identity_state, 
			other_identifier, schema_hash, schema_url, schema_type, issuer, credential_status, revoked, core_claim)
			= (EXCLUDED.expiration, EXCLUDED.updatable, EXCLUDED.version, EXCLUDED.rev_nonce, EXCLUDED.signature_proof,
		EXCLUDED.mtp_proof, EXCLUDED.data, EXCLUDED.identity_state, EXCLUDED.other_identifier, EXCLUDED.schema_hash, 
		EXCLUDED.schema_url, EXCLUDED.schema_type, EXCLUDED.issuer, EXCLUDED.credential_status, EXCLUDED.revoked, EXCLUDED.core_claim)
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
			claim.HIndex).Scan(&id)
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

	log.Errorf("error saving the claim: %v", err.Error())
	return uuid.Nil, fmt.Errorf("error saving the claim: %w", err)
}

func (c *claims) Revoke(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error {
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

func (c *claims) GetByRevocationNonce(ctx context.Context, conn db.Querier, identifier *core.DID, revocationNonce domain.RevNonceUint64) (*domain.Claim, error) {
	claim := domain.Claim{}
	row := conn.QueryRow(
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
				   core_claim
			FROM claims
			LEFT JOIN identity_states ON claims.identity_state = identity_states.state
			WHERE claims.identifier = $1
			  AND claims.rev_nonce = $2`, identifier.String(), revocationNonce)
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
		&claim.SignatureProof,
		&claim.MTPProof,
		&claim.Data,
		&claim.Identifier,
		&claim.IdentityState,
		&claim.CredentialStatus,
		&claim.CoreClaim)

	if err != nil && err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	if err != nil {
		return nil, fmt.Errorf("error getting the claim by nonce: %w", err)
	}

	return &claim, nil
}

func (c *claims) FindOneClaimBySchemaHash(ctx context.Context, conn db.Querier, subject *core.DID, schemaHash string) (*domain.Claim, error) {
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
				AND claims.revoked = false`, subject.String(), schemaHash)

	err := row.Scan(&claim.ID,
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

	if err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	return &claim, err
}

func (c *claims) RevokeNonce(ctx context.Context, conn db.Querier, revocation *domain.Revocation) error {
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

// GetByID get claim by id
func (c *claims) GetByID(ctx context.Context, conn db.Querier, identifier *core.DID, claimID uuid.UUID) (*domain.Claim, error) {
	claim := domain.Claim{}
	// language=PostgreSQL
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
       				core_claim
        FROM claims
        WHERE claims.identifier = $1 AND claims.id = $2`, identifier.String(),
		claimID).Scan(
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
		&claim.CoreClaim)

	if err != nil && err == pgx.ErrNoRows {
		return nil, ErrClaimDoesNotExist
	}

	return &claim, err
}
