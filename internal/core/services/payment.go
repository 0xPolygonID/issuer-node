package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	core "github.com/iden3/go-iden3-core/v2"
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
	schemaService   ports.SchemaService
	paymentsStore   ports.PaymentRepository
	kms             kms.KMSType
}

// NewPaymentService creates a new payment service
func NewPaymentService(payOptsRepo ports.PaymentRepository, resolver network.Resolver, schemaSrv ports.SchemaService, settings *payments.Config, kms kms.KMSType) ports.PaymentService {
	return &payment{
		networkResolver: resolver,
		settings:        *settings,
		schemaService:   schemaSrv,
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
		return nil, fmt.Errorf("failed to get payment option: %w", err)
	}
	schema, err := p.schemaService.GetByID(ctx, req.IssuerDID, req.SchemaID)
	if err != nil {
		log.Error(ctx, "failed to get schema", "err", err, "issuerDID", req.IssuerDID, "schemaID", req.SchemaID)
		return nil, fmt.Errorf("failed to get schema: %w", err)
	}

	paymentRequest := &domain.PaymentRequest{
		ID:        uuid.New(),
		IssuerDID: req.IssuerDID,
		UserDID:   req.UserDID,
		Credentials: []protocol.PaymentRequestInfoCredentials{
			{
				Context: schema.ContextURL,
				Type:    schema.Type,
			},
		},
		Description:     req.Description,
		PaymentOptionID: req.OptionID,
		CreatedAt:       time.Now(),
	}
	for _, chainConfig := range option.Config.PaymentOptions {
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
			PaymentOptionID:  chainConfig.PaymentOptionID,
			SigningKeyID:     chainConfig.SigningKeyID,
			Payment:          data,
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

// GetPaymentRequests returns all payment requests of a issuer
func (p *payment) GetPaymentRequests(ctx context.Context, issuerDID *w3c.DID) ([]domain.PaymentRequest, error) {
	paymentRequests, err := p.paymentsStore.GetAllPaymentRequests(ctx, *issuerDID)
	if err != nil {
		log.Error(ctx, "failed to get payment requests", "err", err, "issuerDID", issuerDID)
		return nil, fmt.Errorf("failed to get payment requests: %w", err)
	}
	return paymentRequests, nil
}

// GetPaymentRequest returns payment request by ID and issuer DID
func (p *payment) GetPaymentRequest(ctx context.Context, issuerDID *w3c.DID, ID uuid.UUID) (*domain.PaymentRequest, error) {
	paymentRequests, err := p.paymentsStore.GetPaymentRequestByID(ctx, *issuerDID, ID)
	if err != nil {
		log.Error(ctx, "failed to get payment request", "err", err, "issuerDID", issuerDID)
		return nil, fmt.Errorf("failed to get payment request: %w", err)
	}
	return paymentRequests, nil
}

// DeletePaymentRequest deletes a payment request
func (p *payment) DeletePaymentRequest(ctx context.Context, issuerDID *w3c.DID, ID uuid.UUID) error {
	err := p.paymentsStore.DeletePaymentRequest(ctx, *issuerDID, ID)
	if err != nil {
		log.Error(ctx, "failed to delete payment request", "err", err, "issuerDID", issuerDID)
		return fmt.Errorf("failed to delete payment request: %w", err)
	}
	return nil
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
func (p *payment) VerifyPayment(ctx context.Context, issuerDID w3c.DID, nonce *big.Int, txHash *string, userDID *w3c.DID) (ports.BlockchainPaymentStatus, error) {
	paymentReqItem, err := p.paymentsStore.GetPaymentRequestItem(ctx, issuerDID, nonce)
	if err != nil {
		return ports.BlockchainPaymentStatusPending, fmt.Errorf("failed to get payment request: %w", err)
	}

	if userDID != nil {
		paymentReq, err := p.paymentsStore.GetPaymentRequestByID(ctx, issuerDID, paymentReqItem.PaymentRequestID)
		if err != nil {
			return ports.BlockchainPaymentStatusPending, fmt.Errorf("failed to get payment request: %w", err)
		}

		if userDID.String() != paymentReq.UserDID.String() {
			return ports.BlockchainPaymentStatusFailed, fmt.Errorf("userDID %s does not match to User DID %s in payment-request", userDID, paymentReq.UserDID)
		}
	}

	setting, found := p.settings[paymentReqItem.PaymentOptionID]
	if !found {
		log.Error(ctx, "chain not found in configuration", "paymentOptionID", paymentReqItem.PaymentOptionID)
		return ports.BlockchainPaymentStatusPending, fmt.Errorf("payment Option <%d> not found in payment configuration", paymentReqItem.PaymentOptionID)
	}

	client, err := p.networkResolver.GetEthClientByChainID(core.ChainID(setting.ChainID))
	if err != nil {
		log.Error(ctx, "failed to get ethereum client from resolvers", "err", err, "key", paymentReqItem.SigningKeyID)
		return ports.BlockchainPaymentStatusPending, fmt.Errorf("failed to get ethereum client from resolvers settings for key <%s>", paymentReqItem.SigningKeyID)
	}

	instance, err := eth.NewPaymentContract(setting.PaymentRails, client.GetEthereumClient())
	if err != nil {
		return ports.BlockchainPaymentStatusPending, err
	}

	signerAddress, err := p.getSignerAddress(ctx, paymentReqItem.SigningKeyID)
	if err != nil {
		return ports.BlockchainPaymentStatusPending, err
	}

	status, err := p.verifyPaymentOnBlockchain(ctx, client, instance, signerAddress, nonce, txHash)
	if err != nil {
		log.Error(ctx, "failed to verify payment on blockchain", "err", err, "txHash", txHash, "nonce", nonce)
		return ports.BlockchainPaymentStatusPending, err
	}
	return status, nil
}

func (p *payment) verifyPaymentOnBlockchain(
	ctx context.Context,
	client *eth.Client,
	contract *eth.PaymentContract,
	signerAddress common.Address,
	nonce *big.Int,
	txID *string,
) (ports.BlockchainPaymentStatus, error) {
	txIdProvided := txID != nil && *txID != ""

	if txIdProvided {
		status, err := handlePaymentTransaction(ctx, client, *txID)
		if err != nil || status != ports.BlockchainPaymentStatusSuccess {
			return status, err
		}
	}

	isPaid, err := contract.IsPaymentDone(&bind.CallOpts{Context: ctx}, signerAddress, nonce)
	if err != nil {
		return ports.BlockchainPaymentStatusPending, nil
	}

	if isPaid {
		return ports.BlockchainPaymentStatusSuccess, nil
	}

	return ports.BlockchainPaymentStatusFailed, nil
}

func handlePaymentTransaction(
	ctx context.Context,
	client *eth.Client,
	txID string,
) (ports.BlockchainPaymentStatus, error) {
	_, isPending, err := client.GetTransactionByID(ctx, txID)
	if err != nil {
		if err.Error() == "not found" {
			return ports.BlockchainPaymentStatusCancelled, nil
		}
		return ports.BlockchainPaymentStatusUnknown, err
	}

	if isPending {
		return ports.BlockchainPaymentStatusPending, nil
	}

	receipt, err := client.GetTransactionReceiptByID(ctx, txID)
	if err != nil {
		return ports.BlockchainPaymentStatusUnknown, err
	}

	if receipt.Status == 1 {
		return ports.BlockchainPaymentStatusSuccess, nil
	}

	return ports.BlockchainPaymentStatusFailed, nil
}

func (p *payment) paymentInfo(ctx context.Context, setting payments.ChainConfig, chainConfig *domain.PaymentOptionConfigItem, nonce *big.Int) (protocol.PaymentRequestInfoDataItem, error) {
	const defaultExpirationDate = 1 * time.Hour
	expirationTime := time.Now().Add(defaultExpirationDate)

	if chainConfig.Expiration != nil {
		expirationTime = *chainConfig.Expiration
	}

	metadata := "0x"
	signature, err := p.paymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
	if err != nil {
		log.Error(ctx, "failed to create payment request signature", "err", err)
		return nil, err
	}

	signerAddress, err := p.getSignerAddress(ctx, chainConfig.SigningKeyID)
	if err != nil {
		log.Error(ctx, "failed to retrieve signer address", "err", err)
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
			Proof:          paymentProof(&setting, signature, signerAddress),
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
			Proof:          paymentProof(&setting, signature, signerAddress),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported payment option type: %s", setting.PaymentOption.Type)
	}
}

func paymentProof(setting *payments.ChainConfig, signature []byte, signerAddress common.Address) protocol.PaymentProof {
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
			VerificationMethod: fmt.Sprintf("did:pkh:eip155:%d:%s", setting.ChainID, signerAddress),
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

	const recoveryIdOffset = 64
	if len(signature) > recoveryIdOffset {
		if signature[recoveryIdOffset] <= 1 {
			signature[recoveryIdOffset] += 27
		}
	}

	return signature, nil
}

func (p *payment) getSignerAddress(ctx context.Context, signingKeyID string) (common.Address, error) {
	bytesPubKey, err := p.kms.PublicKey(kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   signingKeyID,
	})
	if err != nil {
		return common.Address{}, err
	}
	var pubKey *ecdsa.PublicKey
	switch len(bytesPubKey) {
	case eth.CompressedPublicKeyLength:
		pubKey, err = crypto.DecompressPubkey(bytesPubKey)
	case eth.AwsKmsPublicKeyLength:
		pubKey, err = kms.DecodeAWSETHPubKey(ctx, bytesPubKey)
		if err != nil {
			return common.Address{}, err
		}
	default:
		pubKey, err = crypto.UnmarshalPubkey(bytesPubKey)
	}
	if err != nil {
		return common.Address{}, err
	}
	fromAddress := crypto.PubkeyToAddress(*pubKey)
	return fromAddress, nil
}
