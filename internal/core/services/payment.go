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
	"github.com/iden3/go-schema-processor/v2/verifiable"
	comm "github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/protocol"

	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/eth"
	"github.com/polygonid/sh-id-platform/internal/network"
)

type payment struct {
	networkResolver network.Resolver
}

// NewPaymentService creates a new payment service
func NewPaymentService(resolver network.Resolver) ports.PaymentService {
	return &payment{networkResolver: resolver}
}

// PaymentRequestMessageBody TODO: use ones from iden3comm
type PaymentRequestMessageBody struct {
	Agent    string               `json:"agent"`
	Payments []PaymentRequestInfo `json:"payments"`
}

// PaymentRequestInfo TODO: use ones from iden3comm
type PaymentRequestInfo struct {
	Type        *string                         `json:"type,omitempty"`
	Credentials []PaymentRequestInfoCredentials `json:"credentials"`
	Description string                          `json:"description"`
	Data        interface{}                     `json:"data"`
}

// PaymentRequestInfoCredentials TODO: use ones from iden3comm
type PaymentRequestInfoCredentials struct {
	Context string `json:"context,omitempty"`
	Type    string `json:"type,omitempty"`
}

// EthereumEip712Signature2021 TODO: use ones from iden3comm
type EthereumEip712Signature2021 struct {
	Type               verifiable.ProofType `json:"type"`
	ProofPurpose       string               `json:"proofPurpose"`
	ProofValue         string               `json:"proofValue"`
	VerificationMethod string               `json:"verificationMethod"`
	Created            string               `json:"created"`
	Eip712             Eip712Data           `json:"eip712"`
}

// Eip712Data TODO: use ones from iden3comm
type Eip712Data struct {
	Types       string       `json:"types"`
	PrimaryType string       `json:"primaryType"`
	Domain      Eip712Domain `json:"domain"`
}

// Eip712Domain TODO: use ones from iden3comm
type Eip712Domain struct {
	Name              string `json:"name"`
	Version           string `json:"version"`
	ChainID           string `json:"chainId"`
	VerifyingContract string `json:"verifyingContract"`
}

// Iden3PaymentRailsRequestV1 TODO: use ones from iden3comm
type Iden3PaymentRailsRequestV1 struct {
	Nonce          string                        `json:"nonce"`
	Type           string                        `json:"type"`
	Context        []string                      `json:"@context"`
	Recipient      string                        `json:"recipient"`
	Amount         string                        `json:"amount"` // Not negative number
	ExpirationDate string                        `json:"expirationDate"`
	Proof          []EthereumEip712Signature2021 `json:"proof"`
	Metadata       string                        `json:"metadata"`
	Currency       string                        `json:"currency"`
}

// Iden3PaymentRailsERC20RequestV1 TODO: use ones from iden3comm
type Iden3PaymentRailsERC20RequestV1 struct {
	Nonce          string                        `json:"nonce"`
	Type           string                        `json:"type"`
	Context        []string                      `json:"@context"`
	Recipient      string                        `json:"recipient"`
	Amount         string                        `json:"amount"` // Not negative number
	ExpirationDate string                        `json:"expirationDate"`
	Proof          []EthereumEip712Signature2021 `json:"proof"`
	Metadata       string                        `json:"metadata"`
	Currency       string                        `json:"currency"`
	TokenAddress   string                        `json:"tokenAddress"`
	Features       []string                      `json:"features,omitempty"`
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

	domain := Eip712Domain{
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
		return nil, fmt.Errorf("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
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

	nativePayments := Iden3PaymentRailsRequestV1{
		Nonce: randomBigInt.String(),
		Type:  "Iden3PaymentRailsRequestV1",
		Context: []string{
			"https://schema.iden3.io/core/jsonld/payment.jsonld#Iden3PaymentRailsERC20RequestV1",
			"https://w3id.org/security/suites/eip712sig-2021/v1",
		},
		Recipient:      address,
		Amount:         "100",
		ExpirationDate: fmt.Sprint(oneHourLater.Unix()),
		Metadata:       "0x",
		Currency:       "ETHWEI",
		Proof: []EthereumEip712Signature2021{
			{
				Type:               "EthereumEip712Signature2021",
				ProofPurpose:       "assertionMethod",
				ProofValue:         hex.EncodeToString(signature),
				VerificationMethod: "did:pkh:eip155:80002:0xE9D7fCDf32dF4772A7EF7C24c76aB40E4A42274a#blockchainAccountId",
				Created:            now.Format(time.RFC3339),
				Eip712: Eip712Data{
					Types:       "https://schema.iden3.io/core/json/Iden3PaymentRailsERC20RequestV1.json",
					PrimaryType: "Iden3PaymentRailsRequestV1",
					Domain:      domain,
				},
			},
		},
	}

	paymentRequestMessageBody := PaymentRequestMessageBody{
		Agent: "localhost",
		Payments: []PaymentRequestInfo{
			{
				Description: "Payment for credential",
				Data:        nativePayments,
				Credentials: []PaymentRequestInfoCredentials{
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

// Iden3PaymentRailsV1 TODO: Use ones from iden3comm
type Iden3PaymentRailsV1 struct {
	Nonce       string   `json:"nonce"`
	Type        string   `json:"type"`
	Context     []string `json:"@context,omitempty"`
	PaymentData struct {
		TxID    string `json:"txId"`
		ChainID string `json:"chainId"`
	} `json:"paymentData"`
}

// Iden3PaymentRailsV1Body TODO: Use ones from iden3comm
type Iden3PaymentRailsV1Body struct {
	Payments []Iden3PaymentRailsV1 `json:"payments"`
}

func (p *payment) VerifyPayment(ctx context.Context, recipient common.Address, message comm.BasicMessage) (bool, error) {
	const base10 = 10
	var paymentRequest Iden3PaymentRailsV1Body
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

	nonce, _ := new(big.Int).SetString(paymentRequest.Payments[0].Nonce, base10)
	isPaid, err := instance.IsPaymentDone(&bind.CallOpts{Context: context.Background()}, recipient, nonce)
	if err != nil {
		return false, err
	}
	return isPaid, nil
}
