package domain

import (
	"database/sql/driver"
	"strconv"
)

type RevNonceUint64 uint64

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

type Revocation struct {
	ID          int64          `json:"-"`
	Identifier  string         `json:"identifier"`
	Nonce       RevNonceUint64 `json:"nonce"`
	Version     uint32         `json:"version"`
	Status      RevStatus      `json:"status"`
	Description string         `json:"description"`
}

//func RevocationStatusToTreeState(status verifiable.RevocationStatus) circuits.TreeState {
//	return circuits.TreeState{
//		State:          common.StrMTHex(status.Issuer.State),
//		ClaimsRoot:     common.StrMTHex(status.Issuer.ClaimsTreeRoot),
//		RevocationRoot: common.StrMTHex(status.Issuer.RevocationTreeRoot),
//		RootOfRoots:    common.StrMTHex(status.Issuer.RootOfRoots),
//	}
//}
