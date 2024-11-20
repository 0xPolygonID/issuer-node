package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgtype"
)

// VerificationQuery holds the verification data
type VerificationQuery struct {
	ID                  uuid.UUID
	IssuerDID           string
	ChainID             int
	SkipCheckRevocation bool
	Scopes              []VerificationScope
	CreatedAt           time.Time
}

// VerificationScope holds the verification scope data
type VerificationScope struct {
	ID                uuid.UUID
	ScopeID           int
	CircuitID         string
	Context           string
	AllowedIssuers    []string
	CredentialType    string
	CredentialSubject pgtype.JSONB `json:"credential_subject"`
	CreatedAt         time.Time
}

// VerificationResponse holds the verification response data
type VerificationResponse struct {
	ID                  uuid.UUID
	VerificationScopeID uuid.UUID
	UserDID             string
	Response            pgtype.JSONB `json:"response"`
	Pass                bool
	CreatedAt           time.Time
}
