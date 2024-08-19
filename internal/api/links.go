package api

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/common"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
	"github.com/polygonid/sh-id-platform/pkg/link"
)

// GetLinks - Returns a list of links based on a search criteria.
func (s *Server) GetLinks(ctx context.Context, request GetLinksRequestObject) (GetLinksResponseObject, error) {
	var err error
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetLinks400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	status := ports.LinkAll
	if request.Params.Status != nil {
		if status, err = ports.LinkTypeReqFromString(string(*request.Params.Status)); err != nil {
			log.Warn(ctx, "unknown request type getting links", "err", err, "type", request.Params.Status)
			return GetLinks400JSONResponse{N400JSONResponse{Message: "unknown request type. Allowed: all|active|inactive|exceed"}}, nil
		}
	}
	links, err := s.linkService.GetAll(ctx, *issuerDID, status, request.Params.Query)
	if err != nil {
		log.Error(ctx, "getting links", "err", err, "req", request)
	}

	return GetLinks200JSONResponse(getLinkResponses(links)), err
}

// CreateLink - creates a link for issuing a credential
func (s *Server) CreateLink(ctx context.Context, request CreateLinkRequestObject) (CreateLinkResponseObject, error) {
	if request.Body.Expiration != nil {
		if request.Body.Expiration.Before(time.Now()) {
			return CreateLink400JSONResponse{N400JSONResponse{Message: "invalid claimLinkExpiration. Cannot be a date time prior current time."}}, nil
		}
	}
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreateLink400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if !request.Body.MtProof && !request.Body.SignatureProof {
		return CreateLink400JSONResponse{N400JSONResponse{Message: "at least one proof type should be enabled"}}, nil
	}
	if len(request.Body.CredentialSubject) == 0 {
		return CreateLink400JSONResponse{N400JSONResponse{Message: "you must provide at least one attribute"}}, nil
	}

	credSubject := make(domain.CredentialSubject, len(request.Body.CredentialSubject))
	for key, val := range request.Body.CredentialSubject {
		credSubject[key] = val
	}

	if request.Body.LimitedClaims != nil {
		if *request.Body.LimitedClaims <= 0 {
			return CreateLink400JSONResponse{N400JSONResponse{Message: "limitedClaims must be higher than 0"}}, nil
		}
	}

	var expirationDate *time.Time
	if request.Body.CredentialExpiration != nil {
		expirationDate = request.Body.CredentialExpiration
	}

	createdLink, err := s.linkService.Save(ctx, *issuerDID, request.Body.LimitedClaims, request.Body.Expiration, request.Body.SchemaID, expirationDate, request.Body.SignatureProof, request.Body.MtProof, credSubject, toVerifiableRefreshService(request.Body.RefreshService), toDisplayMethodService(request.Body.DisplayMethod))
	if err != nil {
		log.Error(ctx, "error saving the link", "err", err.Error())
		if errors.Is(err, services.ErrLoadingSchema) {
			return CreateLink500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
		}
		return CreateLink400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}
	return CreateLink201JSONResponse{Id: createdLink.ID.String()}, nil
}

// CreateLinkQrCodeCallback - Callback endpoint for the link qr code creation.
func (s *Server) CreateLinkQrCodeCallback(ctx context.Context, request CreateLinkQrCodeCallbackRequestObject) (CreateLinkQrCodeCallbackResponseObject, error) {
	if request.Body == nil || *request.Body == "" {
		log.Error(ctx, "empty request body auth-callback request")
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{"Cannot proceed with empty body"}}, nil
	}

	arm, err := s.identityService.Authenticate(ctx, *request.Body, request.Params.SessionID, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "error authenticating", err.Error())
		return CreateLinkQrCodeCallback500JSONResponse{}, nil
	}

	userDID, err := w3c.ParseDID(arm.From)
	if err != nil {
		log.Error(ctx, "error getting user DID", err.Error())
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "expecting a did in From"}}, nil
	}
	issuerDID, err := w3c.ParseDID(arm.To)
	if err != nil {
		log.Error(ctx, "error getting issuer DID", err.Error())
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "expecting a did in To"}}, nil
	}

	var credentialStatusType verifiable.CredentialStatusType
	if request.Params.CredentialStatusType == nil || *request.Params.CredentialStatusType == "" {
		credentialStatusType = verifiable.Iden3commRevocationStatusV1
	} else {
		allowedCredentialStatuses := []string{string(verifiable.Iden3commRevocationStatusV1), string(verifiable.Iden3ReverseSparseMerkleTreeProof), string(verifiable.Iden3OnchainSparseMerkleTreeProof2023)}
		if !slices.Contains(allowedCredentialStatuses, *request.Params.CredentialStatusType) {
			log.Warn(ctx, "invalid credential status type", "req", request)
			return CreateLinkQrCodeCallback400JSONResponse{
				N400JSONResponse{
					Message: fmt.Sprintf("Invalid Credential Status Type '%s'. Allowed Iden3commRevocationStatusV1.0, Iden3ReverseSparseMerkleTreeProof or Iden3OnchainSparseMerkleTreeProof2023.", *request.Params.CredentialStatusType),
				},
			}, nil
		}
		credentialStatusType = (verifiable.CredentialStatusType)(*request.Params.CredentialStatusType)
	}

	resolverPrefix, err := common.ResolverPrefix(issuerDID)
	if err != nil {
		log.Error(ctx, "error getting resolver prefix", "err", err, "did", issuerDID)
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "error parsing issuer did"}}, nil
	}

	rhsSettings, err := s.networkResolver.GetRhsSettings(ctx, resolverPrefix)
	if err != nil {
		log.Error(ctx, "error getting reverse hash service settings", "err", err)
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "error getting reverse hash service settings"}}, nil
	}

	if !s.networkResolver.IsCredentialStatusTypeSupported(rhsSettings, credentialStatusType) {
		log.Error(ctx, "unsupported credential status type", "type", credentialStatusType)
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("Credential Status Type '%s' is not supported by the issuer", credentialStatusType)}}, nil
	}

	err = s.linkService.IssueClaim(ctx, request.Params.SessionID.String(), *issuerDID, *userDID, request.Params.LinkID, s.cfg.ServerUrl, credentialStatusType)
	if err != nil {
		log.Error(ctx, "error issuing the claim", "error", err)
		return CreateLinkQrCodeCallback500JSONResponse{}, nil
	}

	return CreateLinkQrCodeCallback200Response{}, nil
}

// DeleteLink - delete a link
func (s *Server) DeleteLink(ctx context.Context, request DeleteLinkRequestObject) (DeleteLinkResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return DeleteLink400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if err := s.linkService.Delete(ctx, request.Id, *issuerDID); err != nil {
		if errors.Is(err, repositories.ErrLinkDoesNotExist) {
			return DeleteLink400JSONResponse{N400JSONResponse{Message: "link does not exist"}}, nil
		}
		return DeleteLink500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return DeleteLink200JSONResponse{Message: "link deleted"}, nil
}

// GetLink returns a link from an id
func (s *Server) GetLink(ctx context.Context, request GetLinkRequestObject) (GetLinkResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetLink400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	link, err := s.linkService.GetByID(ctx, *issuerDID, request.Id)
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			return GetLink404JSONResponse{N404JSONResponse{Message: "link not found"}}, nil
		}
		log.Error(ctx, "obtaining a link", "err", err.Error(), "id", request.Id)
		return GetLink500JSONResponse{N500JSONResponse{Message: "error getting link"}}, nil
	}
	return GetLink200JSONResponse(getLinkResponse(*link)), nil
}

// ActivateLink - Activates or deactivates a link
func (s *Server) ActivateLink(ctx context.Context, request ActivateLinkRequestObject) (ActivateLinkResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return ActivateLink400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if err := s.linkService.Activate(ctx, *issuerDID, request.Id, request.Body.Active); err != nil {
		if errors.Is(err, repositories.ErrLinkDoesNotExist) || errors.Is(err, services.ErrLinkAlreadyActive) || errors.Is(err, services.ErrLinkAlreadyInactive) {
			return ActivateLink400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
		}
		log.Error(ctx, "error activating or deactivating link", err.Error(), "id", request.Id)
		return ActivateLink500JSONResponse{N500JSONResponse{Message: err.Error()}}, nil
	}
	return ActivateLink200JSONResponse{Message: "Link updated"}, nil
}

// GetLinkQRCode - returns te qr code for adding the credential
func (s *Server) GetLinkQRCode(ctx context.Context, request GetLinkQRCodeRequestObject) (GetLinkQRCodeResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetLinkQRCode400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	getQRCodeResponse, err := s.linkService.GetQRCode(ctx, request.Params.SessionID, *issuerDID, request.Id)
	if err != nil {
		if errors.Is(services.ErrLinkNotFound, err) {
			return GetLinkQRCode404JSONResponse{Message: "error: link not found"}, nil
		}
		return GetLinkQRCode400JSONResponse{N400JSONResponse{Message: err.Error()}}, nil
	}

	if getQRCodeResponse.State.Status == link.StatusPending || getQRCodeResponse.State.Status == link.StatusDone || getQRCodeResponse.State.Status == link.StatusPendingPublish {
		return GetLinkQRCode200JSONResponse{
			Status:     common.ToPointer(getQRCodeResponse.State.Status),
			QrCode:     getQRCodeResponse.State.QRCode,
			LinkDetail: getLinkSimpleResponse(*getQRCodeResponse.Link),
		}, nil
	}

	return GetLinkQRCode400JSONResponse{N400JSONResponse{
		Message: fmt.Sprintf("error fetching the link qr code: %s", err),
	}}, nil
}

// CreateLinkQrCode - Creates a link QrCode
func (s *Server) CreateLinkQrCode(ctx context.Context, req CreateLinkQrCodeRequestObject) (CreateLinkQrCodeResponseObject, error) {
	issuerDID, err := w3c.ParseDID(req.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", req.Identifier)
		return CreateLinkQrCode400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	createLinkQrCodeResponse, err := s.linkService.CreateQRCode(ctx, *issuerDID, req.Id, s.cfg.ServerUrl)
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			return CreateLinkQrCode404JSONResponse{N404JSONResponse{Message: "error: link not found"}}, nil
		}
		if errors.Is(err, services.ErrLinkAlreadyExpired) || errors.Is(err, services.ErrLinkMaxExceeded) || errors.Is(err, services.ErrLinkInactive) {
			return CreateLinkQrCode404JSONResponse{N404JSONResponse{Message: "error: " + err.Error()}}, nil
		}
		log.Error(ctx, "Unexpected error while creating qr code", "err", err)
		return CreateLinkQrCode500JSONResponse{N500JSONResponse{"Unexpected error while creating qr code"}}, nil
	}

	// Backward compatibility. If the type is raw, we return the raw qr code
	qrCodeRaw, err := s.qrService.Find(ctx, createLinkQrCodeResponse.QrID)
	if err != nil {
		log.Error(ctx, "qr store. Finding qr", "err", err, "id", createLinkQrCodeResponse.QrID)
		return CreateLinkQrCode500JSONResponse{N500JSONResponse{"error looking for qr body"}}, nil
	}

	return CreateLinkQrCode200JSONResponse{
		Issuer: IssuerDescription{
			DisplayName: s.cfg.IssuerName,
			Logo:        s.cfg.IssuerLogo,
		},
		QrCodeLink: createLinkQrCodeResponse.QrCode,
		QrCodeRaw:  string(qrCodeRaw),
		SessionID:  createLinkQrCodeResponse.SessionID,
		LinkDetail: getLinkSimpleResponse(*createLinkQrCodeResponse.Link),
	}, nil
}

func toDisplayMethodService(s *DisplayMethod) *verifiable.DisplayMethod {
	if s == nil {
		return nil
	}
	return &verifiable.DisplayMethod{
		ID:   s.Id,
		Type: verifiable.DisplayMethodType(s.Type),
	}
}
