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

	core "github.com/iden3/go-iden3-core"
	"github.com/iden3/go-merkletree-sql/v2"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
)

// CredentialRequest is a model for credential creation
type CredentialRequest struct {
	CredentialSchema      string          `json:"credentialSchema"`
	Type                  string          `json:"type"`
	CredentialSubject     json.RawMessage `json:"credentialSubject"`
	Expiration            int64           `json:"expiration,omitempty"`
	Version               uint32          `json:"version,omitempty"`
	RevNonce              *uint64         `json:"revNonce,omitempty"`
	SubjectPosition       string          `json:"subjectPosition,omitempty"`
	MerklizedRootPosition string          `json:"merklizedRootPosition,omitempty"`
}

// CreateCredential is util to create a Verifiable credential
func CreateCredential(issuer *core.DID, req CredentialRequest, schema jsonSuite.Schema) (verifiable.W3CCredential, error) {
	var cred verifiable.W3CCredential

	jsonLdContext, ok := schema.Metadata.Uris["jsonLdContext"].(string)
	if !ok {
		return verifiable.W3CCredential{}, fmt.Errorf("invalid jsonLdContext, expected string")
	}
	credentialCtx := []string{verifiable.JSONLDSchemaW3CCredential2018, verifiable.JSONLDSchemaIden3Credential, jsonLdContext}

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

		did, err := core.ParseDID(idSubject)
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
			Type: verifiable.JSONSchemaValidator2018,
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
func CheckGenesisStateDID(did *core.DID, state *big.Int) error {
	stateHash, err := merkletree.NewHashFromBigInt(state)
	if err != nil {
		return err
	}

	t, err := core.BuildDIDType(did.Method, did.Blockchain, did.NetworkID)
	if err != nil {
		return err
	}
	DIDFromState, err := core.DIDGenesisFromIdenState(t, stateHash.BigInt())
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

func stringToBigInt(s string) (*big.Int, error) {
	base := 10
	if bytes.HasPrefix([]byte(s), []byte("0x")) {
		base = 16
		s = strings.TrimPrefix(s, "0x")
	}
	n, ok := new(big.Int).SetString(s, base)
	if !ok {
		return nil, fmt.Errorf("can not parse string to *big.Int: %s", s)
	}
	return n, nil
}
