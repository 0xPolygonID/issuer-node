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

		link, err := s.linkService.GetByID(ctx, *issuerDID, *request.Params.Id)

		if link.AuthorizationRequestMessage == nil {
			log.Error(ctx, "qr store. Finding qr", "err", err, "id", *request.Params.Id)
			return GetQrFromStore400JSONResponse{N400JSONResponse{"error looking for qr body"}}, nil
		}

		return NewQrContentResponse(link.AuthorizationRequestMessage.Bytes), nil
	}
	return NewQrContentResponse(body), nil
}
