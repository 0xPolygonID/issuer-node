package domain

import (
	"database/sql/driver"
	"strconv"

	"github.com/iden3/go-circuits/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
)

// RevNonceUint64 uint64 alias
type RevNonceUint64 uint64

// Value TODO
func (r RevNonceUint64) Value() (driver.Value, error) {
	return strconv.FormatUint(uint64(r), 10), nil
}

// RevStatus status of revocation nonce
type RevStatus int

const (
	// RevPending revocation nonce added but not published on chain
	RevPending RevStatus = 0
	// RevPublished revocation nonce published on chain
	RevPublished RevStatus = 1
)

// Revocation struct
type Revocation struct {
	ID          int64          `json:"-"`
	Identifier  string         `json:"identifier"`
	Nonce       RevNonceUint64 `json:"nonce"`
	Version     uint32         `json:"version"`
	Status      RevStatus      `json:"status"`
	Description string         `json:"description"`
}

// RevocationStatusToTreeState TBD
func RevocationStatusToTreeState(status verifiable.RevocationStatus) circuits.TreeState {
	return circuits.TreeState{
		State:          common.StrMTHex(status.Issuer.State),
		ClaimsRoot:     common.StrMTHex(status.Issuer.ClaimsTreeRoot),
		RevocationRoot: common.StrMTHex(status.Issuer.RevocationTreeRoot),
		RootOfRoots:    common.StrMTHex(status.Issuer.RootOfRoots),
	}
}
