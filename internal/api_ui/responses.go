package api_ui

import (
	"strings"
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
		CredentialSubject: w3c.CredentialSubject,
		CreatedAt:         *w3c.IssuanceDate,
		Expired:           expired,
		ExpiresAt:         w3c.Expiration,
		Id:                credential.ID,
		ProofTypes:        proofs,
		RevNonce:          uint64(credential.RevNonce),
		Revoked:           credential.Revoked,
		SchemaHash:        credential.SchemaHash,
		SchemaType:        shortType(credential.SchemaType),
		SchemaUrl:         credential.SchemaURL,
	}
}

func shortType(id string) string {
	parts := strings.Split(id, "#")
	l := len(parts)
	if l == 0 {
		return ""
	}
	return parts[l-1]
}

func getProofs(w3c *verifiable.W3CCredential, credential *domain.Claim) []string {
	proofs := make([]string, 0)
	if sp := getSigProof(w3c); sp != nil {
		proofs = append(proofs, *sp)
	}

	if credential.MtProof {
		proofs = append(proofs, string(verifiable.SparseMerkleTreeProof))
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

func stateTransactionsResponse(states []domain.IdentityState) StateTransactionsResponse {
	stateTransactions := make([]StateTransaction, len(states))
	for i := range states {
		stateTransactions[i] = toStateTransaction(states[i])
	}
	return stateTransactions
}

func toStateTransaction(state domain.IdentityState) StateTransaction {
	var stateTran, txID string
	if state.State != nil {
		stateTran = *state.State
	}
	if state.TxID != nil {
		txID = *state.TxID
	}
	return StateTransaction{
		Id:          state.StateID,
		PublishDate: state.ModifiedAt,
		State:       stateTran,
		Status:      getTransactionStatus(state.Status),
		TxID:        txID,
	}
}

func getTransactionStatus(status domain.IdentityStatus) StateTransactionStatus {
	switch status {
	case domain.StatusCreated:
		return "pending"
	case domain.StatusTransacted:
		return "transacted"
	case domain.StatusConfirmed:
		return "published"
	default:
		return "failed"
	}
}

func getSigProof(w3c *verifiable.W3CCredential) *string {
	for i := range w3c.Proof {
		if string(w3c.Proof[i].ProofType()) == string(verifiable.BJJSignatureProofType) {
			return common.ToPointer(string(verifiable.BJJSignatureProofType))
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

func getLinkResponse(link domain.Link) Link {
	return Link{
		Active:            link.Active,
		CredentialSubject: link.CredentialSubject,
		Expiration:        link.ValidUntil,
		Id:                link.ID,
		IssuedClaims:      link.IssuedClaims,
		MaxIssuance:       link.MaxIssuance,
		SchemaType:        link.Schema.Type,
		SchemaUrl:         link.Schema.URL,
		Status:            LinkStatus(link.Status()),
		ProofTypes:        getLinkProofs(link),
	}
}

func getLinkProofs(link domain.Link) []string {
	proofs := make([]string, 0)
	if link.CredentialMTPProof {
		proofs = append(proofs, string(verifiable.SparseMerkleTreeProof))
	}

	if link.CredentialSignatureProof {
		proofs = append(proofs, string(verifiable.BJJSignatureProofType))
	}

	return proofs
}

func getLinkResponses(links []domain.Link) []Link {
	res := make([]Link, len(links))
	for i, link := range links {
		res[i] = getLinkResponse(link)
	}
	return res
}

func getLinkQrCodeResponse(linkQrCode *link_state.QRCodeMessage) *GetLinkQrCodeResponseType {
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
