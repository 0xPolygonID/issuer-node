package common

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"time"

	core "github.com/iden3/go-iden3-core"
	jsonSuite "github.com/iden3/go-schema-processor/json"
	"github.com/iden3/go-schema-processor/verifiable"
)

// TODO: Evaluate if this should be here
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
	credentialCtx := []string{verifiable.JSONLDSchemaW3CCredential2018, verifiable.JSONLDSchemaIden3Credential, schema.Metadata.Uris["jsonLdContext"].(string)}

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
	// TODO: this was changed because revocation nonce is cut in dart / js if number is too big
	_, err := rand.Read(buf[:4]) // was rand.Read(buf[:])

	return binary.LittleEndian.Uint64(buf[:]), err
}
