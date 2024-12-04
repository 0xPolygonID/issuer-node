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
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
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
	settings        payments.Settings
	paymentsStore   ports.PaymentRepository
	kms             kms.KMSType
}

// NewPaymentService creates a new payment service
func NewPaymentService(payOptsRepo ports.PaymentRepository, resolver network.Resolver, settings payments.Settings, kms kms.KMSType) ports.PaymentService {
	return &payment{
		networkResolver: resolver,
		settings:        settings,
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
func (p *payment) CreatePaymentRequest(ctx context.Context, req *ports.CreatePaymentRequestReq, baseURL string) (*protocol.PaymentRequestMessage, error) {
	const defaultExpirationDate = 1 * time.Hour

	option, err := p.paymentsStore.GetPaymentOptionByID(ctx, &req.IssuerDID, req.OptionID)
	if err != nil {
		log.Error(ctx, "failed to get payment option", "err", err, "issuerDID", req.IssuerDID, "optionID", req.OptionID)
		return nil, err
	}
	var dataArr protocol.PaymentRequestInfoData
	for _, chainConfig := range option.Config.Chains {
		setting, found := p.settings[chainConfig.ChainId]
		if !found {
			log.Error(ctx, "chain not found in settings", "chainId", chainConfig.ChainId)
			return nil, fmt.Errorf("chain not <%d> not found in payment settings", chainConfig.ChainId)
		}

		expirationTime := time.Now().Add(defaultExpirationDate)
		var address common.Address
		var privateKey *ecdsa.PrivateKey

		pubKey, err := kms.EthPubKey(ctx, p.kms, kms.KeyID{ID: chainConfig.SigningKeyId, Type: kms.KeyTypeEthereum})
		if err != nil {
			log.Error(ctx, "failed to get kms signing key", "err", err, "keyId", chainConfig.SigningKeyId)
			return nil, fmt.Errorf("kms signing key not found: %w", err)
		}
		address = crypto.PubkeyToAddress(*pubKey)

		if chainConfig.Iden3PaymentRailsRequestV1 != nil {
			nativeToken, err := p.newIden3PaymentRailsRequestV1(ctx, chainConfig, setting, expirationTime, address, privateKey)
			if err != nil {
				log.Error(ctx, "failed to create Iden3PaymentRailsRequestV1", "err", err)
				return nil, err
			}

			dataArr = append(dataArr, *nativeToken)
		}

		if chainConfig.Iden3PaymentRailsERC20RequestV1 != nil {
			reqUSDT, err := p.newIden3PaymentRailsERC20RequestV1(ctx, chainConfig, setting, expirationTime, address, payments.USDT, chainConfig.Iden3PaymentRailsERC20RequestV1.USDT.Amount, setting.ERC20.USDT.ContractAddress, privateKey)
			if err != nil {
				log.Error(ctx, "failed to create Iden3PaymentRailsRequestV1", "err", err)
				return nil, err
			}
			dataArr = append(dataArr, reqUSDT)
			reqUSDC, err := p.newIden3PaymentRailsERC20RequestV1(ctx, chainConfig, setting, expirationTime, address, payments.USDC, chainConfig.Iden3PaymentRailsERC20RequestV1.USDC.Amount, setting.ERC20.USDC.ContractAddress, privateKey)
			if err != nil {
				log.Error(ctx, "failed to create Iden3PaymentRailsRequestV1", "err", err)
				return nil, err
			}
			dataArr = append(dataArr, reqUSDC)
		}
	}

	payments := []protocol.PaymentRequestInfo{
		{
			Description: option.Description,
			Credentials: req.Creds,
			Data:        dataArr,
		},
	}
	msgID := uuid.New()
	message := &protocol.PaymentRequestMessage{
		From:     req.IssuerDID.String(),
		To:       req.UserDID.String(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.PaymentRequestMessageType,
		ID:       msgID.String(),
		ThreadID: msgID.String(),
		Body: protocol.PaymentRequestMessageBody{
			Agent:    fmt.Sprintf(ports.AgentUrl, baseURL),
			Payments: payments,
		},
	}

	return message, nil
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
func (p *payment) GetSettings() payments.Settings {
	return p.settings
}

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

func contractAddressFromPayment(data *protocol.Payment, config payments.Settings) (*common.Address, error) {
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
}

func recipientAddressFromPayment(data *protocol.Payment, option *domain.PaymentOption) (*common.Address, error) {
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

func (p *payment) newIden3PaymentRailsRequestV1(
	ctx context.Context,
	chainConfig domain.PaymentOptionConfigChain,
	setting payments.ChainSettings,
	expirationTime time.Time,
	address common.Address,
	signingKeyOpt *ecdsa.PrivateKey, // Temporary solution until we have key management
) (*protocol.Iden3PaymentRailsRequestV1, error) {
	nonce, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(130), nil))
	if err != nil {
		log.Error(ctx, "failed to generate nonce", "err", err)
		return nil, err
	}
	metadata := "0x"
	signature, err := p.paymentRequestSignature(ctx, protocol.Iden3PaymentRailsRequestV1Type, chainConfig.ChainId, setting.MCPayment, chainConfig.Iden3PaymentRailsRequestV1.Amount, "USDT", expirationTime, nonce, metadata, address, chainConfig.SigningKeyId, signingKeyOpt, "")
	if err != nil {
		log.Error(ctx, "failed to create payment request signature", "err", err)
		return nil, err
	}

	amountString := fmt.Sprintf("%f", chainConfig.Iden3PaymentRailsRequestV1.Amount)
	if chainConfig.Iden3PaymentRailsRequestV1.Amount == float64(int(chainConfig.Iden3PaymentRailsRequestV1.Amount)) {
		// No decimals, convert to integer string
		amountString = strconv.Itoa(int(chainConfig.Iden3PaymentRailsRequestV1.Amount))
	}
	paymentInfo := protocol.Iden3PaymentRailsRequestV1{
		Nonce: nonce.String(),
		Type:  protocol.Iden3PaymentRailsRequestV1Type,
		Context: protocol.NewPaymentContextString(
			"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
			"https://w3id.org/security/suites/eip712sig-2021/v1",
		),
		Amount:         amountString,
		ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
		Metadata:       metadata,
		Currency:       chainConfig.Iden3PaymentRailsRequestV1.Currency,
		Recipient:      address.String(),
		Proof: protocol.PaymentProof{
			protocol.EthereumEip712Signature2021{
				Type:               "EthereumEip712Signature2021",
				ProofPurpose:       "assertionMethod",
				ProofValue:         fmt.Sprintf("0x%s", hex.EncodeToString(signature)),
				VerificationMethod: fmt.Sprintf("did:pkh:eip155:%d:%s", chainConfig.ChainId, address),
				Created:            time.Now().Format(time.RFC3339),
				Eip712: protocol.Eip712Data{
					Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsRequestV1.json",
					PrimaryType: string(protocol.Iden3PaymentRailsRequestV1Type),
					Domain: protocol.Eip712Domain{
						Name:              "MCPayment",
						Version:           "1.0.0",
						ChainID:           strconv.Itoa(chainConfig.ChainId),
						VerifyingContract: setting.MCPayment,
					},
				},
			},
		},
	}
	return &paymentInfo, nil
}

// newIden3PaymentRailsERC20RequestV1 creates a new Iden3PaymentRailsERC20RequestV1
func (p *payment) newIden3PaymentRailsERC20RequestV1(
	ctx context.Context,
	chainConfig domain.PaymentOptionConfigChain,
	setting payments.ChainSettings,
	expirationTime time.Time,
	address common.Address,
	currency payments.Coin,
	amount float64,
	tokenAddress string,
	signingKeyOpt *ecdsa.PrivateKey, // Temporary solution until we have key management
) (*protocol.Iden3PaymentRailsERC20RequestV1, error) {
	nonce, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(130), nil))
	if err != nil {
		log.Error(ctx, "failed to generate nonce", "err", err)
		return nil, err
	}
	metadata := "0x"
	signature, err := p.paymentRequestSignature(ctx, protocol.Iden3PaymentRailsERC20RequestV1Type, chainConfig.ChainId, setting.MCPayment, chainConfig.Iden3PaymentRailsERC20RequestV1.USDT.Amount, "USDT", expirationTime, nonce, metadata, address, chainConfig.SigningKeyId, signingKeyOpt, tokenAddress)
	if err != nil {
		log.Error(ctx, "failed to create payment request signature", "err", err)
		return nil, err
	}
	amountString := fmt.Sprintf("%f", amount)
	if amount == float64(int(amount)) {
		// No decimals, convert to integer string
		amountString = strconv.Itoa(int(amount))
	}
	var features []protocol.PaymentFeatures = []protocol.PaymentFeatures{}
	if currency == payments.USDC {
		features = []protocol.PaymentFeatures{
			"EIP-2612",
		}
	}
	paymentInfo := protocol.Iden3PaymentRailsERC20RequestV1{
		Nonce: nonce.String(),
		Type:  protocol.Iden3PaymentRailsERC20RequestV1Type,
		Context: protocol.NewPaymentContextString(
			"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
			"https://w3id.org/security/suites/eip712sig-2021/v1",
		),
		Amount:         amountString,
		ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
		Metadata:       metadata,
		Currency:       string(currency),
		Recipient:      address.String(),
		Features:       features,
		TokenAddress:   tokenAddress,
		Proof: protocol.PaymentProof{
			protocol.EthereumEip712Signature2021{
				Type:               "EthereumEip712Signature2021",
				ProofPurpose:       "assertionMethod",
				ProofValue:         fmt.Sprintf("0x%s", hex.EncodeToString(signature)),
				VerificationMethod: fmt.Sprintf("did:pkh:eip155:%d:%s", chainConfig.ChainId, address),
				Created:            time.Now().Format(time.RFC3339),
				Eip712: protocol.Eip712Data{
					Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsERC20RequestV1.json",
					PrimaryType: string(protocol.Iden3PaymentRailsERC20RequestV1Type),
					Domain: protocol.Eip712Domain{
						Name:              "MCPayment",
						Version:           "1.0.0",
						ChainID:           strconv.Itoa(chainConfig.ChainId),
						VerifyingContract: setting.MCPayment,
					},
				},
			},
		},
	}
	return &paymentInfo, nil
}

func (p *payment) paymentRequestSignature(
	ctx context.Context,
	paymentType protocol.PaymentRequestType,
	chainID int,
	verifContract string,
	amount float64,
	currency string,
	expTime time.Time,
	nonce *big.Int,
	metadata string,
	addr common.Address,
	signingKeyId string,
	signingKeyOpt *ecdsa.PrivateKey, // Temporary solution until we have key management
	tokenAddr string,
) ([]byte, error) {
	if paymentType != protocol.Iden3PaymentRailsRequestV1Type && paymentType != protocol.Iden3PaymentRailsERC20RequestV1Type {
		return nil, fmt.Errorf("unsupported payment type: %s", paymentType)
	}

	keyID := kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   signingKeyId,
	}

	typedData, err := typedDataForHashing(paymentType, chainID, verifContract, addr, amount, currency, expTime, nonce, metadata, tokenAddr)
	if err != nil {
		return nil, err
	}

	typedDataBytes, _, err := apitypes.TypedDataAndHash(*typedData)
	if err != nil {
		return nil, err
	}

	if signingKeyOpt == nil {
		signature, err := p.kms.Sign(ctx, keyID, typedDataBytes)
		if err != nil {
			log.Error(ctx, "failed to sign typed data hash", "err", err, "keyId", signingKeyId)
			return nil, err
		}

		return signature, nil
	} else { // Temporary solution until we have key management. We use SigningKeyOpt as a private key if present
		signature, err := crypto.Sign(typedDataBytes, signingKeyOpt)
		signature[64] += 27
		if err != nil {
			return nil, err
		}
		return signature, nil
	}
}

func typedDataForHashing(paymentType protocol.PaymentRequestType, chainID int, verifyContract string, address common.Address, amount float64, _ string, expTime time.Time, nonce *big.Int, metadata string, tokenAddress string) (*apitypes.TypedData, error) {
	data := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			string(paymentType): []apitypes.Type{
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
		PrimaryType: string(paymentType),
		Domain: apitypes.TypedDataDomain{
			Name:              "MCPayment",
			Version:           "1.0.0",
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: verifyContract,
		},
		Message: apitypes.TypedDataMessage{
			"recipient":      address.String(),
			"amount":         big.NewInt(int64(amount)),
			"expirationDate": fmt.Sprint(expTime.Unix()),
			"nonce":          nonce.String(),
			"metadata":       metadata,
		},
	}
	if paymentType == protocol.Iden3PaymentRailsERC20RequestV1Type {
		data.Types[string(paymentType)] = append(data.Types[string(paymentType)], apitypes.Type{
			Name: "tokenAddress",
			Type: "address",
		})
		data.Message["tokenAddress"] = tokenAddress
	}
	return &data, nil
}
