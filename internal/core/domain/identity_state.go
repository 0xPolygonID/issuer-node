package domain

import (
	"time"

	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
)

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

// IdentityState struct
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

// PublishedState defines the domain object of publish state on chain
type PublishedState struct {
	TxID               *string
	ClaimsTreeRoot     *string
	State              *string
	RevocationTreeRoot *string
	RootOfRoots        *string
}

// ToTreeState returns circuits.TreeState structure
func (i *IdentityState) ToTreeState() (circuits.TreeState, error) {
	return BuildTreeState(i.State, i.ClaimsTreeRoot, i.RevocationTreeRoot, i.RootOfRoots)
}

// TreeState returns
func (i *IdentityState) TreeState() circuits.TreeState {
	return circuits.TreeState{
		State:          common.StrMTHex(i.State),
		ClaimsRoot:     common.StrMTHex(i.ClaimsTreeRoot),
		RevocationRoot: common.StrMTHex(i.RevocationTreeRoot),
		RootOfRoots:    common.StrMTHex(i.RootOfRoots),
	}
}

// ContainsID check if states contains id
func ContainsID(states []IdentityState, id *w3c.DID) bool {
	for i := range states {
		if states[i].Identifier == id.String() {
			return true
		}
	}
	return false
}
