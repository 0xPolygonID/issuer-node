package api

import (
	"encoding/json"
	"net/http"

	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/pagination"
	"github.com/polygonid/sh-id-platform/internal/schema"
	"github.com/polygonid/sh-id-platform/internal/timeapi"
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

func getLinkResponses(links []*domain.Link) []Link {
	res := make([]Link, len(links))
	for i, link := range links {
		res[i] = getLinkResponse(link)
	}
	return res
}

func getLinkResponse(link *domain.Link) Link {
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
		ProofTypes:           getLinkProofs(*link),
		CreatedAt:            TimeUTC(link.CreatedAt),
		Expiration:           validUntil,
		CredentialExpiration: credentialExpiration,
		RefreshService:       refreshService,
		DisplayMethod:        displayMethod,
		DeepLink:             link.DeepLink,
		UniversalLink:        link.UniversalLink,
	}
}

func getLinkProofs(link domain.Link) []string {
	proofs := make([]string, 0)
	if link.CredentialMTPProof {
		proofs = append(proofs, string(verifiable.Iden3SparseMerkleTreeProofType))
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
		proofs = append(proofs, string(verifiable.Iden3SparseMerkleTreeProofType))
	}

	return proofs
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

func connectionResponse(conn *domain.Connection, credentials []*domain.Claim) (GetConnectionResponse, error) {
	credResp := make([]Credential, len(credentials))
	for i := range credentials {
		w3Cred, err := schema.FromClaimModelToW3CCredential(*credentials[i])
		if err != nil {
			return GetConnectionResponse{}, err
		}
		credResp[i] = toGetCredential200Response(w3Cred, credentials[i])
	}
	return GetConnectionResponse{
		CreatedAt:   TimeUTC(conn.CreatedAt),
		Id:          conn.ID.String(),
		UserID:      conn.UserDID.String(),
		IssuerID:    conn.IssuerDID.String(),
		Credentials: credResp,
	}, nil
}

func connectionsResponse(conns []domain.Connection) (GetConnectionsResponse, error) {
	resp := make([]GetConnectionResponse, 0)

	for _, conn := range conns {
		var credentials []*domain.Claim
		if conn.Credentials != nil {
			credentials = *conn.Credentials
		}
		connResp, err := connectionResponse(&conn, credentials)
		if err != nil {
			return GetConnectionsResponse{}, err
		}
		resp = append(resp, connResp)
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

func stateTransactionsPaginatedResponse(idState []domain.IdentityState, pagFilter pagination.Filter, total uint) StateTransactionsPaginated {
	states := make([]StateTransaction, 0)
	for _, state := range idState {
		states = append(states, toStateTransaction(state))
	}
	statesPag := StateTransactionsPaginated{
		Items: states,
		Meta: PaginatedMetadata{
			MaxResults: pagFilter.MaxResults,
			Page:       1, // default
			Total:      total,
		},
	}
	if pagFilter.Page != nil {
		statesPag.Meta.Page = *pagFilter.Page
	}
	return statesPag
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

func toGetPaymentOptionsResponse(opts []domain.PaymentOption) (PaymentOptions, error) {
	var err error
	res := make([]PaymentOption, len(opts))
	for i, opt := range opts {
		res[i], err = toPaymentOption(&opt)
		if err != nil {
			return PaymentOptions{}, err
		}
	}
	return res, nil
}

func toPaymentOption(opt *domain.PaymentOption) (PaymentOption, error) {
	var config map[string]interface{}
	raw, err := json.Marshal(opt.Config)
	if err != nil {
		return PaymentOption{}, err
	}
	if err := json.Unmarshal(raw, &config); err != nil {
		return PaymentOption{}, err
	}
	return PaymentOption{
		Id:          opt.ID,
		IssuerDID:   opt.IssuerDID.String(),
		Name:        opt.Name,
		Description: opt.Description,
		Config:      toPaymentOptionConfig(opt.Config),
		CreatedAt:   TimeUTC(opt.CreatedAt),
		ModifiedAt:  TimeUTC(opt.UpdatedAt),
	}, nil
}

func toPaymentOptionConfig(config domain.PaymentOptionConfig) PaymentOptionConfig {
	cfg := PaymentOptionConfig{
		Config: make([]PaymentOptionConfigItem, len(config.Config)),
	}
	for i, item := range config.Config {
		cfg.Config[i] = PaymentOptionConfigItem{
			PaymentOptionId: item.PaymentOptionID,
			Amount:          item.Amount,
			Recipient:       item.Recipient.String(),
			SigningKeyId:    item.SigningKeyID,
		}
	}
	return cfg
}
