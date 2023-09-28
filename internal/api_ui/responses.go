package api_ui

import (
	"net/http"
	"strings"
	"time"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	openapitypes "github.com/oapi-codegen/runtime/types"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	linkstate "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/schema"
)

// CustomQrContentResponse is a wrapper to return any content as an api response.
// Just implement the Visit* method to satisfy the expected interface for that type of response.
type CustomQrContentResponse struct {
	content []byte
}

// NewQrContentResponse returns a new CustomQrContentResponse.
func NewQrContentResponse(response []byte) *CustomQrContentResponse {
	return &CustomQrContentResponse{content: response}
}

// VisitGetQrFromStoreResponse satisfies the AuthQRCodeResponseObject
func (response CustomQrContentResponse) VisitGetQrFromStoreResponse(w http.ResponseWriter) error {
	return response.visit(w)
}

// VisitAuthQRCodeResponse satisfies the AuthQRCodeResponseObject
func (response CustomQrContentResponse) VisitAuthQRCodeResponse(w http.ResponseWriter) error {
	return response.visit(w)
}

// VisitGetCredentialQrCodeResponse satisfies the AuthQRCodeResponseObject
func (response CustomQrContentResponse) VisitGetCredentialQrCodeResponse(w http.ResponseWriter) error {
	return response.visit(w)
}

func (response CustomQrContentResponse) visit(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(response.content) // Returning the content without encoding it. It was previously encoded
	return err
}

func schemaResponse(s *domain.Schema) Schema {
	hash, _ := s.Hash.MarshalText()
	return Schema{
		Id:          s.ID.String(),
		Type:        s.Type,
		Url:         s.URL,
		BigInt:      s.Hash.BigInt().String(),
		Hash:        string(hash),
		CreatedAt:   s.CreatedAt,
		Version:     s.Version,
		Title:       s.Title,
		Description: s.Description,
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

	proofs := getProofs(credential)

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
		UserID:            credential.OtherIdentifier,
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

func getProofs(credential *domain.Claim) []string {
	proofs := make([]string, 0)
	if credential.SignatureProof.Bytes != nil {
		proofs = append(proofs, string(verifiable.BJJSignatureProofType))
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
	hash, _ := link.Schema.Hash.MarshalText()
	var date *openapitypes.Date
	if link.CredentialExpiration != nil {
		date = &openapitypes.Date{Time: *link.CredentialExpiration}
	}

	return Link{
		Id:                   link.ID,
		Active:               link.Active,
		CredentialSubject:    link.CredentialSubject,
		IssuedClaims:         link.IssuedClaims,
		MaxIssuance:          link.MaxIssuance,
		SchemaType:           link.Schema.Type,
		SchemaUrl:            link.Schema.URL,
		SchemaHash:           string(hash),
		Status:               LinkStatus(link.Status()),
		ProofTypes:           getLinkProofs(link),
		CreatedAt:            link.CreatedAt,
		Expiration:           link.ValidUntil,
		CredentialExpiration: date,
	}
}

func getLinkSimpleResponse(link domain.Link) LinkSimple {
	hash, _ := link.Schema.Hash.MarshalText()
	return LinkSimple{
		Id:         link.ID,
		SchemaType: link.Schema.Type,
		SchemaUrl:  link.Schema.URL,
		SchemaHash: string(hash),
		ProofTypes: getLinkProofs(link),
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

func getLinkQrCodeResponse(linkQrCode *linkstate.QRCodeMessage) *QrCodeResponse {
	if linkQrCode == nil {
		return nil
	}
	credentials := make([]QrCodeCredentialResponse, len(linkQrCode.Body.Credentials))
	for i, c := range linkQrCode.Body.Credentials {
		credentials[i] = QrCodeCredentialResponse{
			Id:          c.ID,
			Description: c.Description,
		}
	}

	return &QrCodeResponse{
		Id:   linkQrCode.ID,
		Thid: linkQrCode.ThreadID,
		Typ:  linkQrCode.Typ,
		Type: linkQrCode.Type,
		From: linkQrCode.From,
		To:   linkQrCode.To,
		Body: QrCodeBodyResponse{
			Url:         linkQrCode.Body.URL,
			Credentials: credentials,
		},
	}
}

func getRevocationStatusResponse(rs *verifiable.RevocationStatus) RevocationStatusResponse {
	response := RevocationStatusResponse{}
	response.Issuer.State = rs.Issuer.State
	response.Issuer.RevocationTreeRoot = rs.Issuer.RevocationTreeRoot
	response.Issuer.RootOfRoots = rs.Issuer.RootOfRoots
	response.Issuer.ClaimsTreeRoot = rs.Issuer.ClaimsTreeRoot
	response.Mtp.Existence = rs.MTP.Existence

	if rs.MTP.NodeAux != nil {
		key := rs.MTP.NodeAux.Key
		decodedKey := key.BigInt().String()
		value := rs.MTP.NodeAux.Value
		decodedValue := value.BigInt().String()
		response.Mtp.NodeAux = &struct {
			Key   *string `json:"key,omitempty"`
			Value *string `json:"value,omitempty"`
		}{
			Key:   &decodedKey,
			Value: &decodedValue,
		}
	}

	response.Mtp.Existence = rs.MTP.Existence
	siblings := make([]string, 0)
	for _, s := range rs.MTP.AllSiblings() {
		siblings = append(siblings, s.BigInt().String())
	}

	response.Mtp.Siblings = &siblings

	return response
}
