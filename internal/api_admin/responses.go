package api_admin

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/iden3/go-schema-processor/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
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

func getLinkResponse(link *domain.Link) (*GetLink200JSONResponse, error) {
	attrs := make([]LinkRequestAttributesType, len(link.CredentialAttributes))
	for i, attr := range link.CredentialAttributes {
		attrs[i].Name = attr.Name
		switch attr.AttrType {
		case domain.TypeString:
			s, ok := attr.Value.(string)
			if !ok {
				return nil, fmt.Errorf("error converting <%v> to string", attr.Value)
			}
			attrs[i].Value = s
		case domain.TypeInteger:
			switch n := attr.Value.(type) {
			case int, int8, int16, int32, int64, float32, float64:
				attrs[i].Value = fmt.Sprintf("%d", n)
			case json.Number:
				attrs[i].Value = n.String()
			default:
				return nil, fmt.Errorf("error converting <%v> to string", attr.Value)
			}
		case domain.TypeBoolean:
			b, ok := attr.Value.(bool)
			if !ok {
				return nil, fmt.Errorf("error converting <%v> to boolean", attr.Value)
			}
			attrs[i].Value = fmt.Sprintf("%t", b)
		default:
			return nil, fmt.Errorf("unknown type <%s>. Error converting <%v>", attr.AttrType, attr.Value)
		}
	}
	return &GetLink200JSONResponse{
		Active:       link.Active,
		Attributes:   attrs,
		Expiration:   link.ValidUntil,
		Id:           link.ID,
		IssuedClaims: link.IssuedClaims, // TODO: Give a value when link redemption is implemented
		MaxIssuance:  link.MaxIssuance,
		SchemaType:   link.Schema.Type,
		SchemaUrl:    link.Schema.URL,
		Status:       LinkStatus(link.Status()),
	}, nil
}
