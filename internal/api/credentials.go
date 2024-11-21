package api

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/internal/schema"
)

// DeleteCredential deletes a credential
func (s *Server) DeleteCredential(ctx context.Context, request DeleteCredentialRequestObject) (DeleteCredentialResponseObject, error) {
	clID, err := uuid.Parse(request.Id)
	if err != nil {
		return DeleteCredential400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	err = s.claimService.Delete(ctx, clID)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return DeleteCredential400JSONResponse{N400JSONResponse{"The given credential does not exist"}}, nil
		}
		return DeleteCredential500JSONResponse{N500JSONResponse{"There was an error deleting the credential"}}, nil
	}

	return DeleteCredential200JSONResponse{Message: "Credential successfully deleted"}, nil
}

// CreateCredential is the creation credential controller. It creates a credential and returns the id
func (s *Server) CreateCredential(ctx context.Context, request CreateCredentialRequestObject) (CreateCredentialResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return CreateCredential400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}
	var expiration *time.Time
	if request.Body.Expiration != nil {
		expiration = common.ToPointer(time.Unix(*request.Body.Expiration, 0))
	}

	claimRequestProofs := ports.ClaimRequestProofs{}
	if request.Body.Proofs == nil {
		claimRequestProofs.BJJSignatureProof2021 = true
		claimRequestProofs.Iden3SparseMerkleTreeProof = true
	} else {
		for _, proof := range *request.Body.Proofs {
			if string(proof) == string(verifiable.BJJSignatureProofType) {
				claimRequestProofs.BJJSignatureProof2021 = true
				continue
			}
			if string(proof) == string(verifiable.Iden3SparseMerkleTreeProofType) {
				claimRequestProofs.Iden3SparseMerkleTreeProof = true
				continue
			}
			return CreateCredential400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("unsupported proof type: %s", proof)}}, nil
		}
	}

	var credentialStatusType *verifiable.CredentialStatusType
	credentialStatusType, err = validateStatusType((*string)(request.Body.CredentialStatusType))
	if err != nil {
		return CreateCredential400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}

	resolverPrefix, err := common.ResolverPrefix(did)
	if err != nil {
		return CreateCredential400JSONResponse{N400JSONResponse{Message: "error parsing did"}}, nil
	}

	rhsSettings, err := s.networkResolver.GetRhsSettings(ctx, resolverPrefix)
	if err != nil {
		return CreateCredential400JSONResponse{N400JSONResponse{Message: "error getting reverse hash service settings"}}, nil
	}

	if !s.networkResolver.IsCredentialStatusTypeSupported(rhsSettings.Mode, *credentialStatusType) {
		log.Warn(ctx, "unsupported credential status type", "req", request)
		return CreateCredential400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("Credential Status Type '%s' is not supported by the issuer", *credentialStatusType)}}, nil
	}

	req := ports.NewCreateClaimRequest(did, request.Body.ClaimID, request.Body.CredentialSchema, request.Body.CredentialSubject, expiration, request.Body.Type, request.Body.Version, request.Body.SubjectPosition, request.Body.MerklizedRootPosition, claimRequestProofs, nil, false, *credentialStatusType, toVerifiableRefreshService(request.Body.RefreshService), request.Body.RevNonce,
		toVerifiableDisplayMethod(request.Body.DisplayMethod))

	resp, err := s.claimService.Save(ctx, req)
	if err != nil {
		if errors.Is(err, services.ErrLoadingSchema) {
			return CreateCredential422JSONResponse{N422JSONResponse{Message: err.Error()}}, nil
		}
		errs := []error{
			services.ErrJSONLdContext,
			services.ErrProcessSchema,
			services.ErrMalformedURL,
			services.ErrParseClaim,
			services.ErrInvalidCredentialSubject,
			services.ErrAssigningMTPProof,
			services.ErrUnsupportedRefreshServiceType,
			services.ErrRefreshServiceLacksExpirationTime,
			services.ErrRefreshServiceLacksURL,
			services.ErrDisplayMethodLacksURL,
			services.ErrUnsupportedDisplayMethodType,
			services.ErrWrongCredentialSubjectID,
			&schema.ParseClaimError{},
		}
		for _, e := range errs {
			if errors.Is(err, e) {
				return CreateCredential400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
			}
		}
		return CreateCredential500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return CreateCredential201JSONResponse{Id: resp.ID.String()}, nil
}

// RevokeCredential is the revocation claim controller
func (s *Server) RevokeCredential(ctx context.Context, request RevokeCredentialRequestObject) (RevokeCredentialResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Warn(ctx, "revoke claim invalid did", "err", err, "req", request)
		return RevokeCredential400JSONResponse{N400JSONResponse{err.Error()}}, nil
	}

	if err := s.claimService.Revoke(ctx, *did, uint64(request.Nonce), ""); err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return RevokeCredential404JSONResponse{N404JSONResponse{
				Message: "the credential does not exist",
			}}, nil
		}

		return RevokeCredential500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return RevokeCredential202JSONResponse{
		Message: "credential revocation request sent",
	}, nil
}

// GetRevocationStatus is the controller to get revocation status
func (s *Server) GetRevocationStatus(ctx context.Context, request GetRevocationStatusRequestObject) (GetRevocationStatusResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetRevocationStatus500JSONResponse{N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	rs, err := s.claimService.GetRevocationStatus(ctx, *issuerDID, uint64(request.Nonce))
	if err != nil {
		return GetRevocationStatus500JSONResponse{N500JSONResponse{
			Message: err.Error(),
		}}, nil
	}

	response := GetRevocationStatus200JSONResponse{}
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

	return response, err
}

// GetRevocationStatusV2 is the controller to get revocation status
func (s *Server) GetRevocationStatusV2(ctx context.Context, request GetRevocationStatusV2RequestObject) (GetRevocationStatusV2ResponseObject, error) {
	response, err := s.GetRevocationStatus(ctx, GetRevocationStatusRequestObject(request))
	if err != nil {
		log.Error(ctx, "getting revocation status", "err", err, "req", request)
		return nil, err
	}
	switch response := response.(type) {
	case GetRevocationStatus200JSONResponse:
		return GetRevocationStatusV2200JSONResponse(response), nil
	case GetRevocationStatus500JSONResponse:
		return GetRevocationStatusV2500JSONResponse(response), nil
	}
	return nil, nil
}

// GetCredentials returns a collection of credentials that matches the request.
func (s *Server) GetCredentials(ctx context.Context, request GetCredentialsRequestObject) (GetCredentialsResponseObject, error) {
	filter, err := getCredentialsFilter(ctx, request)
	if err != nil {
		return GetCredentials400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetCredentials400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	credentials, total, err := s.claimService.GetAll(ctx, *did, filter)
	if err != nil {
		log.Error(ctx, "loading credentials", "err", err, "req", request)
		return GetCredentials500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	response := make([]Credential, len(credentials))
	for i, credential := range credentials {
		w3c, err := schema.FromClaimModelToW3CCredential(*credential)
		if err != nil {
			log.Error(ctx, "creating credentials response", "err", err, "req", request)
			return GetCredentials500JSONResponse{N500JSONResponse{"Invalid claim format"}}, nil
		}
		response[i] = toGetCredential200Response(w3c, credential)
	}

	resp := GetCredentials200JSONResponse{
		Items: response,
		Meta: PaginatedMetadata{
			MaxResults: filter.MaxResults,
			Page:       1, // default
			Total:      total,
		},
	}
	if filter.Page != nil {
		resp.Meta.Page = *filter.Page
	}
	return resp, nil
}

// GetCredential is the controller to get a credential
func (s *Server) GetCredential(ctx context.Context, request GetCredentialRequestObject) (GetCredentialResponseObject, error) {
	if request.Identifier == "" {
		return GetCredential400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetCredential400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	if request.Id == "" {
		return GetCredential400JSONResponse{N400JSONResponse{"cannot proceed with an empty claim id"}}, nil
	}

	clID, err := uuid.Parse(request.Id)
	if err != nil {
		return GetCredential400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	claim, err := s.claimService.GetByID(ctx, did, clID)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return GetCredential404JSONResponse{N404JSONResponse{err.Error()}}, nil
		}
		return GetCredential500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}

	w3c, err := schema.FromClaimModelToW3CCredential(*claim)
	if err != nil {
		return GetCredential500JSONResponse{N500JSONResponse{"invalid claim format"}}, nil
	}

	return GetCredential200JSONResponse(toGetCredential200Response(w3c, claim)), nil
}

// GetCredentialOffer returns a GetCredentialQrCodeResponseObject universalLink, raw or deeplink type based on query parameter `type`
// scan it with privado.id wallet to accept the claim
func (s *Server) GetCredentialOffer(ctx context.Context, request GetCredentialOfferRequestObject) (GetCredentialOfferResponseObject, error) {
	if request.Identifier == "" {
		return GetCredentialOffer400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetCredentialOffer400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	if request.Id == "" {
		return GetCredentialOffer400JSONResponse{N400JSONResponse{"cannot proceed with an empty claim id"}}, nil
	}

	claimID, err := uuid.Parse(request.Id)
	if err != nil {
		return GetCredentialOffer400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	resp, err := s.claimService.GetCredentialQrCode(ctx, did, claimID, s.cfg.ServerUrl)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return GetCredentialOffer404JSONResponse{N404JSONResponse{"QrCode not found"}}, nil
		}
		if errors.Is(err, services.ErrEmptyMTPProof) {
			return GetCredentialOffer409JSONResponse{N409JSONResponse{"State must be published before fetching MTP type credentials"}}, nil
		}
		return GetCredentialOffer500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}
	qrContent, qrType := resp.UniversalLink, GetCredentialOfferParamsTypeUniversalLink
	if request.Params.Type != nil {
		qrType = *request.Params.Type
	}
	switch qrType {
	case GetCredentialOfferParamsTypeDeepLink:
		qrContent = resp.DeepLink
	case GetCredentialOfferParamsTypeUniversalLink:
		qrContent = resp.UniversalLink
	case GetCredentialOfferParamsTypeRaw:
		qrContent = resp.QrRaw
	}

	return GetCredentialOffer200JSONResponse{
		UniversalLink: qrContent,
		SchemaType:    resp.SchemaType,
	}, nil
}

func toVerifiableRefreshService(s *RefreshService) *verifiable.RefreshService {
	if s == nil {
		return nil
	}
	return &verifiable.RefreshService{
		ID:   s.Id,
		Type: verifiable.RefreshServiceType(s.Type),
	}
}

func toVerifiableDisplayMethod(s *DisplayMethod) *verifiable.DisplayMethod {
	if s == nil {
		return nil
	}
	return &verifiable.DisplayMethod{
		ID:   s.Id,
		Type: verifiable.DisplayMethodType(s.Type),
	}
}

func toGetCredential200Response(w3cCredential *verifiable.W3CCredential, cred *domain.Claim) Credential {
	return Credential{
		Vc:         *w3cCredential,
		Id:         cred.ID.String(),
		Revoked:    cred.Revoked,
		SchemaHash: cred.SchemaHash,
		ProofTypes: getProofs(cred),
	}
}

func getCredentialsFilter(ctx context.Context, req GetCredentialsRequestObject) (*ports.ClaimsFilter, error) {
	filter := &ports.ClaimsFilter{}
	if req.Params.CredentialSubject != nil {
		did, err := w3c.ParseDID(*req.Params.CredentialSubject)
		if err != nil {
			log.Warn(ctx, "get credentials. Parsing did", "err", err, "did", did)
			return nil, errors.New("cannot parse did parameter: wrong format")
		}
		filter.Subject, filter.FTSAndCond = did.String(), true
	}
	if req.Params.Status != nil {
		switch GetCredentialsParamsStatus(strings.ToLower(string(*req.Params.Status))) {
		case GetCredentialsParamsStatusRevoked:
			filter.Revoked = common.ToPointer(true)
		case GetCredentialsParamsStatusExpired:
			filter.ExpiredOn = common.ToPointer(time.Now())
		case GetCredentialsParamsStatusAll:
			// Nothing to be done
		default:
			return nil, errors.New("wrong type value. Allowed values: [all, revoked, expired]")
		}
	}
	if req.Params.Query != nil {
		filter.FTSQuery = *req.Params.Query
	}

	filter.MaxResults = 50
	if req.Params.MaxResults != nil {
		if *req.Params.MaxResults <= 0 {
			filter.MaxResults = 50
		} else {
			filter.MaxResults = *req.Params.MaxResults
		}
	}

	if req.Params.Page != nil {
		if *req.Params.Page <= 0 {
			return nil, errors.New("page param must be higher than 0")
		}
		filter.Page = req.Params.Page
	}

	if req.Params.Sort != nil {
		for _, sortBy := range *req.Params.Sort {
			var err error
			field, desc := strings.CutPrefix(strings.TrimSpace(string(sortBy)), "-")
			switch GetCredentialsParamsSort(field) {
			case GetCredentialsParamsSortSchemaType:
				err = filter.OrderBy.Add(ports.CredentialSchemaType, desc)
			case GetCredentialsParamsSortCreatedAt:
				err = filter.OrderBy.Add(ports.CredentialCreatedAt, desc)
			case GetCredentialsParamsSortExpiresAt:
				err = filter.OrderBy.Add(ports.CredentialExpiresAt, desc)
			case GetCredentialsParamsSortRevoked:
				err = filter.OrderBy.Add(ports.CredentialRevoked, desc)
			default:
				return nil, errors.New("wrong sort by value")
			}
			if err != nil {
				return nil, errors.New("repeated sort by value field")
			}
		}
	}
	return filter, nil
}
