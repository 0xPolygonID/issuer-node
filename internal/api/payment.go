package api

import (
	"context"
	"encoding/json"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"

	"github.com/polygonid/sh-id-platform/internal/log"
)

// CreatePaymentRequest is the controller to get qr bodies
func (s *Server) CreatePaymentRequest(ctx context.Context, request CreatePaymentRequestRequestObject) (CreatePaymentRequestResponseObject, error) {
	var err error
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	userDID, err := w3c.ParseDID(request.Body.UserDID)
	if err != nil {
		log.Error(ctx, "parsing user did", "err", err, "did", request.Identifier)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	paymentRequest, err := s.paymentService.CreatePaymentRequest(ctx, issuerDID, userDID, request.Body.SigningKey, request.Body.CredentialContext, request.Body.CredentialType)
	if err != nil {
		log.Error(ctx, "creating payment request", "err", err)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: "can't create payment-request"}}, nil
	}

	basicMessage := BasicMessage{
		From:     paymentRequest.From,
		To:       paymentRequest.To,
		ThreadID: paymentRequest.ThreadID,
		Id:       paymentRequest.ID,
		Typ:      string(paymentRequest.Typ),
		Type:     string(paymentRequest.Type),
		Body:     paymentRequest.Body,
	}
	return CreatePaymentRequest201JSONResponse(basicMessage), nil
}

// GetPaymentSettings is the controller to get payment settings
func (s *Server) GetPaymentSettings(_ context.Context, _ GetPaymentSettingsRequestObject) (GetPaymentSettingsResponseObject, error) {
	return GetPaymentSettings200JSONResponse(s.paymentService.GetSettings()), nil
}

// VerifyPayment is the controller to verify payment
func (s *Server) VerifyPayment(ctx context.Context, request VerifyPaymentRequestObject) (VerifyPaymentResponseObject, error) {
	bodyBytes, err := json.Marshal(request.Body.Body)
	if err != nil {
		log.Error(ctx, "marshaling request body", "err", err)
		return VerifyPayment400JSONResponse{N400JSONResponse{Message: "can't parse payment"}}, nil
	}
	basicMessage := comm.BasicMessage{
		From:     request.Body.From,
		To:       request.Body.To,
		ThreadID: request.Body.ThreadID,
		ID:       request.Body.Id,
		Typ:      comm.MediaType(request.Body.Typ),
		Type:     comm.ProtocolMessage(""),
		Body:     bodyBytes,
	}

	recipient := common.HexToAddress(request.Recipient)

	isPaid, err := s.paymentService.VerifyPayment(ctx, recipient, basicMessage)
	if err != nil {
		log.Error(ctx, "can't process payment message", "err", err)
		return VerifyPayment400JSONResponse{N400JSONResponse{Message: "can't process payment message"}}, nil
	}
	return VerifyPayment200JSONResponse(isPaid), nil
}
