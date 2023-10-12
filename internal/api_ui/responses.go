package api_ui

import (
	"fmt"
	"strings"
	"time"

	openapi_types "github.com/deepmap/oapi-codegen/pkg/types"
	"github.com/google/uuid"
	"github.com/iden3/go-schema-processor/verifiable"
	"github.com/iden3/iden3comm/packers"
	"github.com/iden3/iden3comm/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	link_state "github.com/polygonid/sh-id-platform/pkg/link"
	"github.com/polygonid/sh-id-platform/pkg/schema"
)

const (
	schemaParts = 2
)

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
	var date *openapi_types.Date
	if link.CredentialExpiration != nil {
		date = &openapi_types.Date{Time: *link.CredentialExpiration}
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

func getLinkQrCodeResponse(linkQrCode *link_state.QRCodeMessage) *QrCodeResponse {
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

func getCredentialQrCodeResponse(credential *domain.Claim, hostURL string) QrCodeResponse {
	id := uuid.NewString()
	return QrCodeResponse{
		Body: QrCodeBodyResponse{
			Credentials: []QrCodeCredentialResponse{
				{
					Description: getCredentialType(credential.SchemaType),
					Id:          credential.ID.String(),
				},
			},
			Url: getAgentEndpoint(hostURL),
		},
		From: credential.Issuer,
		Id:   id,
		Thid: id,
		To:   credential.OtherIdentifier,
		Typ:  string(packers.MediaTypePlainMessage),
		Type: string(protocol.CredentialOfferMessageType),
	}
}

func getCredentialType(credentialType string) string {
	parse := strings.Split(credentialType, "#")
	if len(parse) != schemaParts {
		return credentialType
	}
	return parse[1]
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

func getAgentEndpoint(hostURL string) string {
	return fmt.Sprintf("%s/v1/agent", strings.TrimSuffix(hostURL, "/"))
}

func requestsResponse(requets []*domain.Responce) (GetAllRequestsResponse, error) {
	resp := make([]GetRequest200Response, 0)
	for _, req := range requets {
		var res GetRequest200Response
		res.Id = req.Id
		res.IssuerId = req.Issuer_id
		res.SchemaID = req.SchemaID
		res.UserDID = req.UserDID
		res.Active = req.Active
		res.CredentialType = req.CredentialType
		res.RequestType = req.RequestType
		res.RoleType = req.RoleType
		res.ProofType=req.ProofType
		res.ProofId=req.ProofId
		res.Age=req.Age
		res.RequestStatus = req.RequestStatus
		res.VerifierStatus=req.VerifyStatus
		res.WalletStatus=req.WalletStatus
		res.Source = req.Source
		res.CreatedAt = req.CreatedAt
		res.ModifiedAt = req.ModifiedAt
		resp = append(resp, res)
	}

	return resp, nil
}

func notificationResponse(requets []*domain.NotificationReponse) (AllNotifications, error) {
	resp := make([]Notifications200Response, 0)
	for _, req := range requets {
		var res Notifications200Response
		res.Id = req.ID
		res.User_id=req.User_id
		res.Module=req.Module
		res.CreatedAt = req.CreatedAt
		res.NotificationType=req.NotificationType
		res.NotificationTitle=req.NotificationTitle
		res.NotificationMessage=req.NotificationMessage
		resp = append(resp, res)
	}
	return resp, nil
}
