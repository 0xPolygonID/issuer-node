package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core"
)

// Connection struct
type Connection struct {
	ID         uuid.UUID
	IssuerDID  *core.DID
	UserDID    *core.DID
	IssuerDoc  json.RawMessage
	UserDoc    json.RawMessage
	CreatedAt  time.Time
	ModifiedAt time.Time
}
