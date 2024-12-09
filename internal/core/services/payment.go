package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/eth"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

type payment struct {
	networkResolver network.Resolver
	settings        payments.Config
	paymentsStore   ports.PaymentRepository
	kms             kms.KMSType
}

// NewPaymentService creates a new payment service
func NewPaymentService(payOptsRepo ports.PaymentRepository, resolver network.Resolver, settings *payments.Config, kms kms.KMSType) ports.PaymentService {
	return &payment{
		networkResolver: resolver,
		settings:        *settings,
		paymentsStore:   payOptsRepo,
		kms:             kms,
	}
}

// CreatePaymentOption creates a payment option for a specific issuer
func (p *payment) CreatePaymentOption(ctx context.Context, issuerDID *w3c.DID, name, description string, config *domain.PaymentOptionConfig) (uuid.UUID, error) {
	paymentOption := domain.NewPaymentOption(*issuerDID, name, description, config)
	id, err := p.paymentsStore.SavePaymentOption(ctx, paymentOption)
	if err != nil {
		log.Error(ctx, "failed to save payment option", "err", err, "issuerDID", issuerDID, "name", name, "description", description, "config", config)
		return uuid.Nil, err
	}
	return id, nil
}

// GetPaymentOptions returns all payment options of a issuer
func (p *payment) GetPaymentOptions(ctx context.Context, issuerDID *w3c.DID) ([]domain.PaymentOption, error) {
	opts, err := p.paymentsStore.GetAllPaymentOptions(ctx, *issuerDID)
	if err != nil {
		log.Error(ctx, "failed to get payment options", "err", err, "issuerDID", issuerDID)
		return nil, err
	}
	return opts, nil
}

// GetPaymentOptionByID returns a payment option by its ID
func (p *payment) GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error) {
	opt, err := p.paymentsStore.GetPaymentOptionByID(ctx, issuerDID, id)
	if err != nil {
		log.Error(ctx, "failed to get payment option", "err", err, "issuerDID", issuerDID, "id", id)
		return nil, err
	}
	return opt, nil
}

// DeletePaymentOption deletes a payment option
func (p *payment) DeletePaymentOption(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) error {
	err := p.paymentsStore.DeletePaymentOption(ctx, *issuerDID, id)
	if err != nil {
		log.Error(ctx, "failed to delete payment option", "err", err, "issuerDID", issuerDID, "id", id)
		return err
	}
	return nil
}

// CreatePaymentRequest creates a payment request
func (p *payment) CreatePaymentRequest(ctx context.Context, req *ports.CreatePaymentRequestReq) (*domain.PaymentRequest, error) {
	option, err := p.paymentsStore.GetPaymentOptionByID(ctx, &req.IssuerDID, req.OptionID)
	if err != nil {
		log.Error(ctx, "failed to get payment option", "err", err, "issuerDID", req.IssuerDID, "optionID", req.OptionID)
		return nil, err
	}

	paymentRequest := &domain.PaymentRequest{
		ID:              uuid.New(),
		IssuerDID:       req.IssuerDID,
		RecipientDID:    req.UserDID,
		Credentials:     req.Credentials,
		Description:     req.Description,
		PaymentOptionID: req.OptionID,
		CreatedAt:       time.Now(),
	}
	for _, chainConfig := range option.Config.Config {
		setting, found := p.settings[chainConfig.PaymentOptionID]
		if !found {
			log.Error(ctx, "chain not found in configuration", "paymentOptionID", chainConfig.PaymentOptionID)
			return nil, fmt.Errorf("payment Option <%d> not found in payment configuration", chainConfig.PaymentOptionID)
		}

		nonce, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(130), nil)) //nolint: mnd
		if err != nil {
			log.Error(ctx, "failed to generate nonce", "err", err)
			return nil, err
		}

		data, err := p.paymentInfo(ctx, setting, &chainConfig, nonce)
		if err != nil {
			log.Error(ctx, "failed to create payment info", "err", err)
			return nil, err
		}
		item := domain.PaymentRequestItem{
			ID:               uuid.New(),
			Nonce:            *nonce,
			PaymentRequestID: paymentRequest.ID,

			Payment: data,
		}
		paymentRequest.Payments = append(paymentRequest.Payments, item)
	}

	_, err = p.paymentsStore.SavePaymentRequest(ctx, paymentRequest)
	if err != nil {
		log.Error(ctx, "failed to save payment request", "err", err, "paymentRequest", paymentRequest)
		return nil, fmt.Errorf("failed to save payment request: %w", err)
	}

	return paymentRequest, nil
}

// CreatePaymentRequestForProposalRequest creates a payment request for a proposal request
func (p *payment) CreatePaymentRequestForProposalRequest(_ context.Context, proposalRequest *protocol.CredentialsProposalRequestMessage) (*comm.BasicMessage, error) {
	basicMessage := comm.BasicMessage{
		From:     proposalRequest.To,
		To:       proposalRequest.From,
		ThreadID: proposalRequest.ThreadID,
		ID:       proposalRequest.ID,
		Typ:      proposalRequest.Typ,
	}
	return &basicMessage, nil
}

// GetSettings returns the current payment settings
func (p *payment) GetSettings() payments.Config {
	return p.settings
}

// VerifyPayment verifies a payment
// TODO: Total refactor! Reimplement from scratch!!!!
func (p *payment) VerifyPayment(ctx context.Context, paymentOptionID uuid.UUID, message *protocol.PaymentMessage) (bool, error) {
	if len(message.Body.Payments) != 1 {
		return false, fmt.Errorf("expected one payment, got %d", len(message.Body.Payments))
	}

	option, err := p.paymentsStore.GetPaymentOptionByID(ctx, nil, paymentOptionID)
	if err != nil {
		return false, fmt.Errorf("failed to get payment option: %w", err)
	}

	// TODO: Load rpc from network resolvers
	client, err := ethclient.Dial("https://polygon-amoy.g.alchemy.com/v2/DHvucvBBzrBhaHzmjrMp24PGbl7vwee6")
	if err != nil {
		return false, fmt.Errorf("failed to connect to ethereum client: %w", err)
	}

	// contractAddress := common.HexToAddress("0xF8E49b922D5Fb00d3EdD12bd14064f275726D339")
	contractAddress, err := contractAddressFromPayment(&message.Body.Payments[0], p.settings)
	if err != nil {
		return false, fmt.Errorf("failed to get contract address from payment: %w", err)
	}
	instance, err := eth.NewPaymentContract(*contractAddress, client)
	if err != nil {
		return false, err
	}

	// TODO: Iterate over all payments? Right now we only support one payment
	nonce, err := nonceFromPayment(&message.Body.Payments[0])
	if err != nil {
		log.Error(ctx, "failed to get nonce from payment request info data", "err", err)
		return false, err
	}

	recipientAddr, err := recipientAddressFromPayment(&message.Body.Payments[0], option)
	if err != nil {
		log.Error(ctx, "failed to get recipient address from payment", "err", err)
		return false, err
	}

	// TODO: pending, canceled, success, failed
	isPaid, err := instance.IsPaymentDone(&bind.CallOpts{Context: ctx}, *recipientAddr, nonce)
	if err != nil {
		return false, err
	}
	return isPaid, nil
}

func contractAddressFromPayment(data *protocol.Payment, config payments.Config) (*common.Address, error) {
	/*
		var sChainID string
		switch data.Type() {
		case protocol.Iden3PaymentCryptoV1Type:
			return nil, nil
		case protocol.Iden3PaymentRailsV1Type:
			d := data.Data()
			t, ok := d.(*protocol.Iden3PaymentRailsV1)
			if !ok {
				return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsRequestV1")
			}
			sChainID = t.PaymentData.ChainID
		case protocol.Iden3PaymentRailsERC20V1Type:
			t, ok := data.Data().(*protocol.Iden3PaymentRailsERC20V1)
			if !ok {
				return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsERC20RequestV1")
			}
			sChainID = t.PaymentData.ChainID
		default:
			return nil, fmt.Errorf("unsupported payment request data type: %s", data.Type())
		}
		chainID, err := strconv.Atoi(sChainID)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chain id: %w", err)
		}
		c, found := config[chainID]
		if !found {
			return nil, fmt.Errorf("chain id not found in settings: %d", chainID)
		}
		addr := common.HexToAddress(c.MCPayment)
		return &addr, nil

	*/
	return &common.Address{}, nil
}

func recipientAddressFromPayment(data *protocol.Payment, option *domain.PaymentOption) (*common.Address, error) {
	/*
		var address common.Address
		switch data.Type() {
		case protocol.Iden3PaymentCryptoV1Type:
			address = common.Address{}
		case protocol.Iden3PaymentRailsV1Type:
			t, ok := data.Data().(*protocol.Iden3PaymentRailsV1)
			if !ok {
				return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsRequestV1")
			}
			for _, chain := range option.Config.Chains {
				if strconv.Itoa(chain.ChainId) == t.PaymentData.ChainID {
					address = common.HexToAddress(chain.Recipient)
					break
				}
			}
		case protocol.Iden3PaymentRailsERC20V1Type:
			t, ok := data.Data().(*protocol.Iden3PaymentRailsERC20V1)
			if !ok {
				return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsERC20RequestV1")
			}
			for _, chain := range option.Config.Chains {
				if strconv.Itoa(chain.ChainId) == t.PaymentData.ChainID {
					address = common.HexToAddress(chain.Recipient)
					break
				}
			}
		default:
			return nil, fmt.Errorf("unsupported payment request data type: %s", data.Type())
		}
		return &address, nil

	*/
	return &common.Address{}, nil
}

func nonceFromPayment(data *protocol.Payment) (*big.Int, error) {
	const base10 = 10
	var nonce string
	switch data.Type() {
	case protocol.Iden3PaymentCryptoV1Type:
		nonce = ""
	case protocol.Iden3PaymentRailsV1Type:
		t, ok := data.Data().(*protocol.Iden3PaymentRailsV1)
		if !ok {
			return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsRequestV1")
		}
		nonce = t.Nonce
	case protocol.Iden3PaymentRailsERC20V1Type:
		t, ok := data.Data().(*protocol.Iden3PaymentRailsERC20V1)
		if !ok {
			return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsERC20RequestV1")
		}
		nonce = t.Nonce
	default:
		return nil, fmt.Errorf("unsupported payment request data type: %s", data.Type())
	}
	bigIntNonce, ok := new(big.Int).SetString(nonce, base10)
	if !ok {
		return nil, fmt.Errorf("failed to parse nonce creating big int: %s", nonce)
	}
	return bigIntNonce, nil
}

func (p *payment) paymentInfo(ctx context.Context, setting payments.ChainConfig, chainConfig *domain.PaymentOptionConfigItem, nonce *big.Int) (protocol.PaymentRequestInfoDataItem, error) {
	const defaultExpirationDate = 1 * time.Hour
	expirationTime := time.Now().Add(defaultExpirationDate)

	metadata := "0x"
	signature, err := p.paymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
	if err != nil {
		log.Error(ctx, "failed to create payment request signature", "err", err)
		return nil, err
	}

	switch setting.PaymentOption.Type {
	case protocol.Iden3PaymentRailsRequestV1Type:
		return &protocol.Iden3PaymentRailsRequestV1{
			Nonce: nonce.String(),
			Type:  protocol.Iden3PaymentRailsRequestV1Type,
			Context: protocol.NewPaymentContextString(
				"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsRequestV1",
				"https://w3id.org/security/suites/eip712sig-2021/v1",
			),
			Amount:         chainConfig.Amount.String(),
			ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
			Metadata:       metadata,
			Recipient:      chainConfig.Recipient.String(),
			Proof:          paymentProof(&setting, signature),
		}, nil

	case protocol.Iden3PaymentRailsERC20RequestV1Type:
		return &protocol.Iden3PaymentRailsERC20RequestV1{
			Nonce: nonce.String(),
			Type:  protocol.Iden3PaymentRailsERC20RequestV1Type,
			Context: protocol.NewPaymentContextString(
				"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
				"https://w3id.org/security/suites/eip712sig-2021/v1",
			),
			Amount:         chainConfig.Amount.String(),
			ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
			Metadata:       metadata,
			Recipient:      chainConfig.Recipient.String(),
			Features:       setting.PaymentOption.Features,
			TokenAddress:   setting.PaymentOption.ContractAddress.String(),
			Proof:          paymentProof(&setting, signature),
		}, nil

	case protocol.Iden3PaymentRequestCryptoV1Type:
		return &protocol.Iden3PaymentRequestCryptoV1{}, nil
	default:
		return nil, fmt.Errorf("unsupported payment option type: %s", setting.PaymentOption.Type)
	}
}

func paymentProof(setting *payments.ChainConfig, signature []byte) protocol.PaymentProof {
	var eip712DataTypes string
	if setting.PaymentOption.Type == protocol.Iden3PaymentRailsRequestV1Type {
		eip712DataTypes = "https://schema.iden3.io/core/json/Iden3PaymentRailsRequestV1.json"
	}
	if setting.PaymentOption.Type == protocol.Iden3PaymentRailsERC20RequestV1Type {
		eip712DataTypes = "https://schema.iden3.io/core/json/Iden3PaymentRailsERC20RequestV1.json"
	}
	return protocol.PaymentProof{
		protocol.EthereumEip712Signature2021{
			Type:               "EthereumEip712Signature2021",
			ProofPurpose:       "assertionMethod",
			ProofValue:         fmt.Sprintf("0x%s", hex.EncodeToString(signature)),
			VerificationMethod: fmt.Sprintf("did:pkh:eip155:%d:%s", setting.ChainID, setting.PaymentRails),
			Created:            time.Now().Format(time.RFC3339),
			Eip712: protocol.Eip712Data{
				Types:       eip712DataTypes,
				PrimaryType: string(setting.PaymentOption.Type),
				Domain: protocol.Eip712Domain{
					Name:              "MCPayment",
					Version:           "1.0.0",
					ChainID:           strconv.Itoa(setting.ChainID),
					VerifyingContract: setting.PaymentRails.String(),
				},
			},
		},
	}
}

func (p *payment) paymentRequestSignature(
	ctx context.Context,
	setting payments.ChainConfig,
	chainConfig *domain.PaymentOptionConfigItem,
	expTime time.Time,
	nonce *big.Int,
	metadata string,
) ([]byte, error) {
	paymentType := string(setting.PaymentOption.Type)

	keyID := kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   chainConfig.SigningKeyID,
	}

	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			paymentType: []apitypes.Type{
				{
					Name: "recipient",
					Type: "address",
				},
				{
					Name: "amount",
					Type: "uint256",
				},
				{
					Name: "expirationDate",
					Type: "uint256",
				},
				{
					Name: "nonce",
					Type: "uint256",
				},
				{
					Name: "metadata",
					Type: "bytes",
				},
			},
		},
		PrimaryType: paymentType,
		Domain: apitypes.TypedDataDomain{
			Name:              "MCPayment",
			Version:           "1.0.0",
			ChainId:           math.NewHexOrDecimal256(int64(setting.ChainID)),
			VerifyingContract: setting.PaymentRails.String(),
		},
		Message: apitypes.TypedDataMessage{
			"recipient":      chainConfig.Recipient.String(),
			"amount":         chainConfig.Amount.String(),
			"expirationDate": big.NewInt(expTime.Unix()),
			"nonce":          nonce,
			"metadata":       metadata,
		},
	}
	if paymentType == string(protocol.Iden3PaymentRailsERC20RequestV1Type) {
		typedData.Types[paymentType] = []apitypes.Type{
			{
				Name: "tokenAddress",
				Type: "address",
			},
			{
				Name: "recipient",
				Type: "address",
			},
			{
				Name: "amount",
				Type: "uint256",
			},
			{
				Name: "expirationDate",
				Type: "uint256",
			},
			{
				Name: "nonce",
				Type: "uint256",
			},
			{
				Name: "metadata",
				Type: "bytes",
			},
		}
		typedData.Message["tokenAddress"] = setting.PaymentOption.ContractAddress.String()
	}
	typedDataBytes, _, err := apitypes.TypedDataAndHash(typedData)
	if err != nil {
		return nil, err
	}

	signature, err := p.kms.Sign(ctx, keyID, typedDataBytes)
	if err != nil {
		log.Error(ctx, "failed to sign typed data hash", "err", err, "keyId", keyID)
		return nil, err
	}
	return signature, nil
}
