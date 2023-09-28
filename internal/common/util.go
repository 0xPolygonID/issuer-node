package common

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/verifiable"
)

// CredentialRequest is a model for credential creation
type CredentialRequest struct {
	CredentialSchema      string          `json:"credentialSchema"`
	LDContext             string          `json:"ldContext"`
	Type                  string          `json:"type"`
	CredentialSubject     json.RawMessage `json:"credentialSubject"`
	Expiration            int64           `json:"expiration,omitempty"`
	Version               uint32          `json:"version,omitempty"`
	RevNonce              *uint64         `json:"revNonce,omitempty"`
	SubjectPosition       string          `json:"subjectPosition,omitempty"`
	MerklizedRootPosition string          `json:"merklizedRootPosition,omitempty"`
}

// CreateCredential is util to create a Verifiable credential
func CreateCredential(issuer *w3c.DID, req CredentialRequest) (verifiable.W3CCredential, error) {
	var cred verifiable.W3CCredential

	credentialCtx := []string{
		verifiable.JSONLDSchemaW3CCredential2018,
		verifiable.JSONLDSchemaIden3Credential,
		req.LDContext,
	}

	credentialType := []string{verifiable.TypeW3CVerifiableCredential, req.Type}

	expirationTime := time.Unix(req.Expiration, 0)

	issuanceDate := time.Now()

	var credentialSubject map[string]interface{}
	err := json.Unmarshal(req.CredentialSubject, &credentialSubject)
	if err != nil {
		return verifiable.W3CCredential{}, err
	}

	idSubject, ok := credentialSubject["id"].(string)
	if ok {

		did, err := w3c.ParseDID(idSubject)
		if err != nil {
			return verifiable.W3CCredential{}, err
		}

		credentialSubject["id"] = did.String()
	}

	credentialSubject["type"] = req.Type

	cred = verifiable.W3CCredential{
		Context:           credentialCtx,
		Type:              credentialType,
		IssuanceDate:      &issuanceDate,
		CredentialSubject: credentialSubject,
		Issuer:            issuer.String(),
		CredentialSchema: verifiable.CredentialSchema{
			ID:   req.CredentialSchema,
			Type: verifiable.JSONSchema2023,
		},
	}
	if req.Expiration != 0 {
		cred.Expiration = &expirationTime
	}
	return cred, nil
}

// RandInt64 generate random uint64
func RandInt64() (uint64, error) {
	var buf [8]byte
	_, err := rand.Read(buf[:4])

	return binary.LittleEndian.Uint64(buf[:]), err
}

// CheckGenesisStateDID return nil if state is genesis state.
func CheckGenesisStateDID(did *w3c.DID, state *big.Int) error {
	stateHash, err := merkletree.NewHashFromBigInt(state)
	if err != nil {
		return err
	}

	id, err := core.IDFromDID(*did)
	if err != nil {
		return err
	}
	DIDFromState, err := core.NewDIDFromIdenState(id.Type(), stateHash.BigInt())
	if err != nil {
		return err
	}

	if DIDFromState.String() != did.String() {
		return fmt.Errorf("did from genesis state (%s) and provided (%s) don't match", DIDFromState.String(), did.String())
	}
	return nil
}

// StrMTHex string to merkle tree hash
func StrMTHex(s *string) *merkletree.Hash {
	if s == nil {
		return &merkletree.HashZero
	}

	h, err := merkletree.NewHashFromHex(*s)
	if err != nil {
		return &merkletree.HashZero
	}
	return h
}

// CompareMerkleTreeHash compare merkletree.Hash
func CompareMerkleTreeHash(_state1, _state2 *merkletree.Hash) bool {
	return bytes.Equal(_state1[:], _state2[:])
}

// ArrayStringToBigInt converts array of string to big int
func ArrayStringToBigInt(s []string) ([]*big.Int, error) {
	var o []*big.Int
	for i := 0; i < len(s); i++ {
		si, err := stringToBigInt(s[i])
		if err != nil {
			return o, nil
		}
		o = append(o, si)
	}
	return o, nil
}

// ArrayOfStringArraysToBigInt converts array of string arrays to big int
func ArrayOfStringArraysToBigInt(s [][]string) ([][]*big.Int, error) {
	var o [][]*big.Int
	for i := 0; i < len(s); i++ {
		si, err := ArrayStringToBigInt(s[i])
		if err != nil {
			return o, nil
		}
		o = append(o, si)
	}
	return o, nil
}

func stringToBigInt(s string) (*big.Int, error) {
	base := 10
	if bytes.HasPrefix([]byte(s), []byte("0x")) {
		base = 16
		s = strings.TrimPrefix(s, "0x")
	}
	n, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("cannot parse string to *big.Int: %s", s)
	}
	return n, nil
}
