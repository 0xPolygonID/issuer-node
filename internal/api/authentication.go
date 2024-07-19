package api

import (
	"context"
	"errors"

	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/services"
	"github.com/polygonid/sh-id-platform/internal/log"
)

// AuthCallback receives the authentication information of a holder
func (s *Server) AuthCallback(ctx context.Context, request AuthCallbackRequestObject) (AuthCallbackResponseObject, error) {
	if request.Body == nil || *request.Body == "" {
		log.Debug(ctx, "empty request body auth-callback request")
		return AuthCallback400JSONResponse{N400JSONResponse{"Cannot proceed with empty body"}}, nil
	}

	_, err := s.identityService.Authenticate(ctx, *request.Body, request.Params.SessionID, s.cfg.APIUI.ServerURL, s.cfg.APIUI.IssuerDID)
	if err != nil {
		log.Error(ctx, "error authenticating", err.Error())
		return AuthCallback500JSONResponse{}, nil
	}

	return AuthCallback200Response{}, nil
}

// AuthQRCode returns the qr code for authenticating a user
func (s *Server) AuthQRCode(ctx context.Context, req AuthQRCodeRequestObject) (AuthQRCodeResponseObject, error) {
	did, err := w3c.ParseDID(req.IssuerDID)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err)
		return AuthQRCode400JSONResponse{N400JSONResponse{"Invalid issuer did"}}, nil
	}
	resp, err := s.identityService.CreateAuthenticationQRCode(ctx, s.cfg.ServerUrl, *did)
	if err != nil {
		return AuthQRCode500JSONResponse{N500JSONResponse{"Unexpected error while creating qr code"}}, nil
	}
	if req.Params.Type != nil && *req.Params.Type == Raw {
		body, err := s.qrService.Find(ctx, resp.QrID)
		if err != nil {
			log.Error(ctx, "qr store. Finding qr", "err", err, "QrID", resp.QrID)
			return AuthQRCode500JSONResponse{N500JSONResponse{"error looking for qr body"}}, nil
		}
		return AuthQRCode200JSONResponse{
			QrCodeLink: string(body),
			SessionID:  resp.SessionID.String(),
		}, nil
	}
	return AuthQRCode200JSONResponse{
		QrCodeLink: resp.QRCodeURL,
		SessionID:  resp.SessionID.String(),
	}, nil
}

// GetAuthenticationConnection returns the connection related to a given session
func (s *Server) GetAuthenticationConnection(ctx context.Context, req GetAuthenticationConnectionRequestObject) (GetAuthenticationConnectionResponseObject, error) {
	conn, err := s.connectionsService.GetByUserSessionID(ctx, req.Id)
	if err != nil {
		log.Error(ctx, "get authentication connection", "err", err, "req", req)
		if errors.Is(err, services.ErrConnectionDoesNotExist) {
			return GetAuthenticationConnection404JSONResponse{N404JSONResponse{err.Error()}}, nil
		}
		return GetAuthenticationConnection500JSONResponse{N500JSONResponse{"Unexpected error while getting authentication session"}}, nil
	}

	return GetAuthenticationConnection200JSONResponse{
		Connection: AuthenticationConnection{
			Id:         conn.ID.String(),
			UserID:     conn.UserDID.String(),
			IssuerID:   conn.IssuerDID.String(),
			CreatedAt:  TimeUTC(conn.CreatedAt),
			ModifiedAt: TimeUTC(conn.ModifiedAt),
		},
	}, nil
}
