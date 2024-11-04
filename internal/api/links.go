package api

import (
	"context"
	"errors"
	"time"

	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/repositories"
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

	filter := ports.LinksFilter{
		Status: status,
		Query:  request.Params.Query,
		Page:   request.Params.Page,
	}

	filter.MaxResults = 10
	if request.Params.MaxResults != nil {
		if *request.Params.MaxResults <= 0 {
			filter.MaxResults = 10
		} else {
			filter.MaxResults = *request.Params.MaxResults
		}
	}

	links, total, err := s.linkService.GetAll(ctx, *issuerDID, filter, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "getting links", "err", err, "req", request)
	}

	resp := GetLinks200JSONResponse{
		Items: getLinkResponses(links),
		Meta: PaginatedMetadata{
			MaxResults: filter.MaxResults,
			Page:       1, // default
			Total:      total,
		},
	}
	if filter.Page != nil {
		resp.Meta.Page = *filter.Page
	}
	return resp, err
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
// It's processed after the user scans the qr code and the mobile app sends the callback.
func (s *Server) CreateLinkQrCodeCallback(ctx context.Context, request CreateLinkQrCodeCallbackRequestObject) (CreateLinkQrCodeCallbackResponseObject, error) {
	if request.Body == nil || *request.Body == "" {
		log.Error(ctx, "empty request body auth-callback request")
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{"Cannot proceed with empty body"}}, nil
	}

	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	offer, err := s.linkService.ProcessCallBack(ctx, *issuerDID, *request.Body, request.Params.LinkID, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "error issuing the claim", "error", err)
		if errors.Is(err, services.ErrLinkAlreadyExpired) || errors.Is(err, services.ErrLinkMaxExceeded) || errors.Is(err, services.ErrLinkInactive) {
			return CreateLinkQrCodeCallback400JSONResponse{N400JSONResponse{Message: "error: " + err.Error()}}, nil
		}
		return CreateLinkQrCodeCallback500JSONResponse{
			N500JSONResponse{
				Message: "error processing the callback",
			},
		}, nil
	}

	var offerResponse CreateLinkQrCodeCallback200JSONResponse
	if offer != nil {
		offerResponse = CreateLinkQrCodeCallback200JSONResponse{
			Body:     offer.Body,
			From:     offer.From,
			ThreadID: offer.ThreadID,
			ID:       offer.ID,
			To:       offer.To,
			Typ:      offer.Typ,
			Type:     offer.Type,
		}
	}

	return offerResponse, nil
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
		if errors.Is(err, repositories.ErrorLinkWithClaims); err != nil {
			return DeleteLink400JSONResponse{N400JSONResponse{Message: "link has claims associated"}}, nil
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
	link, err := s.linkService.GetByID(ctx, *issuerDID, request.Id, s.cfg.ServerUrl)
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			return GetLink404JSONResponse{N404JSONResponse{Message: "link not found"}}, nil
		}
		log.Error(ctx, "obtaining a link", "err", err.Error(), "id", request.Id)
		return GetLink500JSONResponse{N500JSONResponse{Message: "error getting link"}}, nil
	}
	return GetLink200JSONResponse(getLinkResponse(link)), nil
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

// CreateLinkOffer - Creates a link offer (qr code)
func (s *Server) CreateLinkOffer(ctx context.Context, req CreateLinkOfferRequestObject) (CreateLinkOfferResponseObject, error) {
	issuerDID, err := w3c.ParseDID(req.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", req.Identifier)
		return CreateLinkOffer400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	createLinkQrCodeResponse, err := s.linkService.CreateQRCode(ctx, *issuerDID, req.Id, s.cfg.ServerUrl)
	if err != nil {
		if errors.Is(err, services.ErrLinkNotFound) {
			return CreateLinkOffer404JSONResponse{N404JSONResponse{Message: "error: link not found"}}, nil
		}
		if errors.Is(err, services.ErrLinkAlreadyExpired) || errors.Is(err, services.ErrLinkMaxExceeded) || errors.Is(err, services.ErrLinkInactive) {
			return CreateLinkOffer404JSONResponse{N404JSONResponse{Message: "error: " + err.Error()}}, nil
		}
		log.Error(ctx, "Unexpected error while creating qr code", "err", err)
		return CreateLinkOffer500JSONResponse{N500JSONResponse{"Unexpected error while creating qr code"}}, nil
	}

	return CreateLinkOffer200JSONResponse{
		Issuer: IssuerDescription{
			DisplayName: s.cfg.IssuerName,
			Logo:        s.cfg.IssuerLogo,
		},
		DeepLink:      createLinkQrCodeResponse.DeepLink,
		UniversalLink: createLinkQrCodeResponse.UniversalLink,
		Message:       createLinkQrCodeResponse.QrCodeRaw,
		LinkDetail:    getLinkSimpleResponse(*createLinkQrCodeResponse.Link),
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
