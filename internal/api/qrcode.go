package api

import (
	"context"

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
		return GetQrFromStore500JSONResponse{N500JSONResponse{"error looking for qr body"}}, nil
	}
	return NewQrContentResponse(body), nil
}
