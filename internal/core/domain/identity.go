package domain

import core "github.com/iden3/go-iden3-core"

type Identity struct {
	Identifier string
	Relay      string
	Immutable  bool
	State      IdentityState
}

// NewIdentityFromIdentifier default identity model from identity and root state
func NewIdentityFromIdentifier(id *core.DID, rootState string) *Identity {
	return &Identity{
		Identifier: id.String(),
		Relay:      "",
		Immutable:  false,
		State: IdentityState{
			Identifier: id.String(),
			State:      &rootState,
		},
	}
}
