package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	b64 "encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"github.com/google/uuid"
	abi "github.com/iden3/contracts-abi/multi-chain-payment/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/near/borsh-go"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/eth"
	"github.com/polygonid/sh-id-platform/internal/kms"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

const (
	SolanaDevChainID  = 103
	SolanaTestChainID = 102
	SolanaMainChainID = 101

	SolanaChainRefDevnet  = "EtWTRABZaYq6iMfeYKouRu166VU2xqa1"
	SolanaChainRefTestnet = "4uhcVJyU9pJkvQyS88uRDiswHXSCkY3z"
	SolanaChainRefMainnet = "5eykt4UsFv8P8NJdTREpY1vzqKqZKvdp"
)

type payment struct {
	networkResolver                      network.Resolver
	settings                             payments.Config
	schemaService                        ports.SchemaService
	paymentsStore                        ports.PaymentRepository
	kms                                  kms.KMSType
	iden3PaymentRailsRequestV1Types      apitypes.Types
	iden3PaymentRailsERC20RequestV1Types apitypes.Types
}

// NewPaymentService creates a new payment service
func NewPaymentService(payOptsRepo ports.PaymentRepository, resolver network.Resolver, schemaSrv ports.SchemaService, settings *payments.Config, kms kms.KMSType) (ports.PaymentService, error) {
	iden3PaymentRailsRequestV1Types := apitypes.Types{}
	iden3PaymentRailsERC20RequestV1Types := apitypes.Types{}
	err := json.Unmarshal([]byte(domain.Iden3PaymentRailsRequestV1SchemaJSON), &iden3PaymentRailsRequestV1Types)
	if err != nil {
		log.Error(context.Background(), "failed to unmarshal Iden3PaymentRailsRequestV1 schema", "err", err)
		return nil, err
	}
	err = json.Unmarshal([]byte(domain.Iden3PaymentRailsERC20RequestV1SchemaJSON), &iden3PaymentRailsERC20RequestV1Types)
	if err != nil {
		log.Error(context.Background(), "failed to unmarshal Iden3PaymentRailsERC20RequestV1 schema", "err", err)
		return nil, err
	}
	return &payment{
		networkResolver:                      resolver,
		settings:                             *settings,
		schemaService:                        schemaSrv,
		paymentsStore:                        payOptsRepo,
		kms:                                  kms,
		iden3PaymentRailsRequestV1Types:      iden3PaymentRailsRequestV1Types,
		iden3PaymentRailsERC20RequestV1Types: iden3PaymentRailsERC20RequestV1Types,
	}, nil
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

func (p *payment) UpdatePaymentOption(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID, name, description *string, config *domain.PaymentOptionConfig) error {
	paymentOption, err := p.GetPaymentOptionByID(ctx, issuerDID, id)
	if err != nil {
		log.Error(ctx, "failed to get payment option", "err", err, "issuerDID", issuerDID, "id", id)
		return err
	}

	if name != nil {
		paymentOption.Name = *name
	}

	if description != nil {
		paymentOption.Description = *description
	}

	if config != nil {
		paymentOption.Config = *config
	}

	_, err = p.paymentsStore.SavePaymentOption(ctx, paymentOption)
	return err
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

	createTime := time.Now()
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
		SchemaID:        &schema.ID,
		Description:     req.Description,
		PaymentOptionID: req.OptionID,
		CreatedAt:       createTime,
		ModifietAt:      createTime,
		Status:          domain.PaymentRequestStatusNotVerified,
	}
	for _, chainConfig := range option.Config.PaymentOptions {
		setting, found := p.settings[chainConfig.PaymentOptionID]
		if !found {
			log.Error(ctx, "chain not found in configuration", "paymentOptionID", chainConfig.PaymentOptionID)
			return nil, fmt.Errorf("payment Option <%d> not found in payment configuration", chainConfig.PaymentOptionID)
		}

		nonce, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(64), nil)) //nolint: mnd
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
func (p *payment) GetPaymentRequests(ctx context.Context, issuerDID *w3c.DID, queryParams *domain.PaymentRequestsQueryParams) ([]domain.PaymentRequest, error) {
	paymentRequests, err := p.paymentsStore.GetAllPaymentRequests(ctx, *issuerDID, queryParams)
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
func (p *payment) VerifyPayment(ctx context.Context, issuerDID w3c.DID, nonce *big.Int, txHash *string, userDID *w3c.DID) (status ports.BlockchainPaymentStatus, paymentRequestID uuid.UUID, err error) {
	paymentReqItem, err := p.paymentsStore.GetPaymentRequestItem(ctx, issuerDID, nonce)
	if err != nil {
		return ports.BlockchainPaymentStatusPending, uuid.Nil, fmt.Errorf("failed to get payment request: %w", err)
	}

	paymentReq, err := p.paymentsStore.GetPaymentRequestByID(ctx, issuerDID, paymentReqItem.PaymentRequestID)
	if err != nil {
		return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, fmt.Errorf("failed to get payment request: %w", err)
	}

	if userDID != nil {
		if userDID.String() != paymentReq.UserDID.String() {
			return ports.BlockchainPaymentStatusFailed, paymentReqItem.PaymentRequestID, fmt.Errorf("userDID %s does not match to User DID %s in payment-request", userDID, paymentReq.UserDID)
		}
	}

	setting, found := p.settings[paymentReqItem.PaymentOptionID]
	if !found {
		log.Error(ctx, "chain not found in configuration", "paymentOptionID", paymentReqItem.PaymentOptionID)
		return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, fmt.Errorf("payment Option <%d> not found in payment configuration", paymentReqItem.PaymentOptionID)
	}

	if setting.PaymentOption.Type == protocol.Iden3PaymentRailsSolanaRequestV1Type ||
		setting.PaymentOption.Type == protocol.Iden3PaymentRailsSolanaSPLRequestV1Type {
		signerAddress, err := p.getSolSignerAddress(ctx, paymentReqItem.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to get signer address", "err", err, "SigningKeyID", paymentReqItem.SigningKeyID)
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, err
		}
		status, err = p.verifySolanaPaymentOnBlockchain(ctx, setting, nonce, signerAddress, txHash)
		if err != nil {
			log.Error(ctx, "failed to verify Solana payment on blockchain", "err", err, "txHash", txHash, "nonce", nonce)
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, err
		}
	} else {
		client, err := p.networkResolver.GetEthClientByChainID(core.ChainID(setting.ChainID))
		if err != nil {
			log.Error(ctx, "failed to get ethereum client from resolvers", "err", err, "chainID", setting.ChainID)
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, fmt.Errorf("failed to get ethereum client from resolvers settings for chainID <%d>", setting.ChainID)
		}

		instance, err := abi.NewMCPayment(common.HexToAddress(setting.PaymentRails), client.GetEthereumClient())
		if err != nil {
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, err
		}

		signerAddress, err := p.getEthSignerAddress(ctx, paymentReqItem.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to get signer address", "err", err, "SigningKeyID", paymentReqItem.SigningKeyID)
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, err
		}

		status, err = p.verifyPaymentOnBlockchain(ctx, client, instance, signerAddress, nonce, txHash)
		if err != nil {
			log.Error(ctx, "failed to verify payment on blockchain", "err", err, "txHash", txHash, "nonce", nonce)
			return ports.BlockchainPaymentStatusPending, paymentReqItem.PaymentRequestID, err
		}
	}

	paymentReqStatus := getPaymentRequestStatusFromBlockChainStatus(status)
	if paymentReqStatus != paymentReq.Status && paymentReq.Status != domain.PaymentRequestStatusSuccess {
		var paidNonce *big.Int
		if paymentReqStatus == domain.PaymentRequestStatusSuccess {
			paidNonce = nonce
		}
		err = p.paymentsStore.UpdatePaymentRequestStatus(ctx, issuerDID, paymentReq.ID, paymentReqStatus, paidNonce)
		if err != nil {
			log.Error(ctx, "failed to update payment-request with new status", "err", err, "txHash", txHash, "nonce", nonce, "status", status)
			return status, paymentReqItem.PaymentRequestID, err
		}

	}

	return status, paymentReqItem.PaymentRequestID, nil
}

func getPaymentRequestStatusFromBlockChainStatus(status ports.BlockchainPaymentStatus) domain.PaymentRequestStatus {
	switch status {
	case ports.BlockchainPaymentStatusPending:
		return domain.PaymentRequestStatusPending
	case ports.BlockchainPaymentStatusSuccess:
		return domain.PaymentRequestStatusSuccess
	case ports.BlockchainPaymentStatusCancelled:
		return domain.PaymentRequestStatusCanceled
	case ports.BlockchainPaymentStatusFailed:
		return domain.PaymentRequestStatusFailed
	default:
		return domain.PaymentRequestStatusNotVerified
	}
}

func (p *payment) verifyPaymentOnBlockchain(
	ctx context.Context,
	client *eth.Client,
	contract *abi.MCPayment,
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

func (p *payment) verifySolanaPaymentOnBlockchain(ctx context.Context, setting payments.ChainConfig, nonce *big.Int, signer string, txHash *string) (ports.BlockchainPaymentStatus, error) {
	var client *rpc.Client
	switch setting.ChainID {
	case SolanaDevChainID:
		client = rpc.New(rpc.DevNet_RPC)
	case SolanaTestChainID:
		client = rpc.New(rpc.TestNet_RPC)
	case SolanaMainChainID:
		client = rpc.New(rpc.MainNetBeta_RPC)
	default:
		log.Error(ctx, "unsupported chain ID for Solana payment verification", "chainID", setting.ChainID)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("unsupported chain ID for Solana payment verification: %d", setting.ChainID)
	}
	programID, err := solana.PublicKeyFromBase58(setting.PaymentRails)
	if err != nil {
		log.Error(ctx, "failed to parse program ID", "err", err, "programID", setting.PaymentRails)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to parse program ID: %w", err)
	}

	txIdProvided := txHash != nil && *txHash != ""
	if txIdProvided {
		status, err := handleSolanaPaymentTransaction(ctx, client, *txHash)
		if err != nil || status != ports.BlockchainPaymentStatusSuccess {
			return status, err
		}
	}

	bytesForUint64 := 8
	nonceLe := make([]byte, bytesForUint64)
	binary.LittleEndian.PutUint64(nonceLe, nonce.Uint64())

	pubKey, err := solana.PublicKeyFromBase58(signer)
	if err != nil {
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to parse signer public key: %w", err)
	}
	seeds := [][]byte{
		[]byte("payment"),
		pubKey.Bytes(),
		nonceLe,
	}
	pda, _, err := solana.FindProgramAddress(seeds, programID)
	if err != nil {
		log.Error(ctx, "failed to find program address", "err", err, "programID", programID, "seeds", seeds)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to find program address: %w", err)
	}

	ai, err := client.GetAccountInfo(ctx, pda)
	if err != nil {
		log.Error(ctx, "failed to get account info", "err", err, "pda", pda)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to get account info: %w", err)
	}

	if ai == nil || ai.Value == nil {
		log.Error(ctx, "account info not found", "pda", pda)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("account info not found for PDA: %s", pda)
	}

	data := ai.Value.Data.GetBinary()
	var paymentRecord paymentRecord
	err = borsh.Deserialize(&paymentRecord, data)
	if err != nil {
		log.Error(ctx, "failed to deserialize payment request", "err", err, "pda", pda, "data", data)
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to deserialize payment request: %w", err)
	}

	if txHash != nil && *txHash != "" {
		return handleSolanaPaymentTransaction(ctx, client, *txHash)
	}

	if paymentRecord.IsPaid != 0 {
		return ports.BlockchainPaymentStatusSuccess, nil
	}

	return ports.BlockchainPaymentStatusUnknown, nil
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

func handleSolanaPaymentTransaction(
	ctx context.Context,
	client *rpc.Client,
	txSig string,
) (ports.BlockchainPaymentStatus, error) {
	sig, err := solana.SignatureFromBase58(txSig)
	if err != nil {
		return ports.BlockchainPaymentStatusUnknown, fmt.Errorf("failed to parse transaction signature: %w", err)
	}
	resp, err := client.GetSignatureStatuses(ctx, true, sig)
	if err != nil {
		return ports.BlockchainPaymentStatusUnknown, err
	}

	if len(resp.Value) == 0 || resp.Value[0] == nil {
		// No record in ledger yet — could be dropped or never sent
		return ports.BlockchainPaymentStatusCancelled, nil
	}

	sigStatus := resp.Value[0]

	if sigStatus.Err != nil {
		return ports.BlockchainPaymentStatusFailed, nil
	}

	switch sigStatus.ConfirmationStatus {
	case "processed", "confirmed":
		return ports.BlockchainPaymentStatusPending, nil
	case "finalized":
		return ports.BlockchainPaymentStatusSuccess, nil
	default:
		return ports.BlockchainPaymentStatusUnknown, nil
	}
}

func (p *payment) paymentInfo(ctx context.Context, setting payments.ChainConfig, chainConfig *domain.PaymentOptionConfigItem, nonce *big.Int) (protocol.PaymentRequestInfoDataItem, error) {
	const defaultExpirationDate = 1 * time.Hour
	expirationTime := time.Now().Add(defaultExpirationDate)

	if chainConfig.Expiration != nil {
		expirationTime = *chainConfig.Expiration
	}

	metadata := "0x"
	switch setting.PaymentOption.Type {
	case protocol.Iden3PaymentRailsRequestV1Type:
		signature, err := p.eip712PaymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
		if err != nil {
			log.Error(ctx, "failed to create payment request signature", "err", err)
			return nil, err
		}
		signerAddress, err := p.getEthSignerAddress(ctx, chainConfig.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to retrieve signer address", "err", err)
			return nil, err
		}
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
			Recipient:      chainConfig.Recipient,
			Proof:          eip712PaymentProof(&setting, signature, signerAddress),
		}, nil

	case protocol.Iden3PaymentRailsERC20RequestV1Type:
		signature, err := p.eip712PaymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
		if err != nil {
			log.Error(ctx, "failed to create payment request signature", "err", err)
			return nil, err
		}
		signerAddress, err := p.getEthSignerAddress(ctx, chainConfig.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to retrieve signer address", "err", err)
			return nil, err
		}
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
			Recipient:      chainConfig.Recipient,
			Features:       setting.PaymentOption.Features,
			TokenAddress:   setting.PaymentOption.ContractAddress,
			Proof:          eip712PaymentProof(&setting, signature, signerAddress),
		}, nil
	case protocol.Iden3PaymentRailsSolanaRequestV1Type:
		signature, err := p.ed25519PaymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
		if err != nil {
			log.Error(ctx, "failed to create payment request signature", "err", err)
			return nil, err
		}
		signerAddress, err := p.getSolSignerAddress(ctx, chainConfig.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to retrieve signer address", "err", err, "SigningKeyID", chainConfig.SigningKeyID)
			return nil, err
		}
		return &protocol.Iden3PaymentRailsSolanaRequestV1{
			Nonce: nonce.String(),
			Type:  protocol.Iden3PaymentRailsSolanaRequestV1Type,
			Context: protocol.NewPaymentContextString(
				"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsSolanaRequestV1",
				"https://schema.iden3.io/core/jsonld/solanaEd25519.jsonld",
			),
			Amount:         chainConfig.Amount.String(),
			ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
			Metadata:       metadata,
			Recipient:      chainConfig.Recipient,
			Proof:          solanaEd25519PaymentProof(&setting, signature, signerAddress),
		}, nil
	case protocol.Iden3PaymentRailsSolanaSPLRequestV1Type:
		signature, err := p.ed25519PaymentRequestSignature(ctx, setting, chainConfig, expirationTime, nonce, metadata)
		if err != nil {
			log.Error(ctx, "failed to create payment request signature", "err", err)
			return nil, err
		}
		signerAddress, err := p.getSolSignerAddress(ctx, chainConfig.SigningKeyID)
		if err != nil {
			log.Error(ctx, "failed to retrieve signer address", "err", err, "SigningKeyID", chainConfig.SigningKeyID)
			return nil, err
		}
		return &protocol.Iden3PaymentRailsSolanaSPLRequestV1{
			Nonce: nonce.String(),
			Type:  protocol.Iden3PaymentRailsSolanaSPLRequestV1Type,
			Context: protocol.NewPaymentContextString(
				"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsSolanaSPLRequestV1",
				"https://schema.iden3.io/core/jsonld/solanaEd25519.jsonld",
			),
			Amount:         chainConfig.Amount.String(),
			ExpirationDate: fmt.Sprint(expirationTime.Format(time.RFC3339)),
			Metadata:       metadata,
			Recipient:      chainConfig.Recipient,
			Features:       setting.PaymentOption.Features,
			TokenAddress:   setting.PaymentOption.ContractAddress,
			Proof:          solanaEd25519PaymentProof(&setting, signature, signerAddress),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported payment option type: %s", setting.PaymentOption.Type)
	}
}

func eip712PaymentProof(setting *payments.ChainConfig, signature []byte, signerAddress common.Address) protocol.PaymentProof {
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
					VerifyingContract: setting.PaymentRails,
				},
			},
		},
	}
}

func solanaEd25519PaymentProof(setting *payments.ChainConfig, signature []byte, publicKey string) protocol.PaymentProof {
	var proof protocol.PaymentProof
	verificationMethodChainRef := strconv.Itoa(setting.ChainID)
	switch setting.ChainID {
	case SolanaMainChainID:
		verificationMethodChainRef = SolanaChainRefMainnet
	case SolanaTestChainID:
		verificationMethodChainRef = SolanaChainRefTestnet
	case SolanaDevChainID:
		verificationMethodChainRef = SolanaChainRefDevnet
	}

	switch setting.PaymentOption.Type {
	case protocol.Iden3PaymentRailsSolanaRequestV1Type:
		proof = protocol.PaymentProof{
			protocol.SolanaEd25519Signature2025{
				Type:               protocol.SolanaEd25519Signature2025Type,
				ProofPurpose:       "assertionMethod",
				ProofValue:         hex.EncodeToString(signature),
				VerificationMethod: fmt.Sprintf("did:pkh:solana:%s:%s", verificationMethodChainRef, publicKey),
				Created:            time.Now().Format(time.RFC3339),
				Domain: protocol.SolanaEd25519Domain{
					Version:           "SolanaEd25519NativeV1",
					ChainID:           strconv.Itoa(setting.ChainID),
					VerifyingContract: setting.PaymentRails,
				},
			},
		}
	case protocol.Iden3PaymentRailsSolanaSPLRequestV1Type:
		proof = protocol.PaymentProof{
			protocol.SolanaEd25519Signature2025{
				Type:               protocol.SolanaEd25519Signature2025Type,
				ProofPurpose:       "assertionMethod",
				ProofValue:         hex.EncodeToString(signature),
				VerificationMethod: fmt.Sprintf("did:pkh:solana:%s:%s", verificationMethodChainRef, publicKey),
				Created:            time.Now().Format(time.RFC3339),
				Domain: protocol.SolanaEd25519Domain{
					Version:           "SolanaEd25519SPLV1",
					ChainID:           strconv.Itoa(setting.ChainID),
					VerifyingContract: setting.PaymentRails,
				},
			},
		}
	}
	return proof
}

func (p *payment) eip712PaymentRequestSignature(
	ctx context.Context,
	setting payments.ChainConfig,
	chainConfig *domain.PaymentOptionConfigItem,
	expTime time.Time,
	nonce *big.Int,
	metadata string,
) ([]byte, error) {
	paymentType := string(setting.PaymentOption.Type)

	decodedKeyID, err := b64.StdEncoding.DecodeString(chainConfig.SigningKeyID)
	if err != nil {
		log.Error(ctx, "decoding base64 key id", "err", err)
		return nil, err
	}

	keyID := kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   string(decodedKeyID),
	}

	var types apitypes.Types
	switch paymentType {
	case string(protocol.Iden3PaymentRailsRequestV1Type):
		types = p.iden3PaymentRailsRequestV1Types
	case string(protocol.Iden3PaymentRailsERC20RequestV1Type):
		types = p.iden3PaymentRailsERC20RequestV1Types
	default:
		log.Error(ctx, fmt.Sprintf("unsupported payment type '%s'", paymentType), "err", err)
		return nil, fmt.Errorf("unsupported payment type '%s:'", paymentType)
	}

	typedData := apitypes.TypedData{
		Types:       types,
		PrimaryType: paymentType,
		Domain: apitypes.TypedDataDomain{
			Name:              "MCPayment",
			Version:           "1.0.0",
			ChainId:           math.NewHexOrDecimal256(int64(setting.ChainID)),
			VerifyingContract: setting.PaymentRails,
		},
		Message: apitypes.TypedDataMessage{
			"recipient":      chainConfig.Recipient,
			"amount":         chainConfig.Amount.String(),
			"expirationDate": big.NewInt(expTime.Unix()),
			"nonce":          nonce,
			"metadata":       metadata,
		},
	}
	if paymentType == string(protocol.Iden3PaymentRailsERC20RequestV1Type) {
		typedData.Message["tokenAddress"] = setting.PaymentOption.ContractAddress
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

func (p *payment) getEthSignerAddress(ctx context.Context, signingKeyID string) (common.Address, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(signingKeyID)
	if err != nil {
		log.Error(ctx, "decoding base64 key id", "err", err)
		return common.Address{}, err
	}

	bytesPubKey, err := p.kms.PublicKey(kms.KeyID{
		Type: kms.KeyTypeEthereum,
		ID:   string(decodedKeyID),
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

func (p *payment) getSolSignerAddress(ctx context.Context, signingKeyID string) (string, error) {
	decodedKeyID, err := b64.StdEncoding.DecodeString(signingKeyID)
	if err != nil {
		log.Error(ctx, "decoding base64 key id", "err", err)
		return "", err
	}

	bytesPubKey, err := p.kms.PublicKey(kms.KeyID{
		Type: kms.KeyTypeEd25519,
		ID:   string(decodedKeyID),
	})
	if err != nil {
		return "", err
	}
	pubKey := solana.PublicKey(bytesPubKey)
	return pubKey.String(), nil
}

type paymentRecord struct {
	IsPaid uint8 `borsh:"is_paid"`
}

type solanaNativePaymentRequest struct {
	Version           []byte   `borsh:"version"`
	ChainID           uint64   `borsh:"chainId"`
	VerifyingContract [32]byte `borsh:"verifyingContract"`
	Recipient         [32]byte `borsh:"recipient"`
	Amount            uint64   `borsh:"amount"`
	ExpirationDate    uint64   `borsh:"expirationDate"`
	Nonce             uint64   `borsh:"nonce"`
	Metadata          []byte   `borsh:"metadata"`
}

type solanaSplPaymentRequest struct {
	Version           []byte   `borsh:"version"`
	ChainID           uint64   `borsh:"chainId"`
	VerifyingContract [32]byte `borsh:"verifyingContract"`
	TokenAddress      [32]byte `borsh:"tokenAddress"`
	Recipient         [32]byte `borsh:"recipient"`
	Amount            int64    `borsh:"amount"`
	ExpirationDate    uint64   `borsh:"expirationDate"`
	Nonce             uint64   `borsh:"nonce"`
	Metadata          []byte   `borsh:"metadata"`
}

func (p *payment) ed25519PaymentRequestSignature(
	ctx context.Context,
	setting payments.ChainConfig,
	chainConfig *domain.PaymentOptionConfigItem,
	expTime time.Time,
	nonce *big.Int,
	metadata string,
) (signature []byte, err error) {
	recipient, err := solana.PublicKeyFromBase58(chainConfig.Recipient)
	if err != nil {
		log.Error(ctx, "failed to parse recipient public key", "err", err, "recipient", chainConfig.Recipient)
		return nil, fmt.Errorf("failed to parse recipient public key: %w", err)
	}

	paymentRails, err := solana.PublicKeyFromBase58(setting.PaymentRails)
	if err != nil {
		log.Error(ctx, "failed to parse payment rails public key", "err", err, "paymentRails", setting.PaymentRails)
		return nil, fmt.Errorf("failed to parse payment rails public key: %w", err)
	}

	var serialized []byte
	switch setting.PaymentOption.Type {
	case protocol.Iden3PaymentRailsSolanaRequestV1Type:
		req := solanaNativePaymentRequest{
			Version:           []byte("SolanaEd25519NativeV1"),
			ChainID:           uint64(setting.ChainID),
			VerifyingContract: toKey32(paymentRails),
			Recipient:         toKey32(recipient),
			Amount:            chainConfig.Amount.Uint64(),
			ExpirationDate:    uint64(expTime.Unix()),
			Nonce:             nonce.Uint64(),
			Metadata:          []byte(metadata),
		}
		serialized, err = borsh.Serialize(req)
		if err != nil {
			log.Error(ctx, "failed to serialize solana native payment request", "err", err)
			return nil, fmt.Errorf("failed to serialize solana native payment request: %w", err)
		}
	case protocol.Iden3PaymentRailsSolanaSPLRequestV1Type:
		tokenAddress, err := pubKey32(setting.PaymentOption.ContractAddress)
		if err != nil {
			log.Error(ctx, "failed to parse token address public key", "err", err, "tokenAddress", setting.PaymentOption.ContractAddress)
			return nil, fmt.Errorf("failed to parse token address public key: %w", err)
		}
		req := solanaSplPaymentRequest{
			Version:           []byte("SolanaEd25519SPLV1"),
			ChainID:           uint64(setting.ChainID),
			VerifyingContract: toKey32(paymentRails),
			TokenAddress:      tokenAddress,
			Recipient:         toKey32(recipient),
			Amount:            chainConfig.Amount.Int64(),
			ExpirationDate:    uint64(expTime.Unix()),
			Nonce:             nonce.Uint64(),
			Metadata:          []byte(metadata),
		}
		serialized, err = borsh.Serialize(req)
		if err != nil {
			log.Error(ctx, "failed to serialize solana SPL payment request", "err", err)
			return nil, fmt.Errorf("failed to serialize solana SPL payment request: %w", err)
		}
	default:
		log.Error(ctx, fmt.Sprintf("unsupported payment type '%s'", setting.PaymentOption.Type), "err", err)
		return nil, fmt.Errorf("unsupported payment type '%s:'", setting.PaymentOption.Type)
	}

	decodedKeyID, err := b64.StdEncoding.DecodeString(chainConfig.SigningKeyID)
	if err != nil {
		log.Error(ctx, "decoding base64 key id", "err", err)
		return nil, err
	}

	keyID := kms.KeyID{
		Type: kms.KeyTypeEd25519,
		ID:   string(decodedKeyID),
	}

	signature, err = p.kms.Sign(ctx, keyID, serialized)
	if err != nil {
		log.Error(ctx, "failed to sign typed data hash", "err", err, "keyId", keyID)
		return nil, fmt.Errorf("failed to sign serialized data (ed25519): %w", err)
	}

	return signature, nil
}

func pubKey32(b58 string) ([32]byte, error) {
	var out [32]byte
	pk, err := solana.PublicKeyFromBase58(b58)
	if err != nil {
		return out, err
	}
	copy(out[:], pk.Bytes())
	return out, nil
}

func toKey32(pk solana.PublicKey) [32]byte {
	var out [32]byte
	copy(out[:], pk.Bytes())
	return out
}
