package domain

import "time"

// IdentityStatus represents type for state Status
type IdentityStatus string

const (
	// StatusCreated is default status for identity state
	StatusCreated IdentityStatus = "created"
	// StatusTransacted is a status for state that was published but result is not known
	StatusTransacted IdentityStatus = "transacted"
	// StatusConfirmed is a status for confirmed transaction
	StatusConfirmed IdentityStatus = "confirmed"
	// StatusFailed is a status for failed transaction
	StatusFailed IdentityStatus = "failed"
)

type IdentityState struct {
	StateID            int64          `json:"-"`
	Identifier         string         `json:"-"`
	State              *string        `json:"state"`
	RootOfRoots        *string        `json:"root_of_roots,omitempty"`
	ClaimsTreeRoot     *string        `json:"claims_tree_root,omitempty"`
	RevocationTreeRoot *string        `json:"revocation_tree_root,omitempty"`
	BlockTimestamp     *int           `json:"block_timestamp,omitempty"`
	BlockNumber        *int           `json:"block_number,omitempty"`
	TxID               *string        `json:"tx_id,omitempty"`
	PreviousState      *string        `json:"previous_state,omitempty"`
	Status             IdentityStatus `json:"status,omitempty"`
	ModifiedAt         time.Time      `json:"modified_at,omitempty"`
	CreatedAt          time.Time      `json:"created_at,omitempty"`
}
