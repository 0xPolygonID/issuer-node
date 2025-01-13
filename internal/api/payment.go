package api

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/payments"
	"github.com/polygonid/sh-id-platform/internal/repositories"
)

// GetPaymentOptions is the controller to get payment options
func (s *Server) GetPaymentOptions(ctx context.Context, request GetPaymentOptionsRequestObject) (GetPaymentOptionsResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetPaymentOptions400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	paymentOptions, err := s.paymentService.GetPaymentOptions(ctx, issuerDID)
	if err != nil {
		log.Error(ctx, "getting payment options", "err", err, "issuer", issuerDID)
		return GetPaymentOptions500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't get payment-options: <%s>", err.Error())}}, nil
	}
	items, err := toGetPaymentOptionsResponse(paymentOptions)
	if err != nil {
		log.Error(ctx, "creating payment options response", "err", err)
		return GetPaymentOptions500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't convert payment-options: <%s>", err.Error())}}, nil
	}
	return GetPaymentOptions200JSONResponse(
		PaymentOptionsPaginated{
			Items: items,
			Meta: PaginatedMetadata{ // No pagination by now, just return all
				MaxResults: uint(len(paymentOptions)),
				Page:       1,
				Total:      uint(len(paymentOptions)),
			},
		}), nil
}

// CreatePaymentOption is the controller to create a payment option
func (s *Server) CreatePaymentOption(ctx context.Context, request CreatePaymentOptionRequestObject) (CreatePaymentOptionResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreatePaymentOption400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	payOptConf, err := newPaymentOptionConfig(request.Body.PaymentOptions)
	if err != nil {
		log.Error(ctx, "creating payment option config", "err", err)
		return CreatePaymentOption400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("invalid config: %s", err)}}, nil
	}
	id, err := s.paymentService.CreatePaymentOption(ctx, issuerDID, request.Body.Name, request.Body.Description, payOptConf)
	if err != nil {
		log.Error(ctx, "creating payment option", "err", err, "issuer", issuerDID, "request", request.Body)
		if errors.Is(err, repositories.ErrIdentityNotFound) {
			return CreatePaymentOption400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
		}
		if errors.Is(err, repositories.ErrPaymentOptionAlreadyExists) {
			return CreatePaymentOption409JSONResponse{N409JSONResponse{Message: "payment option name already exists"}}, nil
		}
		return CreatePaymentOption500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't create payment-option: <%s>", err.Error())}}, nil
	}
	return CreatePaymentOption201JSONResponse{Id: id.String()}, nil
}

// DeletePaymentOption is the controller to delete a payment option
func (s *Server) DeletePaymentOption(ctx context.Context, request DeletePaymentOptionRequestObject) (DeletePaymentOptionResponseObject, error) {
	issuerID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return DeletePaymentOption400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	if err := s.paymentService.DeletePaymentOption(ctx, issuerID, request.Id); err != nil {
		log.Error(ctx, "deleting payment option", "err", err, "issuer", issuerID, "id", request.Id)
		if errors.Is(err, repositories.ErrPaymentOptionDoesNotExists) {
			return DeletePaymentOption400JSONResponse{N400JSONResponse{Message: "payment option not found"}}, nil
		}
		return DeletePaymentOption500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't delete payment-option: <%s>", err.Error())}}, nil
	}
	return DeletePaymentOption200JSONResponse{Message: "deleted"}, nil
}

// GetPaymentOption is the controller to get a payment option
func (s *Server) GetPaymentOption(ctx context.Context, request GetPaymentOptionRequestObject) (GetPaymentOptionResponseObject, error) {
	issuerID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetPaymentOption400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	paymentOption, err := s.paymentService.GetPaymentOptionByID(ctx, issuerID, request.Id)
	if err != nil {
		if errors.Is(err, repositories.ErrPaymentOptionDoesNotExists) {
			return GetPaymentOption404JSONResponse{N404JSONResponse{Message: "payment option not found"}}, nil
		}
		log.Error(ctx, "getting payment option", "err", err, "issuer", issuerID, "id", request.Id)
		return GetPaymentOption500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't get payment-option: <%s>", err.Error())}}, nil
	}

	option, err := toPaymentOption(paymentOption)
	if err != nil {
		log.Error(ctx, "creating payment option response", "err", err)
		return GetPaymentOption500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't convert payment-option: <%s>", err.Error())}}, nil
	}
	return GetPaymentOption200JSONResponse(option), nil
}

// CreatePaymentRequest is the controller to create payment request
func (s *Server) CreatePaymentRequest(ctx context.Context, request CreatePaymentRequestRequestObject) (CreatePaymentRequestResponseObject, error) {
	var err error
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	userDID, err := w3c.ParseDID(request.Body.UserDID)
	if err != nil {
		log.Error(ctx, "parsing user did", "err", err, "did", request.Body.UserDID)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid userDID"}}, nil
	}

	req := &ports.CreatePaymentRequestReq{
		IssuerDID:   *issuerDID,
		UserDID:     *userDID,
		SchemaID:    request.Body.SchemaID,
		OptionID:    request.Body.OptionID,
		Description: request.Body.Description,
	}

	payReq, err := s.paymentService.CreatePaymentRequest(ctx, req)
	if err != nil {
		log.Error(ctx, "creating payment request", "err", err)
		return CreatePaymentRequest400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("can't create payment-request: %s", err)}}, nil
	}
	return CreatePaymentRequest201JSONResponse(toCreatePaymentRequestResponse(payReq)), nil
}

// GetPaymentRequests is the controller to get all payment requests for an issuer
func (s *Server) GetPaymentRequests(ctx context.Context, request GetPaymentRequestsRequestObject) (GetPaymentRequestsResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetPaymentRequests400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	paymentRequests, err := s.paymentService.GetPaymentRequests(ctx, issuerDID)
	if err != nil {
		return GetPaymentRequests500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't get payment-requests: %s", err)}}, nil
	}
	return GetPaymentRequests200JSONResponse(toGetPaymentRequestsResponse(paymentRequests)), nil
}

// GetPaymentRequest is the controller to get payment request by ID
func (s *Server) GetPaymentRequest(ctx context.Context, request GetPaymentRequestRequestObject) (GetPaymentRequestResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return GetPaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	paymentRequest, err := s.paymentService.GetPaymentRequest(ctx, issuerDID, request.Id)
	if err != nil {
		return GetPaymentRequest500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't get payment-request: %s", err)}}, nil
	}
	return GetPaymentRequest200JSONResponse(toCreatePaymentRequestResponse(paymentRequest)), nil
}

// DeletePaymentRequest is the controller to delete payment request
func (s *Server) DeletePaymentRequest(ctx context.Context, request DeletePaymentRequestRequestObject) (DeletePaymentRequestResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return DeletePaymentRequest400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	err = s.paymentService.DeletePaymentRequest(ctx, issuerDID, request.Id)
	if err != nil {
		return DeletePaymentRequest500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't delete payment-request: %s", err)}}, nil
	}
	return DeletePaymentRequest200JSONResponse{Message: "deleted"}, nil
}

// GetPaymentSettings is the controller to get payment settings
func (s *Server) GetPaymentSettings(_ context.Context, _ GetPaymentSettingsRequestObject) (GetPaymentSettingsResponseObject, error) {
	return GetPaymentSettings200JSONResponse(s.paymentService.GetSettings()), nil
}

// VerifyPayment is the controller to verify payment
func (s *Server) VerifyPayment(ctx context.Context, request VerifyPaymentRequestObject) (VerifyPaymentResponseObject, error) {
	const base10 = 10
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return VerifyPayment400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}
	nonce, ok := new(big.Int).SetString(request.Nonce, base10)
	if !ok {
		log.Error(ctx, "parsing nonce on verify payment", "err", err, "nonce", request.Nonce)
		return VerifyPayment400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("invalid nonce: <%s>", request.Nonce)}}, nil
	}

	var userDID *w3c.DID
	var txHash *string

	if request.Body != nil {
		if request.Body.TxHash != nil && *request.Body.TxHash != "" {
			txHash = request.Body.TxHash
		}
		if request.Body.UserDID != nil && *request.Body.UserDID != "" {
			userDID, err = w3c.ParseDID(*request.Body.UserDID)
			if err != nil {
				log.Error(ctx, "parsing user did", "err", err, "did", request.Body.UserDID)
				return VerifyPayment400JSONResponse{N400JSONResponse{Message: "invalid user did"}}, nil
			}
		}
	}
	status, err := s.paymentService.VerifyPayment(ctx, *issuerDID, nonce, txHash, userDID)
	if err != nil {
		log.Error(ctx, "can't verify payment", "err", err, "nonce", request.Nonce, "txID", txHash)
		return VerifyPayment400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("can't verify payment: %s", err.Error())}}, nil
	}
	return toVerifyPaymentResponse(status)
}

// UpdatePaymentOption - updates a payment option
func (s *Server) UpdatePaymentOption(ctx context.Context, request UpdatePaymentOptionRequestObject) (UpdatePaymentOptionResponseObject, error) {
	issuerDID, err := w3c.ParseDID(request.Identifier)
	if err != nil {
		log.Error(ctx, "parsing issuer did", "err", err, "did", request.Identifier)
		return UpdatePaymentOption400JSONResponse{N400JSONResponse{Message: "invalid issuer did"}}, nil
	}

	var payOptConf *domain.PaymentOptionConfig
	if request.Body.PaymentOptions != nil {
		payOptConf, err = newPaymentOptionConfig(*request.Body.PaymentOptions)
		if err != nil {
			log.Error(ctx, "creating payment option config", "err", err)
			return UpdatePaymentOption400JSONResponse{N400JSONResponse{Message: fmt.Sprintf("invalid config: %s", err)}}, nil
		}
	}

	err = s.paymentService.UpdatePaymentOption(ctx, issuerDID, request.Id, request.Body.Name, request.Body.Description, payOptConf)
	if err != nil {
		log.Error(ctx, "updating payment option", "err", err, "issuer", issuerDID, "request", request.Body)
		if errors.Is(err, repositories.ErrPaymentOptionDoesNotExists) {
			return UpdatePaymentOption400JSONResponse{N400JSONResponse{Message: "payment option not found"}}, nil
		}
		return UpdatePaymentOption500JSONResponse{N500JSONResponse{Message: fmt.Sprintf("can't update payment-option: <%s>", err.Error())}}, nil
	}
	return UpdatePaymentOption200JSONResponse{}, nil
}

func newPaymentOptionConfig(config PaymentOptionConfig) (*domain.PaymentOptionConfig, error) {
	const base10 = 10
	cfg := &domain.PaymentOptionConfig{
		PaymentOptions: make([]domain.PaymentOptionConfigItem, len(config)),
	}
	for i, item := range config {
		if !common.IsHexAddress(item.Recipient) {
			return nil, fmt.Errorf("invalid recipient address: %s", item.Recipient)
		}
		amount, ok := new(big.Int).SetString(item.Amount, base10)
		if !ok {
			return nil, fmt.Errorf("could not parse amount: %s", item.Amount)
		}

		cfg.PaymentOptions[i] = domain.PaymentOptionConfigItem{
			PaymentOptionID: payments.OptionConfigIDType(item.PaymentOptionID),
			Amount:          *amount,
			Recipient:       common.HexToAddress(item.Recipient),
			SigningKeyID:    item.SigningKeyID,
			Expiration:      item.Expiration,
		}
	}
	return cfg, nil
}
