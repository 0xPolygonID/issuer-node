package domain

import (
	"errors"
	"math/big"
	"strings"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/kms"
)

// AuthCoreClaimRevocationStatus - auth core claim revocation status
type AuthCoreClaimRevocationStatus struct {
	ID              string `json:"id"`
	Type            string `json:"type"`
	RevocationNonce uint   `json:"revocationNonce"`
}

// ErrInvalidIdentifier - invalid identifier error
var ErrInvalidIdentifier = errors.New("invalid identifier")

// Identity struct
type Identity struct {
	Identifier                    string
	State                         IdentityState
	DisplayName                   *string                       `json:"displayName"`
	Relay                         string                        `json:"relay"`
	Immutable                     bool                          `json:"immutable"`
	KeyType                       string                        `json:"keyType"`
	Address                       *string                       `json:"address"`
	Balance                       *big.Int                      `json:"balance"`
	AuthCoreClaimRevocationStatus AuthCoreClaimRevocationStatus `json:"authCoreClaimRevocationStatus"`
	AuthCredentialsIDs            []string                      `json:"authCredentialsIDs"`
}

// IdentityDisplayName struct
type IdentityDisplayName struct {
	Identifier  string  `json:"identifier"`
	DisplayName *string `json:"displayName"`
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

// GetResolverPrefix get resolver prefix
func (i *Identity) GetResolverPrefix() (string, error) {
	const itemsLen = 4
	items := strings.Split(i.Identifier, ":")
	if len(items) < itemsLen {
		return "", ErrInvalidIdentifier
	}
	chain := items[2]
	network := items[3]
	return chain + ":" + network, nil
}
