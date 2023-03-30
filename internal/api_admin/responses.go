package api_admin

import (
	"time"

	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	link_state "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/schema"
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

func schemaCollectionResponse(schemas []domain.Schema) []Schema {
	res := make([]Schema, len(schemas))
	for i, s := range schemas {
		res[i] = schemaResponse(&s)
	}
	return res
}

func credentialResponse(w3c *verifiable.W3CCredential, credential *domain.Claim) Credential {
	expired := false
	if w3c.Expiration != nil {
		if time.Now().UTC().After(w3c.Expiration.UTC()) {
			expired = true
		}
	}

	proofs := getProofs(w3c, credential)

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

func getProofs(w3c *verifiable.W3CCredential, credential *domain.Claim) []string {
	proofs := make([]string, 0)
	if sp := getSigProof(w3c); sp != nil {
		proofs = append(proofs, *sp)
	}

	if credential.MtProof {
		proofs = append(proofs, "MTP")
	}
	return proofs
}

func connectionsResponse(conns []*domain.Connection) (GetConnectionsResponse, error) {
	resp := make([]GetConnectionResponse, 0)
	var err error
	for _, conn := range conns {
		var w3creds []*verifiable.W3CCredential
		var connCreds domain.Credentials
		if conn.Credentials != nil {
			connCreds = *conn.Credentials
			w3creds, err = schema.FromClaimsModelToW3CCredential(connCreds)
			if err != nil {
				return nil, err
			}
		}
		resp = append(resp, connectionResponse(conn, w3creds, connCreds))
	}

	return resp, nil
}

func connectionResponse(conn *domain.Connection, w3cs []*verifiable.W3CCredential, credentials []*domain.Claim) GetConnectionResponse {
	credResp := make([]Credential, len(w3cs))
	if w3cs != nil {
		for i := range credentials {
			credResp[i] = credentialResponse(w3cs[i], credentials[i])
		}
	}

	return GetConnectionResponse{
		CreatedAt:   conn.CreatedAt,
		Id:          conn.ID.String(),
		UserID:      conn.UserDID.String(),
		IssuerID:    conn.IssuerDID.String(),
		Credentials: credResp,
	}
}

func getSigProof(w3c *verifiable.W3CCredential) *string {
	for i := range w3c.Proof {
		if string(w3c.Proof[i].ProofType()) == "BJJSignature2021" {
			return common.ToPointer("BJJSignature2021")
		}
	}
	return nil
}

func deleteConnectionResponse(deleteCredentials bool, revokeCredentials bool) string {
	msg := "Connection successfully deleted."
	if deleteCredentials {
		msg += " Credentials successfully deleted."
	}
	if revokeCredentials {
		msg += " Credentials successfully revoked."
	}

	return msg
}

func deleteConnection500Response(deleteCredentials bool, revokeCredentials bool) string {
	msg := "There was an error deleting the connection."
	if deleteCredentials {
		msg += " There was an error deleting the connection credentials."
	}
	if revokeCredentials {
		msg += " Credentials successfully revoked."
	}

	return msg
}

func getLinQrCodeResponse(linkQrCode *link_state.QRCodeMessage) *GetLinkQrCodeResponseType {
	if linkQrCode == nil {
		return nil
	}
	credentials := make([]GetLinkQrCodeCredentialsResponseType, len(linkQrCode.Body.Credentials))
	for i, c := range linkQrCode.Body.Credentials {
		credentials[i] = GetLinkQrCodeCredentialsResponseType{
			Id:          c.ID,
			Description: c.Description,
		}
	}

	return &GetLinkQrCodeResponseType{
		Id:   linkQrCode.ID,
		Thid: linkQrCode.ThreadID,
		Typ:  linkQrCode.Typ,
		Type: linkQrCode.Type,
		From: linkQrCode.From,
		To:   linkQrCode.To,
		Body: GetLinkQrCodeResponseBodyType{
			Url:         linkQrCode.Body.URL,
			Credentials: credentials,
		},
	}
}
