package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
)

// PaymentOption represents a payment option
type PaymentOption struct {
	ID          uuid.UUID
	IssuerDID   w3c.DID
	Name        string
	Description string
	Config      *PaymentOptionConfig
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewPaymentOption creates a new PaymentOption
func NewPaymentOption(issuerDID w3c.DID, name string, description string, config *PaymentOptionConfig) *PaymentOption {
	return &PaymentOption{
		ID:          uuid.New(),
		IssuerDID:   issuerDID,
		Name:        name,
		Description: description,
		Config:      config,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// PaymentOptionConfig represents the configuration of a payment option
type PaymentOptionConfig struct {
	Chains []PaymentOptionConfigChain `json:"Chains"`
}

// PaymentOptionConfigChain represents the configuration of a payment option chain
type PaymentOptionConfigChain struct {
	ChainId                         int                                                      `json:"ChainId"`
	Recipient                       string                                                   `json:"Recipient"`
	SigningKeyId                    string                                                   `json:"SigningKeyId"`
	Iden3PaymentRailsRequestV1      *PaymentOptionConfigChainIden3PaymentRailsRequestV1      `json:"Iden3PaymentRailsRequestV1"`
	Iden3PaymentRailsERC20RequestV1 *PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 `json:"Iden3PaymentRailsERC20RequestV1"`
}

// PaymentOptionConfigChainIden3PaymentRailsRequestV1 represents the configuration of a payment option rails
type PaymentOptionConfigChainIden3PaymentRailsRequestV1 struct {
	Amount   string `json:"Amount"`
	Currency string `json:"Currency"`
}

// PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 represents the configuration of a rails ERC20 payment option
type PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 struct {
	USDT struct {
		Amount string `json:"Amount"`
	} `json:"USDT"`
	USDC struct {
		Amount string `json:"Amount"`
	} `json:"USDC"`
}
