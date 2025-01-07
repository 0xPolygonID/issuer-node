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
	Scope               *pgtype.JSONB `json:"scopes"`
	CreatedAt           time.Time
}

// VerificationResponse holds the verification response data
type VerificationResponse struct {
	ID                  uuid.UUID
	VerificationQueryID uuid.UUID
	UserDID             string
	Response            *pgtype.JSONB `json:"response"`
	Pass                bool
	CreatedAt           time.Time
}
