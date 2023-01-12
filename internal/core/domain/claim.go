package domain

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
	"github.com/jackc/pgtype"
)

// CoreClaim is an alias for the core.Claim struct
type CoreClaim core.Claim

type Claim struct {
	ID              uuid.UUID `json:"-"`
	Identifier      *string   `json:"identifier"`
	Issuer          string    `json:"issuer"`
	SchemaHash      string    `json:"schema_hash"`
	SchemaURL       string    `json:"schema_url"`
	SchemaType      string    `json:"schema_type"`
	OtherIdentifier string    `json:"other_identifier"`
	Expiration      int64     `json:"expiration"`
	// TODO(illia-korotia): delete from db but left in struct.
	Updatable        bool            `json:"updatable"`
	Version          uint32          `json:"version"`
	RevNonce         RevNonceUint64  `json:"rev_nonce"`
	Revoked          bool            `json:"revoked"`
	Data             pgtype.JSONB    `json:"data"`
	CoreClaim        CoreClaim       `json:"core_claim"`
	MTPProof         pgtype.JSONB    `json:"mtp_proof"`
	SignatureProof   pgtype.JSONB    `json:"signature_proof"`
	IdentityState    *string         `json:"-"`
	Status           *IdentityStatus `json:"status"`
	CredentialStatus pgtype.JSONB    `json:"credential_status"`
	HIndex           string          `json:"-"`
}

func FromClaimer(claim *core.Claim, schemaURL, schemaType string) (*Claim, error) {
	otherIdentifier := ""
	id, err := claim.GetID()
	switch err {
	case core.ErrNoID:
	case nil:
		otherDID, errIn := core.ParseDIDFromID(id)
		if errIn != nil {
			return nil, fmt.Errorf("ID is not DID: %w", err)
		}
		otherIdentifier = otherDID.String()

	default:
		return nil, fmt.Errorf("can't get ID: %w", err)
	}

	var expiration int64
	if expirationDate, ok := claim.GetExpirationDate(); ok {
		expiration = expirationDate.Unix()
	}

	hindex, err := claim.HIndex()
	if err != nil {
		return nil, err
	}

	sb := claim.GetSchemaHash()
	schemaHash := hex.EncodeToString(sb[:])
	res := Claim{
		SchemaHash:      schemaHash,
		SchemaURL:       schemaURL,
		SchemaType:      schemaType,
		OtherIdentifier: otherIdentifier,
		Expiration:      expiration,
		Updatable:       claim.GetFlagUpdatable(),
		Version:         claim.GetVersion(),
		RevNonce:        RevNonceUint64(claim.GetRevocationNonce()),
		CoreClaim:       CoreClaim(*claim),
		HIndex:          hindex.String(),
	}

	return &res, nil
}

// Value implementation of valuer interface to convert CoreClaim value for storing in Postgres
func (c CoreClaim) Value() (driver.Value, error) {
	cc := core.Claim(c)
	jsonStr, err := json.Marshal(cc)
	return string(jsonStr), err
}

func (c *CoreClaim) Scan(value interface{}) error {
	s := value.(string)
	var claim core.Claim
	err := json.Unmarshal([]byte(s), &claim)
	if err != nil {
		return err
	}
	*c = CoreClaim(claim)
	return nil
}

// Get returns the value of the core claim
func (c *CoreClaim) Get() *core.Claim {
	return (*core.Claim)(c)
}
