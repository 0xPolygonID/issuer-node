package domain

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-circuits/v2"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// CoreClaim is an alias for the core.Claim struct
type CoreClaim core.Claim

// Claim struct
type Claim struct {
	ID               uuid.UUID       `json:"-"`
	Identifier       *string         `json:"identifier"`
	Issuer           string          `json:"issuer"`
	SchemaHash       string          `json:"schema_hash"`
	SchemaURL        string          `json:"schema_url"`
	SchemaType       string          `json:"schema_type"`
	OtherIdentifier  string          `json:"other_identifier"`
	Expiration       int64           `json:"expiration"`
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

	MtProof               bool       `json:"mt_poof"`
	LinkID                *uuid.UUID `json:"-"`
	CreatedAt             time.Time  `json:"-"`
	SchemaTypeDescription *string    `json:"schema_type_description"`
}

// Credentials is the type of array of credential
type Credentials []*Claim

// FromClaimer TODO add description
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

// Scan TODO
func (c *CoreClaim) Scan(value interface{}) error {
	valueStr, ok := value.(string)
	if !ok {
		return fmt.Errorf("invalid value type, expected string")
	}
	var claim core.Claim
	err := json.Unmarshal([]byte(valueStr), &claim)
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

// ValidProof returns true if the claim has a valid proof
func (c *Claim) ValidProof() bool {
	if !c.MtProof || c.SignatureProof.Status != pgtype.Null { // this second condition is for credentials that has both types of proofs
		return true
	}

	return c.MTPProof.Status != pgtype.Null
}

// BuildTreeState returns circuits.TreeState structure
func BuildTreeState(state, claimsTreeRoot, revocationTreeRoot, rootOfRoots *string) (circuits.TreeState, error) {
	return circuits.TreeState{
		State:          common.StrMTHex(state),
		ClaimsRoot:     common.StrMTHex(claimsTreeRoot),
		RevocationRoot: common.StrMTHex(revocationTreeRoot),
		RootOfRoots:    common.StrMTHex(rootOfRoots),
	}, nil
}

// GetBJJSignatureProof2021 TBD
func (c *Claim) GetBJJSignatureProof2021() (*verifiable.BJJSignatureProof2021, error) {
	var sigProof verifiable.BJJSignatureProof2021
	err := c.SignatureProof.AssignTo(&sigProof)
	if err != nil {
		return &sigProof, err
	}
	return &sigProof, nil
}

// GetVerifiableCredential TBD
func (c *Claim) GetVerifiableCredential() (verifiable.W3CCredential, error) {
	var vc verifiable.W3CCredential
	err := c.Data.AssignTo(&vc)
	if err != nil {
		return vc, err
	}
	return vc, nil
}

// GetCircuitIncProof TBD
func (c *Claim) GetCircuitIncProof() (circuits.MTProof, error) {
	var proof verifiable.Iden3SparseMerkleTreeProof
	err := c.MTPProof.AssignTo(&proof)
	if err != nil {
		return circuits.MTProof{}, err
	}

	return circuits.MTProof{
		Proof: proof.MTP,
		TreeState: circuits.TreeState{
			State:          common.StrMTHex(proof.IssuerData.State.Value),
			ClaimsRoot:     common.StrMTHex(proof.IssuerData.State.ClaimsTreeRoot),
			RevocationRoot: common.StrMTHex(proof.IssuerData.State.RevocationTreeRoot),
			RootOfRoots:    common.StrMTHex(proof.IssuerData.State.RootOfRoots),
		},
	}, nil
}

// NewClaimModel creates domain.Claim with common fields filled from core.Claim
func NewClaimModel(jsonSchemaURL string, credentialType string, coreClaim core.Claim, did *w3c.DID) (*Claim, error) {
	hindex, err := coreClaim.HIndex()
	if err != nil {
		return nil, errors.Join(err)
	}

	schemaHash := coreClaim.GetSchemaHash()

	claimModel := Claim{
		SchemaHash: hex.EncodeToString(schemaHash[:]),
		SchemaURL:  jsonSchemaURL,
		SchemaType: credentialType,
		Updatable:  coreClaim.GetFlagUpdatable(),
		Version:    coreClaim.GetVersion(),
		RevNonce:   RevNonceUint64(coreClaim.GetRevocationNonce()),
		CoreClaim:  CoreClaim(coreClaim),
		HIndex:     hindex.String(),
	}

	if did != nil {
		var claimID, id core.ID
		id, err = core.IDFromDID(*did)
		if err != nil {
			return nil, err
		}

		claimID, err = coreClaim.GetID()
		if err != nil {
			return nil, nil
		}

		if claimID != id {
			return nil, errors.New("claim has ID, but it's not match with DID")
		}
		claimModel.OtherIdentifier = did.String()
	} else {
		_, err = coreClaim.GetID()
		if !errors.Is(err, core.ErrNoID) {
			return nil, errors.New("claim has ID, but no DID")
		}
	}

	if expDate, ok := coreClaim.GetExpirationDate(); ok {
		claimModel.Expiration = expDate.Unix()
	}

	return &claimModel, nil
}

// GetCredentialStatus returns CredentialStatus deserialized object
func (c *Claim) GetCredentialStatus() (*verifiable.CredentialStatus, error) {
	cStatus := new(verifiable.CredentialStatus)
	err := c.CredentialStatus.AssignTo(cStatus)
	if err != nil {
		return nil, err
	}
	return cStatus, nil
}
