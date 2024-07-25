package api

import (
	"net/http"

	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
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
