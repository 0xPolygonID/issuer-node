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

	_, err := s.identityService.Authenticate(ctx, *request.Body, request.Params.SessionID, s.cfg.ServerUrl)
	if err != nil {
		log.Error(ctx, "error authenticating", err.Error())
		return AuthCallback500JSONResponse{}, nil
	}

	return AuthCallback200Response{}, nil
}

// AuthCallbackPublic receives the authentication information of a holder
func (s *Server) AuthCallbackPublic(ctx context.Context, request AuthCallbackPublicRequestObject) (AuthCallbackPublicResponseObject, error) {
	resp, err := s.AuthCallback(ctx, AuthCallbackRequestObject{
		Body:   request.Body,
		Params: AuthCallbackParams(request.Params),
	})
	if err != nil {
		return nil, err
	}

	switch response := resp.(type) {
	case AuthCallback200Response:
		return AuthCallbackPublic200Response{}, nil
	case AuthCallback400JSONResponse:
		return AuthCallbackPublic400JSONResponse(response), nil
	case AuthCallback500JSONResponse:
		return AuthCallbackPublic500JSONResponse(response), nil
	default:
		return AuthCallbackPublic500JSONResponse{N500JSONResponse{"unexpected response"}}, nil
	}
}

// Authentication returns the qr code for authenticating a user
func (s *Server) Authentication(ctx context.Context, req AuthenticationRequestObject) (AuthenticationResponseObject, error) {
	did, err := w3c.ParseDID(req.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err)
		return Authentication400JSONResponse{N400JSONResponse{"Invalid issuer did"}}, nil
	}
	resp, err := s.identityService.CreateAuthenticationQRCode(ctx, s.cfg.ServerUrl, *did)
	if err != nil {
		log.Error(ctx, "creating qr code", "err", err)
		return Authentication500JSONResponse{N500JSONResponse{"Unexpected error while creating qr code"}}, nil
	}
	if req.Params.Type != nil && *req.Params.Type == AuthenticationParamsTypeRaw {
		body, err := s.qrService.Find(ctx, resp.QrID)
		if err != nil {
			log.Error(ctx, "qr store. Finding qr", "err", err, "QrID", resp.QrID)
			return Authentication500JSONResponse{N500JSONResponse{"error looking for qr body"}}, nil
		}
		return Authentication200JSONResponse{
			Message:   string(body),
			SessionID: resp.SessionID.String(),
		}, nil
	}
	return Authentication200JSONResponse{
		Message:   resp.QRCodeURL,
		SessionID: resp.SessionID.String(),
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
