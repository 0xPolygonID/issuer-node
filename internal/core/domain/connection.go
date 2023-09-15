package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

// Connection struct
type Connection struct {
	ID          uuid.UUID
	IssuerDID   w3c.DID
	UserDID     w3c.DID
	IssuerDoc   json.RawMessage
	UserDoc     json.RawMessage
	CreatedAt   time.Time
	ModifiedAt  time.Time
	Credentials *Credentials
}
