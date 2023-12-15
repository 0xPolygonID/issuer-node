package domain

import (
	"math/big"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/issuer-node/internal/common"
	"github.com/polygonid/issuer-node/internal/kms"
)

// Identity struct
type Identity struct {
	Identifier string
	State      IdentityState
	Relay      string   `json:"relay"`
	Immutable  bool     `json:"immutable"`
	KeyType    string   `json:"keyType"`
	Address    *string  `json:"address"`
	Balance    *big.Int `json:"balance"`
}

// NewIdentityFromIdentifier default identity model from identity and root state
func NewIdentityFromIdentifier(did *w3c.DID, rootState string) (*Identity, error) {
	keyType := string(kms.KeyTypeBabyJubJub)
	isEthIdentity, address, err := common.CheckEthIdentityByDID(did)
	if err != nil {
		return nil, err
	}
	if isEthIdentity {
		keyType = string(kms.KeyTypeEthereum)
	}

	return &Identity{
		Identifier: did.String(),
		Relay:      "",
		Immutable:  false,
		KeyType:    keyType,
		Address:    &address,
		State: IdentityState{
			Identifier: did.String(),
			State:      &rootState,
		},
	}, nil
}
