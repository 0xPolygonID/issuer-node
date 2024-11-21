package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/eth"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

const (
	iden3PaymentRailsRequestV1Type      = "Iden3PaymentRailsRequestV1"
	iden3PaymentRailsERC20RequestV1Type = "Iden3PaymentRailsERC20RequestV1"
)

type payment struct {
	networkResolver network.Resolver
	settings        payments.Settings
	paymentsStore   ports.PaymentRepository
}

// NewPaymentService creates a new payment service
func NewPaymentService(payOptsRepo ports.PaymentRepository, resolver network.Resolver, settings payments.Settings) ports.PaymentService {
	return &payment{
		networkResolver: resolver,
		settings:        settings,
		paymentsStore:   payOptsRepo,
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
	opt, err := p.paymentsStore.GetPaymentOptionByID(ctx, *issuerDID, id)
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
func (p *payment) CreatePaymentRequest(ctx context.Context, req *ports.CreatePaymentRequestReq) (*protocol.PaymentRequestMessage, error) {
	const defaultExpirationDate = 1 * time.Hour

	option, err := p.paymentsStore.GetPaymentOptionByID(ctx, req.IssuerDID, req.OptionID)
	if err != nil {
		log.Error(ctx, "failed to get payment option", "err", err, "issuerDID", req.IssuerDID, "optionID", req.OptionID)
		return nil, err
	}

	paymentsList := make([]protocol.PaymentRequestInfo, 0)
	for _, chainConfig := range option.Config.Chains {
		setting, found := p.settings[chainConfig.ChainId]
		if !found {
			log.Error(ctx, "chain not found in settings", "chainId", chainConfig.ChainId)
			return nil, fmt.Errorf("chain not <%d> not found in payment settings", chainConfig.ChainId)
		}

		//nolint: mnd
		nonce, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(130), nil))
		if err != nil {
			return nil, err
		}
		expirationTime := time.Now().Add(defaultExpirationDate)
		if chainConfig.Iden3PaymentRailsRequestV1 != nil {
			data, err := p.newIden3PaymentRailsRequestV1(ctx, chainConfig, setting, expirationTime, nonce)
			if err != nil {
				log.Error(ctx, "failed to create Iden3PaymentRailsRequestV1", "err", err)
				return nil, err
			}

			paymentsList = append(paymentsList, protocol.PaymentRequestInfo{
				Description: option.Description,
				Credentials: req.Creds,
				Data:        *data,
			})
		}
		if chainConfig.Iden3PaymentRailsERC20RequestV1 != nil {
			data, err := p.newIden3PaymentRailsERC20RequestV1(ctx, chainConfig, setting, expirationTime, nonce)
			if err != nil {
				log.Error(ctx, "failed to create Iden3PaymentRailsRequestV1", "err", err)
				return nil, err
			}

			paymentsList = append(paymentsList, protocol.PaymentRequestInfo{
				Description: option.Description,
				Credentials: req.Creds,
				Data:        *data,
			})

		}

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
			Agent:    "localhost", // TODO!!!,
			Payments: paymentsList,
		},
	}

	return message, nil
}

// TODO: something
func (p *payment) newIden3PaymentRailsERC20RequestV1(ctx context.Context, chainConfig domain.PaymentOptionConfigChain, setting payments.ChainSettings, expirationTime time.Time, nonce *big.Int) (*protocol.PaymentRequestInfoData, error) {
	data := protocol.NewPaymentRequestInfoDataRailsERC20(protocol.Iden3PaymentRailsERC20RequestV1{})
	return &data, nil
}

func (p *payment) newIden3PaymentRailsRequestV1(ctx context.Context, chainConfig domain.PaymentOptionConfigChain, setting payments.ChainSettings, expirationTime time.Time, nonce *big.Int) (*protocol.PaymentRequestInfoData, error) {
	metadata := "0x"
	signature, address, err := p.paymentRequestSignature(chainConfig.ChainId, setting.MCPayment, chainConfig.Iden3PaymentRailsRequestV1.Amount, expirationTime, nonce, metadata, chainConfig.SigningKeyId)
	if err != nil {
		log.Error(ctx, "failed to create payment request signature", "err", err)
		return nil, err
	}

	paymentInfo := protocol.NewPaymentRequestInfoDataRails(protocol.Iden3PaymentRailsRequestV1{
		Nonce: nonce.String(),
		Type:  iden3PaymentRailsRequestV1Type,
		Context: protocol.NewPaymentContextStringCol([]string{
			"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
			"https://w3id.org/security/suites/eip712sig-2021/v1",
		}),
		Amount:         chainConfig.Iden3PaymentRailsRequestV1.Amount,
		ExpirationDate: fmt.Sprint(expirationTime.Unix()),
		Metadata:       metadata,
		Currency:       chainConfig.Iden3PaymentRailsRequestV1.Currency,
		Recipient:      address.String(),
		Proof: protocol.NewPaymentProofEip712Signature([]protocol.EthereumEip712Signature2021{
			{
				Type:               "EthereumEip712Signature2021",
				ProofPurpose:       "assertionMethod",
				ProofValue:         hex.EncodeToString(signature),
				VerificationMethod: fmt.Sprintf("did:pkh:eip155:%d:%s", chainConfig.ChainId, address),
				Created:            time.Now().Format(time.RFC3339),
				Eip712: protocol.Eip712Data{
					Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsERC20RequestV1.json",
					PrimaryType: string(protocol.Iden3PaymentRailsRequestV1Type),
					Domain: protocol.Eip712Domain{
						Name:              "MCPayment",
						Version:           "1.0.0",
						ChainID:           strconv.Itoa(chainConfig.ChainId),
						VerifyingContract: setting.MCPayment,
					},
				},
			},
		}),
	})
	return &paymentInfo, nil
}

func (p *payment) paymentRequestSignature(chainID int, verifContract string, amount string, expTime time.Time, nonce *big.Int, metadata string, signingKeyId string) ([]byte, common.Address, error) {
	// TODO: Sign message with SigningKeyId. Get the "recipient"? address from private key
	/*
		signingKey := "0x" + option.Config.Chains[0].SigningKeyId
		privateKeyBytes, err := hex.DecodeString(signingKey)
		if err != nil {
			return nil, common.Address{}, err
		}
		privateKey, err := crypto.ToECDSA(privateKeyBytes)
		if err != nil {
			return nil, common.Address{}, err
		}
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		if !ok {
			return nil, common.Address{}, errors.New("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
		}

		address := crypto.PubkeyToAddress(*publicKeyECDSA)
	*/
	address := common.Address{} // Get from private key
	typedData, err := typedDataForHashing(iden3PaymentRailsRequestV1Type, chainID, verifContract, address, amount, expTime, nonce, metadata)
	if err != nil {
		return nil, common.Address{}, err
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, common.Address{}, err
	}

	// TODO: Sign the hash with the private key
	/*
		signature, err := crypto.Sign(typedDataHash[:], privateKey)
		if err != nil {
			return nil, common.Address{}, err
		}
	*/
	_ = typedDataHash
	signature := []byte{} // Sign with private key
	return signature, address, nil
}

func typedDataForHashing(paymentType string, chainID int, verifyContract string, address common.Address, amount string, expTime time.Time, nonce *big.Int, metadata string) (*apitypes.TypedData, error) {
	if paymentType != iden3PaymentRailsRequestV1Type && paymentType != iden3PaymentRailsERC20RequestV1Type {
		return nil, fmt.Errorf("unsupported payment type: %s", paymentType)
	}

	data := apitypes.TypedData{
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
		PrimaryType: "Iden3PaymentRailsRequestV1",
		Domain: apitypes.TypedDataDomain{
			Name:              "MCPayment",
			Version:           "1.0.0",
			ChainId:           math.NewHexOrDecimal256(int64(chainID)),
			VerifyingContract: verifyContract, // 2. config
		},
		Message: apitypes.TypedDataMessage{
			"recipient":      address,
			"amount":         amount,
			"expirationDate": fmt.Sprint(expTime.Unix()),
			"nonce":          nonce.String(),
			"metadata":       metadata,
		},
	}
	if paymentType == iden3PaymentRailsERC20RequestV1Type {
		data.Types[paymentType] = append(data.Types[paymentType], apitypes.Type{
			Name: "tokenAddress",
			Type: "address",
		})
		data.Message["tokenAddress"] = "" // TODO: What is this?
	}
	return &data, nil
}

// CreatePaymentRequestForProposalRequest creates a payment request for a proposal request
func (p *payment) CreatePaymentRequestForProposalRequest(ctx context.Context, proposalRequest *protocol.CredentialsProposalRequestMessage) (*comm.BasicMessage, error) {
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

func (p *payment) VerifyPayment(ctx context.Context, recipient common.Address, message comm.BasicMessage) (bool, error) {
	var paymentRequest protocol.PaymentRequestMessageBody
	err := json.Unmarshal(message.Body, &paymentRequest)
	if err != nil {
		return false, err
	}

	client, err := ethclient.Dial("https://polygon-amoy.g.alchemy.com/v2/DHvucvBBzrBhaHzmjrMp24PGbl7vwee6")
	if err != nil {
		return false, err
	}

	contractAddress := common.HexToAddress("0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774")
	instance, err := eth.NewPaymentContract(contractAddress, client)
	if err != nil {
		return false, err
	}

	// nonce, _ := new(big.Int).SetString(paymentRequest.Payments[0].Nonce, base10)
	nonce, err := nonceFromPaymentRequestInfoData(paymentRequest.Payments[0].Data)
	if err != nil {
		log.Error(ctx, "failed to get nonce from payment request info data", "err", err)
		return false, err
	}
	isPaid, err := instance.IsPaymentDone(&bind.CallOpts{Context: context.Background()}, recipient, nonce)
	if err != nil {
		return false, err
	}
	return isPaid, nil
}

func nonceFromPaymentRequestInfoData(data protocol.PaymentRequestInfoData) (*big.Int, error) {
	const base10 = 10
	var nonce string
	switch data.Type() {
	case protocol.Iden3PaymentRequestCryptoV1Type:
		nonce = ""
	case protocol.Iden3PaymentRailsRequestV1Type:
		t, ok := data.Data().(protocol.Iden3PaymentRailsRequestV1)
		if !ok {
			return nil, fmt.Errorf("failed to cast payment request data to Iden3PaymentRailsRequestV1")
		}
		nonce = t.Nonce
	case protocol.Iden3PaymentRailsERC20RequestV1Type:
		t, ok := data.Data().(protocol.Iden3PaymentRailsERC20RequestV1)
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
