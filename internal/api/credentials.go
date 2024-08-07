package api

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/schema"
)

// CreateCredential is the creation credential controller. It creates a credential and returns the id
func (s *Server) CreateCredential(ctx context.Context, request CreateCredentialRequestObject) (CreateCredentialResponseObject, error) {
	ret, err := s.CreateClaim(ctx, CreateClaimRequestObject(request))
	if err != nil {
		return CreateCredential500JSONResponse{N500JSONResponse{Message: err.Error()}}, err
	}
	switch ret := ret.(type) {
	case CreateClaim201JSONResponse:
		return CreateCredential201JSONResponse(ret), nil
	case CreateClaim400JSONResponse:
		return CreateCredential400JSONResponse{N400JSONResponse{Message: ret.Message}}, nil
	case CreateClaim422JSONResponse:
		return CreateCredential422JSONResponse{N422JSONResponse{Message: ret.Message}}, nil
	case CreateClaim500JSONResponse:
		return CreateCredential500JSONResponse{N500JSONResponse{Message: ret.Message}}, nil
	default:
		log.Error(ctx, "unexpected return type", "type", fmt.Sprintf("%T", ret))
		return CreateCredential500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("unexpected return type: %T", ret)}}, nil
	}
}

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

// CreateClaim is claim creation controller
// deprecated - Use CreateCredential instead
func (s *Server) CreateClaim(ctx context.Context, request CreateClaimRequestObject) (CreateClaimResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return CreateClaim400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
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
			return CreateClaim400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("unsupported proof type: %s", proof)}}, nil
		}
	}

	var credentialStatusType verifiable.CredentialStatusType
	if request.Body.CredentialStatusType == nil || *request.Body.CredentialStatusType == "" {
		credentialStatusType = verifiable.Iden3commRevocationStatusV1
	} else {
		allowedCredentialStatuses := []string{string(verifiable.Iden3commRevocationStatusV1), string(verifiable.Iden3ReverseSparseMerkleTreeProof), string(verifiable.Iden3OnchainSparseMerkleTreeProof2023)}
		if !slices.Contains(allowedCredentialStatuses, string(*request.Body.CredentialStatusType)) {
			log.Warn(ctx, "invalid credential status type", "req", request)
			return CreateClaim400JSONResponse{
				N400JSONResponse{
					Message: fmt.Sprintf("Invalid Credential Status Type '%s'. Allowed Iden3commRevocationStatusV1.0, Iden3ReverseSparseMerkleTreeProof or Iden3OnchainSparseMerkleTreeProof2023.", *request.Body.CredentialStatusType),
				},
			}, nil
		}
		credentialStatusType = (verifiable.CredentialStatusType)(*request.Body.CredentialStatusType)
	}

	resolverPrefix, err := common.ResolverPrefix(did)
	if err != nil {
		return CreateClaim400JSONResponse{N400JSONResponse{Message: "error parsing did"}}, nil
	}

	rhsSettings, err := s.networkResolver.GetRhsSettings(ctx, resolverPrefix)
	if err != nil {
		return CreateClaim400JSONResponse{N400JSONResponse{Message: "error getting reverse hash service settings"}}, nil
	}

	if !s.networkResolver.IsCredentialStatusTypeSupported(rhsSettings, credentialStatusType) {
		log.Warn(ctx, "unsupported credential status type", "req", request)
		return CreateClaim400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("Credential Status Type '%s' is not supported by the issuer", credentialStatusType)}}, nil
	}

	req := ports.NewCreateClaimRequest(did, request.Body.ClaimID, request.Body.CredentialSchema, request.Body.CredentialSubject, expiration, request.Body.Type, request.Body.Version, request.Body.SubjectPosition, request.Body.MerklizedRootPosition, claimRequestProofs, nil, false, credentialStatusType, toVerifiableRefreshService(request.Body.RefreshService), request.Body.RevNonce,
		toVerifiableDisplayMethod(request.Body.DisplayMethod))

	resp, err := s.claimService.Save(ctx, req)
	if err != nil {
		if errors.Is(err, services.ErrLoadingSchema) {
			return CreateClaim422JSONResponse{N422JSONResponse{Message: err.Error()}}, nil
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
		}
		for _, e := range errs {
			if errors.Is(err, e) {
				return CreateClaim400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
			}
		}
		return CreateClaim500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return CreateClaim201JSONResponse{Id: resp.ID.String()}, nil
}

// RevokeCredential is the revocation claim controller
func (s *Server) RevokeCredential(ctx context.Context, request RevokeCredentialRequestObject) (RevokeCredentialResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Warn(ctx, "revoke credential: invalid did", "err", err, "req", request)
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

// RevokeClaim is the revocation claim controller.
// Deprecated - use RevokeCredential instead
func (s *Server) RevokeClaim(ctx context.Context, request RevokeClaimRequestObject) (RevokeClaimResponseObject, error) {
	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Warn(ctx, "revoke claim invalid did", "err", err, "req", request)
		return RevokeClaim400JSONResponse{N400JSONResponse{err.Error()}}, nil
	}

	if err := s.claimService.Revoke(ctx, *did, uint64(request.Nonce), ""); err != nil {
		if errors.Is(err, repositories.ErrClaimDoesNotExist) {
			return RevokeClaim404JSONResponse{N404JSONResponse{
				Message: "the claim does not exist",
			}}, nil
		}

		return RevokeClaim500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return RevokeClaim202JSONResponse{
		Message: "claim revocation request sent",
	}, nil
}

// GetRevocationStatusV2 is the controller to get revocation status
func (s *Server) GetRevocationStatusV2(ctx context.Context, request GetRevocationStatusV2RequestObject) (GetRevocationStatusV2ResponseObject, error) {
	resp, err := s.GetRevocationStatus(ctx, GetRevocationStatusRequestObject(request))
	if err != nil {
		return GetRevocationStatusV2500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	switch ret := resp.(type) {
	case GetRevocationStatus200JSONResponse:
		return GetRevocationStatusV2200JSONResponse(ret), nil
	case GetRevocationStatus500JSONResponse:
		return GetRevocationStatusV2500JSONResponse{N500JSONResponse{Message: ret.Message}}, nil
	default:
		log.Error(ctx, "unexpected return type", "type", fmt.Sprintf("%T", ret))
		return GetRevocationStatusV2500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("unexpected return type: %T", ret)}}, nil
	}
}

// GetRevocationStatus is the controller to get revocation status
// Deprecated - use GetRevocationStatusV2 instead
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

// GetCredential is the controller to get a credential
func (s *Server) GetCredential(ctx context.Context, request GetCredentialRequestObject) (GetCredentialResponseObject, error) {
	resp, err := s.GetClaim(ctx, GetClaimRequestObject(request))
	if err != nil {
		return GetCredential500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	switch ret := resp.(type) {
	case GetClaim200JSONResponse:
		return GetCredential200JSONResponse{
			Context: ret.Context,
			CredentialSchema: CredentialSchema{
				Id:   ret.CredentialSchema.Id,
				Type: ret.CredentialSchema.Type,
			},
			CredentialStatus:  ret.CredentialStatus,
			CredentialSubject: ret.CredentialSubject,
			DisplayMethod:     ret.DisplayMethod,
			ExpirationDate:    ret.ExpirationDate,
			Id:                ret.Id,
			IssuanceDate:      ret.IssuanceDate,
			Issuer:            ret.Issuer,
			Proof:             ret.Proof,
			RefreshService:    ret.RefreshService,
			Type:              ret.Type,
		}, nil
	case GetClaim400JSONResponse:
		return GetCredential400JSONResponse{N400JSONResponse{Message: ret.Message}}, nil
	case GetClaim404JSONResponse:
		return GetCredential404JSONResponse{N404JSONResponse{Message: ret.Message}}, nil
	case GetClaim500JSONResponse:
		return GetCredential500JSONResponse{N500JSONResponse{Message: ret.Message}}, nil
	default:
		log.Error(ctx, "unexpected return type", "type", fmt.Sprintf("%T", ret))
		return GetCredential500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("unexpected return type: %T", ret)}}, nil
	}
}

// GetCredentialsPaginated returns a collection of credentials that matches the request.
func (s *Server) GetCredentialsPaginated(ctx context.Context, request GetCredentialsPaginatedRequestObject) (GetCredentialsPaginatedResponseObject, error) {
	filter, err := getCredentialsFilter(ctx, request)
	if err != nil {
		return GetCredentialsPaginated400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetCredentialsPaginated400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	credentials, total, err := s.claimService.GetAll(ctx, *did, filter)
	if err != nil {
		log.Error(ctx, "loading credentials", "err", err, "req", request)
		return GetCredentialsPaginated500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	response := make([]Credential, len(credentials))
	for i, credential := range credentials {
		w3c, err := schema.FromClaimModelToW3CCredential(*credential)
		if err != nil {
			log.Error(ctx, "creating credentials response", "err", err, "req", request)
			return GetCredentialsPaginated500JSONResponse{N500JSONResponse{"Invalid claim format"}}, nil
		}
		response[i] = credentialResponse(w3c, credential)
	}

	resp := GetCredentialsPaginated200JSONResponse{
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

// GetClaim is the controller to get a claim.
// Deprecated - use GetCredential instead
func (s *Server) GetClaim(ctx context.Context, request GetClaimRequestObject) (GetClaimResponseObject, error) {
	if request.Identifier == "" {
		return GetClaim400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetClaim400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	if request.Id == "" {
		return GetClaim400JSONResponse{N400JSONResponse{"cannot proceed with an empty claim id"}}, nil
	}

	clID, err := uuid.Parse(request.Id)
	if err != nil {
		return GetClaim400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	claim, err := s.claimService.GetByID(ctx, did, clID)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return GetClaim404JSONResponse{N404JSONResponse{err.Error()}}, nil
		}
		return GetClaim500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}

	w3c, err := schema.FromClaimModelToW3CCredential(*claim)
	if err != nil {
		return GetClaim500JSONResponse{N500JSONResponse{"invalid claim format"}}, nil
	}

	return GetClaim200JSONResponse(toGetClaim200Response(w3c)), nil
}

// GetCredentials is the controller to get multiple credentials of a determined identity
func (s *Server) GetCredentials(ctx context.Context, request GetCredentialsRequestObject) (GetCredentialsResponseObject, error) {
	resp, err := s.GetClaims(ctx, GetClaimsRequestObject{
		Identifier: request.Identifier,
		Params:     GetClaimsParams(request.Params),
	})
	if err != nil {
		return GetCredentials500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	switch ret := resp.(type) {
	case GetClaims200JSONResponse:
		return GetCredentials200JSONResponse(ret), nil
	case GetClaims400JSONResponse:
		return GetCredentials400JSONResponse{N400JSONResponse{Message: ret.Message}}, nil
	case GetClaims500JSONResponse:
		return GetCredentials500JSONResponse{N500JSONResponse{Message: ret.Message}}, nil
	default:
		log.Error(ctx, "unexpected return type", "type", fmt.Sprintf("%T", ret))
		return GetCredentials500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("unexpected return type: %T", ret)}}, nil
	}
}

// GetClaims is the controller to get multiple claims of a determined identity
// deprecated: use GetCredentials instead
func (s *Server) GetClaims(ctx context.Context, request GetClaimsRequestObject) (GetClaimsResponseObject, error) {
	if request.Identifier == "" {
		return GetClaims400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetClaims400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	filter, err := ports.NewClaimsFilter(
		request.Params.SchemaHash,
		request.Params.SchemaType,
		request.Params.Subject,
		request.Params.QueryField,
		request.Params.QueryValue,
		request.Params.Self,
		request.Params.Revoked)
	if err != nil {
		return GetClaims400JSONResponse{N400JSONResponse{err.Error()}}, nil
	}

	claims, _, err := s.claimService.GetAll(ctx, *did, filter)
	if err != nil && !errors.Is(err, services.ErrCredentialNotFound) {
		return GetClaims500JSONResponse{N500JSONResponse{"there was an internal error trying to retrieve claims for the requested identifier"}}, nil
	}

	w3Claims, err := schema.FromClaimsModelToW3CCredential(claims)
	if err != nil {
		return GetClaims500JSONResponse{N500JSONResponse{"there was an internal error parsing the claims"}}, nil
	}

	return GetClaims200JSONResponse(toGetClaims200Response(w3Claims)), nil
}

// GetCredentialQrCode returns a GetCredentialQrCodeResponseObject raw or link type based on query parameter `type`
// scan it with privado.id wallet to accept the claim
func (s *Server) GetCredentialQrCode(ctx context.Context, request GetCredentialQrCodeRequestObject) (GetCredentialQrCodeResponseObject, error) {
	if request.Identifier == "" {
		return GetCredentialQrCode400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetCredentialQrCode400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	if request.Id == "" {
		return GetCredentialQrCode400JSONResponse{N400JSONResponse{"cannot proceed with an empty claim id"}}, nil
	}

	claimID, err := uuid.Parse(request.Id)
	if err != nil {
		return GetCredentialQrCode400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	resp, err := s.claimService.GetCredentialQrCode(ctx, did, claimID, s.cfg.ServerUrl)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return GetCredentialQrCode404JSONResponse{N404JSONResponse{"Credential not found"}}, nil
		}
		if errors.Is(err, services.ErrEmptyMTPProof) {
			return GetCredentialQrCode409JSONResponse{N409JSONResponse{"State must be published before fetching MTP type credentials"}}, nil
		}
		return GetCredentialQrCode500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}
	qrContent := resp.QrCodeURL
	// Backward compatibility. If the type is raw, we return the raw qr code
	if request.Params.Type != nil && *request.Params.Type == GetCredentialQrCodeParamsTypeRaw {
		rawQrCode, err := s.qrService.Find(ctx, resp.QrID)
		if err != nil {
			log.Error(ctx, "qr store. Finding qr", "err", err, "id", resp.QrID)
			return GetCredentialQrCode500JSONResponse{N500JSONResponse{"error looking for qr body"}}, nil
		}
		qrContent = string(rawQrCode)
	}
	return GetCredentialQrCode200JSONResponse{
		QrCodeLink: qrContent,
		SchemaType: resp.SchemaType,
	}, nil
}

// GetClaimQrCode returns a GetClaimQrCodeResponseObject that can be used with any QR generator to create a QR and
// scan it with polygon wallet to accept the claim
// deprecated - use GetCredentialQrCode instead
func (s *Server) GetClaimQrCode(ctx context.Context, request GetClaimQrCodeRequestObject) (GetClaimQrCodeResponseObject, error) {
	if request.Identifier == "" {
		return GetClaimQrCode400JSONResponse{N400JSONResponse{"invalid did, cannot be empty"}}, nil
	}

	did, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		return GetClaimQrCode400JSONResponse{N400JSONResponse{"invalid did"}}, nil
	}

	if request.Id == "" {
		return GetClaimQrCode400JSONResponse{N400JSONResponse{"cannot proceed with an empty claim id"}}, nil
	}

	claimID, err := uuid.Parse(request.Id)
	if err != nil {
		return GetClaimQrCode400JSONResponse{N400JSONResponse{"invalid claim id"}}, nil
	}

	claim, err := s.claimService.GetByID(ctx, did, claimID)
	if err != nil {
		if errors.Is(err, services.ErrCredentialNotFound) {
			return GetClaimQrCode404JSONResponse{N404JSONResponse{err.Error()}}, nil
		}
		return GetClaimQrCode500JSONResponse{N500JSONResponse{err.Error()}}, nil
	}

	if !claim.ValidProof() {
		return GetClaimQrCode409JSONResponse{N409JSONResponse{"State must be published before fetching MTP type credential"}}, nil
	}

	return toGetClaimQrCode200JSONResponse(claim, s.cfg.ServerUrl), nil
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

func toGetClaims200Response(claims []*verifiable.W3CCredential) GetClaimsResponse {
	response := make(GetClaims200JSONResponse, len(claims))
	for i := range claims {
		response[i] = toGetClaim200Response(claims[i])
	}

	return response
}

func toGetClaim200Response(claim *verifiable.W3CCredential) GetClaimResponse {
	var claimExpiration, claimIssuanceDate *TimeUTC
	if claim.Expiration != nil {
		claimExpiration = common.ToPointer(TimeUTC(*claim.Expiration))
	}
	if claim.IssuanceDate != nil {
		claimIssuanceDate = common.ToPointer(TimeUTC(*claim.IssuanceDate))
	}

	var refreshService *RefreshService
	if claim.RefreshService != nil {
		refreshService = &RefreshService{
			Id:   claim.RefreshService.ID,
			Type: RefreshServiceType(claim.RefreshService.Type),
		}
	}

	var displayMethod *DisplayMethod
	if claim.DisplayMethod != nil {
		displayMethod = &DisplayMethod{
			Id:   claim.DisplayMethod.ID,
			Type: DisplayMethodType(claim.DisplayMethod.Type),
		}
	}

	return GetClaimResponse{
		Context: claim.Context,
		CredentialSchema: CredentialSchema{
			claim.CredentialSchema.ID,
			claim.CredentialSchema.Type,
		},
		CredentialStatus:  claim.CredentialStatus,
		CredentialSubject: claim.CredentialSubject,
		ExpirationDate:    claimExpiration,
		Id:                claim.ID,
		IssuanceDate:      claimIssuanceDate,
		Issuer:            claim.Issuer,
		Proof:             claim.Proof,
		Type:              claim.Type,
		RefreshService:    refreshService,
		DisplayMethod:     displayMethod,
	}
}

func toGetClaimQrCode200JSONResponse(claim *domain.Claim, hostURL string) GetClaimQrCode200JSONResponse {
	id := uuid.New()
	return GetClaimQrCode200JSONResponse{
		Body: struct {
			Credentials []struct {
				Description string `json:"description"`
				Id          string `json:"id"`
			} `json:"credentials"`
			Url string `json:"url"`
		}{
			Credentials: []struct {
				Description string `json:"description"`
				Id          string `json:"id"`
			}{
				{
					Description: claim.SchemaType,
					Id:          claim.ID.String(),
				},
			},
			Url: fmt.Sprintf("%s/v1/agent", strings.TrimSuffix(hostURL, "/")),
		},
		From: claim.Issuer,
		Id:   id.String(),
		Thid: id.String(),
		To:   claim.OtherIdentifier,
		Typ:  string(packers.MediaTypePlainMessage),
		Type: string(protocol.CredentialOfferMessageType),
	}
}

func getCredentialsFilter(ctx context.Context, req GetCredentialsPaginatedRequestObject) (*ports.ClaimsFilter, error) {
	filter := &ports.ClaimsFilter{}
	if req.Params.Did != nil {
		did, err := w3c.ParseDID(*req.Params.Did)
		if err != nil {
			log.Warn(ctx, "get credentials. Parsing did", "err", err, "did", did)
			return nil, errors.New("cannot parse did parameter: wrong format")
		}
		filter.Subject, filter.FTSAndCond = did.String(), true
	}
	if req.Params.Status != nil {
		switch GetCredentialsPaginatedParamsStatus(strings.ToLower(string(*req.Params.Status))) {
		case Revoked:
			filter.Revoked = common.ToPointer(true)
		case Expired:
			filter.ExpiredOn = common.ToPointer(time.Now())
		case All:
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
			switch GetCredentialsPaginatedParamsSort(field) {
			case GetCredentialsPaginatedParamsSortSchemaType:
				err = filter.OrderBy.Add(ports.CredentialSchemaType, desc)
			case GetCredentialsPaginatedParamsSortCreatedAt:
				err = filter.OrderBy.Add(ports.CredentialCreatedAt, desc)
			case GetCredentialsPaginatedParamsSortExpiresAt:
				err = filter.OrderBy.Add(ports.CredentialExpiresAt, desc)
			case GetCredentialsPaginatedParamsSortRevoked:
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
