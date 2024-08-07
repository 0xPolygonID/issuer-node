package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/timeapi"
	"github.com/polygonid/sh-id-platform/pkg/pagination"
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

func (response CustomQrContentResponse) visit(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(response.content) // Returning the content without encoding it. It was previously encoded
	return err
}

func getLinkResponses(links []domain.Link) []Link {
	res := make([]Link, len(links))
	for i, link := range links {
		res[i] = getLinkResponse(link)
	}
	return res
}

func getLinkResponse(link domain.Link) Link {
	hash, _ := link.Schema.Hash.MarshalText()
	var credentialExpiration *timeapi.Time
	if link.CredentialExpiration != nil {
		t := timeapi.Time(*link.CredentialExpiration)
		credentialExpiration = common.ToPointer(t.UTCZeroHHMMSS())
	}

	var validUntil *TimeUTC
	if link.ValidUntil != nil {
		validUntil = common.ToPointer(TimeUTC(*link.ValidUntil))
	}

	var refreshService *RefreshService
	if link.RefreshService != nil {
		refreshService = &RefreshService{
			Id:   link.RefreshService.ID,
			Type: RefreshServiceType(link.RefreshService.Type),
		}
	}

	var displayMethod *DisplayMethod
	if link.DisplayMethod != nil {
		displayMethod = &DisplayMethod{
			Id:   link.DisplayMethod.ID,
			Type: DisplayMethodType(link.DisplayMethod.Type),
		}
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
		CreatedAt:            TimeUTC(link.CreatedAt),
		Expiration:           validUntil,
		CredentialExpiration: credentialExpiration,
		RefreshService:       refreshService,
		DisplayMethod:        displayMethod,
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

func credentialResponse(w3c *verifiable.W3CCredential, credential *domain.Claim) Credential {
	var expiresAt *TimeUTC
	expired := false
	if w3c.Expiration != nil {
		if time.Now().UTC().After(w3c.Expiration.UTC()) {
			expired = true
		}
		expiresAt = common.ToPointer(TimeUTC(*w3c.Expiration))
	}

	proofs := getProofs(credential)

	var refreshService *RefreshService
	if w3c.RefreshService != nil {
		refreshService = &RefreshService{
			Id:   w3c.RefreshService.ID,
			Type: RefreshServiceType(w3c.RefreshService.Type),
		}
	}

	var displayService *DisplayMethod
	if w3c.DisplayMethod != nil {
		displayService = &DisplayMethod{
			Id:   w3c.DisplayMethod.ID,
			Type: DisplayMethodType(w3c.DisplayMethod.Type),
		}
	}

	return Credential{
		CredentialSubject: w3c.CredentialSubject,
		CreatedAt:         TimeUTC(*w3c.IssuanceDate),
		Expired:           expired,
		ExpiresAt:         expiresAt,
		Id:                credential.ID,
		ProofTypes:        proofs,
		RevNonce:          uint64(credential.RevNonce),
		Revoked:           credential.Revoked,
		SchemaHash:        credential.SchemaHash,
		SchemaType:        shortType(credential.SchemaType),
		SchemaUrl:         credential.SchemaURL,
		UserID:            credential.OtherIdentifier,
		RefreshService:    refreshService,
		DisplayMethod:     displayService,
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

func schemaResponse(s *domain.Schema) Schema {
	hash, _ := s.Hash.MarshalText()
	return Schema{
		Id:          s.ID.String(),
		Type:        s.Type,
		Url:         s.URL,
		BigInt:      s.Hash.BigInt().String(),
		Hash:        string(hash),
		CreatedAt:   TimeUTC(s.CreatedAt),
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

func connectionResponse(conn *domain.Connection, w3cs []*verifiable.W3CCredential, credentials []*domain.Claim) GetConnectionResponse {
	credResp := make([]Credential, len(w3cs))
	if w3cs != nil {
		for i := range credentials {
			credResp[i] = credentialResponse(w3cs[i], credentials[i])
		}
	}

	return GetConnectionResponse{
		CreatedAt:   TimeUTC(conn.CreatedAt),
		Id:          conn.ID.String(),
		UserID:      conn.UserDID.String(),
		IssuerID:    conn.IssuerDID.String(),
		Credentials: credResp,
	}
}

func connectionsResponse(conns []domain.Connection) (GetConnectionsResponse, error) {
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
		resp = append(resp, connectionResponse(&conn, w3creds, connCreds))
	}

	return resp, nil
}

func connectionsPaginatedResponse(conns []domain.Connection, pagFilter pagination.Filter, total uint) (ConnectionsPaginated, error) {
	resp, err := connectionsResponse(conns)
	if err != nil {
		return ConnectionsPaginated{}, err
	}

	connsPag := ConnectionsPaginated{
		Items: resp,
		Meta: PaginatedMetadata{
			MaxResults: pagFilter.MaxResults,
			Page:       1, // default
			Total:      total,
		},
	}
	if pagFilter.Page != nil {
		connsPag.Meta.Page = *pagFilter.Page
	}

	return connsPag, nil
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

func stateTransactionsResponse(states []ports.IdentityStatePaginationDto) StateTransactionsResponse {
	stateTransactions := make([]StateTransaction, len(states))
	for i := range states {
		stateTransactions[i] = toStateTransaction(states[i])
	}
	total := 0
	if len(states) > 0 {
		total = states[0].Total
	}
	result := StateTransactionsResponse{
		Transactions: stateTransactions,
		Total:        total,
	}
	return result
}

func toStateTransaction(stateDto ports.IdentityStatePaginationDto) StateTransaction {
	var stateTran, txID string
	state := stateDto.IdentityState
	if state.State != nil {
		stateTran = *state.State
	}
	if state.TxID != nil {
		txID = *state.TxID
	}
	return StateTransaction{
		Id:          state.StateID,
		PublishDate: TimeUTC(state.ModifiedAt),
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
