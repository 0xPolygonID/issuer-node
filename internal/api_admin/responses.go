package api_admin

import (
	"time"

	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
)

func schemaResponse(s *domain.Schema) Schema {
	hash, _ := s.Hash.MarshalText()
	return Schema{
		Id:        s.ID.String(),
		Type:      s.Type,
		Url:       s.URL,
		BigInt:    s.Hash.BigInt().String(),
		Hash:      string(hash),
		CreatedAt: s.CreatedAt,
	}
}

func credentialResponse(w3c *verifiable.W3CCredential, credential *domain.Claim) Credential {
	expired := false
	if w3c.Expiration != nil {
		if time.Now().UTC().After(w3c.Expiration.UTC()) {
			expired = true
		}
	}

	proofs := make([]string, len(w3c.Proof))
	for i := range w3c.Proof {
		proofs[i] = string(w3c.Proof[i].ProofType())
	}

	return Credential{
		Attributes: w3c.CredentialSubject,
		CreatedAt:  *w3c.IssuanceDate,
		Expired:    expired,
		ExpiresAt:  w3c.Expiration,
		Id:         credential.ID,
		ProofTypes: proofs,
		RevNonce:   uint64(credential.RevNonce),
		Revoked:    credential.Revoked,
		SchemaHash: credential.SchemaHash,
		SchemaType: credential.SchemaType,
	}
}

func connectionResponse(conn *domain.Connection, w3cs []*verifiable.W3CCredential, credentials []*domain.Claim) GetConnectionResponse {
	credResp := make([]Credential, len(w3cs))
	for i := range credentials {
		credResp[i] = credentialResponse(w3cs[i], credentials[i])
	}

	return GetConnectionResponse{
		Connection: Connection{
			CreatedAt: conn.CreatedAt,
			Id:        conn.ID.String(),
			UserID:    conn.UserDID.String(),
			IssuerID:  conn.IssuerDID.String(),
		},
		Credentials: credResp,
	}
}
