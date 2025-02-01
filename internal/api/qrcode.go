package api

import (
	"context"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// GetQrFromStore is the controller to get qr bodies
func (s *Server) GetQrFromStore(ctx context.Context, request GetQrFromStoreRequestObject) (GetQrFromStoreResponseObject, error) {
	if request.Params.Id == nil {
		log.Warn(ctx, "qr store. Missing id parameter")
		return GetQrFromStore400JSONResponse{N400JSONResponse{"id is required"}}, nil
	}
	body, err := s.qrService.Find(ctx, *request.Params.Id)
	if err != nil {
		log.Error(ctx, "qr store. Finding qr", "err", err, "id", *request.Params.Id)

		if request.Params.Issuer == nil {
			return GetQrFromStore400JSONResponse{N400JSONResponse{"error looking for qr body"}}, nil
		}

		issuerDID, err := w3c.ParseDID(*request.Params.Issuer)
		if err != nil {
			log.Error(ctx, "parsing issuer did", "err", err, "did", *request.Params.Issuer)
			return GetQrFromStore400JSONResponse{N400JSONResponse{"invalid issuer did"}}, nil
		}

		link, err := s.linkService.GetByID(ctx, *issuerDID, *request.Params.Id, s.cfg.ServerUrl)
		if err != nil {
			log.Error(ctx, "getting link by id", "err", err, "link id", *request.Params.Id)
			return GetQrFromStore404JSONResponse{N404JSONResponse{"link not found"}}, nil
		}
		if err := s.linkService.Validate(ctx, link); err != nil {
			log.Error(ctx, "validating link", "err", err, "link id", *request.Params.Id)
			return GetQrFromStore410JSONResponse{N410JSONResponse{err.Error()}}, nil
		}

		if link.AuthorizationRequestMessage == nil {
			log.Error(ctx, "qr store. Finding qr", "err", err, "id", *request.Params.Id)
			return GetQrFromStore400JSONResponse{N400JSONResponse{"error looking for qr body"}}, nil
		}

		return NewQrContentResponse(link.AuthorizationRequestMessage.Bytes), nil
	}
	return NewQrContentResponse(body), nil
}

// GetQrFromStorePublic is the controller to get qr bodies
func (s *Server) GetQrFromStorePublic(ctx context.Context, request GetQrFromStorePublicRequestObject) (GetQrFromStorePublicResponseObject, error) {
	req := GetQrFromStoreRequestObject{
		Params: GetQrFromStoreParams{
			Id:     request.Params.Id,
			Issuer: request.Params.Issuer,
		},
	}
	resp, err := s.GetQrFromStore(ctx, req)
	if err != nil {
		return nil, err
	}

	switch response := resp.(type) {
	case GetQrFromStore400JSONResponse:
		return GetQrFromStorePublic400JSONResponse(response), nil
	case GetQrFromStore404JSONResponse:
		return GetQrFromStorePublic404JSONResponse(response), nil
	case GetQrFromStore410JSONResponse:
		return GetQrFromStorePublic410JSONResponse(response), nil
	case GetQrFromStore500JSONResponse:
		return GetQrFromStorePublic500JSONResponse(response), nil
	case GetQrFromStoreResponseObject:
		r, ok := resp.(GetQrFromStorePublicResponseObject)
		if !ok {
			log.Error(ctx, "unexpected response type", "response", response)
			return GetQrFromStorePublic500JSONResponse{N500JSONResponse{"unexpected response type"}}, nil
		}
		return r, nil
	default:
		log.Error(ctx, "unexpected response type", "response", response)
		return GetQrFromStorePublic500JSONResponse{N500JSONResponse{"unexpected response type"}}, nil
	}
}
