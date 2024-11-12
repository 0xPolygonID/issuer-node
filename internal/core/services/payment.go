package services

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
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
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/eth"
	"github.com/polygonid/sh-id-platform/internal/log"
	"github.com/polygonid/sh-id-platform/internal/network"
)

type payment struct {
	networkResolver network.Resolver
}

// NewPaymentService creates a new payment service
func NewPaymentService(resolver network.Resolver) ports.PaymentService {
	return &payment{networkResolver: resolver}
}

// CreatePaymentRequest creates a payment request
func (p *payment) CreatePaymentRequest(ctx context.Context, issuerDID *w3c.DID, userDID *w3c.DID, signingKey string, credContext string, credType string) (*comm.BasicMessage, error) {
	id := uuid.New().String()
	basicMessage := comm.BasicMessage{
		From:     issuerDID.String(),
		To:       userDID.String(),
		Typ:      "application/iden3comm-plain-json",
		Type:     "https://iden3-communication.io/credentials/0.1/payment-reques",
		ID:       id,
		ThreadID: id,
	}

	//nolint:mnd
	randomBigInt, err := rand.Int(rand.Reader, big.NewInt(0).Exp(big.NewInt(2), big.NewInt(130), nil))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	oneHourLater := now.Add(1 * time.Hour)

	domain := protocol.Eip712Domain{
		Name:              "MCPayment",
		Version:           "1.0.0",
		ChainID:           "80002",
		VerifyingContract: "0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774",
	}

	privateKeyBytes, err := hex.DecodeString(signingKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	typedData := apitypes.TypedData{
		Types: apitypes.Types{
			"EIP712Domain": []apitypes.Type{
				{Name: "name", Type: "string"},
				{Name: "version", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
			"Iden3PaymentRailsRequestV1": []apitypes.Type{
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
			ChainId:           math.NewHexOrDecimal256(80002),               // nolint:mnd
			VerifyingContract: "0x380dd90852d3Fe75B4f08D0c47416D6c4E0dC774", // 2. config
		},
		Message: apitypes.TypedDataMessage{
			"recipient":      address, // 3. derive from PK
			"amount":         "100",   // 4. config per credential
			"expirationDate": fmt.Sprint(oneHourLater.Unix()),
			"nonce":          randomBigInt.String(),
			"metadata":       "0x",
		},
	}

	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return nil, err
	}

	signature, err := crypto.Sign(typedDataHash[:], privateKey)
	if err != nil {
		return nil, err
	}

	nativePayments := protocol.Iden3PaymentRailsRequestV1{
		Nonce: randomBigInt.String(),
		Type:  "Iden3PaymentRailsRequestV1",
		Context: protocol.NewPaymentContextStringCol([]string{
			"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
			"https://w3id.org/security/suites/eip712sig-2021/v1",
		}),
		Recipient:      address,
		Amount:         "100",
		ExpirationDate: fmt.Sprint(oneHourLater.Unix()),
		Metadata:       "0x",
		Currency:       "ETHWEI",
		Proof: protocol.NewPaymentProofEip712Signature([]protocol.EthereumEip712Signature2021{
			{
				Type:               "EthereumEip712Signature2021",
				ProofPurpose:       "assertionMethod",
				ProofValue:         hex.EncodeToString(signature),
				VerificationMethod: "did:pkh:eip155:80002:0xE9D7fCDf32dF4772A7EF7C24c76aB40E4A42274a#blockchainAccountId",
				Created:            now.Format(time.RFC3339),
				Eip712: protocol.Eip712Data{
					Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsERC20RequestV1.json",
					PrimaryType: "Iden3PaymentRailsRequestV1",
					Domain:      domain,
				},
			},
		}),
	}

	paymentRequestMessageBody := protocol.PaymentRequestMessageBody{
		Agent: "localhost",
		Payments: []protocol.PaymentRequestInfo{
			{
				Description: "Payment for credential",
				Data:        protocol.NewPaymentRequestInfoDataRails(nativePayments),
				Credentials: []protocol.PaymentRequestInfoCredentials{
					{
						Context: credContext,
						Type:    credType,
					},
				},
			},
		},
	}
	basicMessage.Body, err = json.Marshal(paymentRequestMessageBody)
	if err != nil {
		return nil, err
	}

	return &basicMessage, nil
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
