package domain

import (
	"encoding/json"
	"errors"
	"math/big"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
)

// PaymentRequest represents a payment request
type PaymentRequest struct {
	ID              uuid.UUID
	IssuerDID       w3c.DID
	RecipientDID    w3c.DID
	ThreadID        string
	PaymentOptionID uuid.UUID
	Payments        []PaymentRequestItem
	CreatedAt       time.Time
}

// PaymentRequestItem represents a payment request item
type PaymentRequestItem struct {
	ID               uuid.UUID
	Nonce            big.Int
	PaymentRequestID uuid.UUID
	Payment          protocol.Payment
}

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
	ChainId      int    `json:"ChainId"`
	Recipient    string `json:"Recipient"`
	SigningKeyId string `json:"SigningKeyId"`

	// TODO: remove field SigningKeyOpt once we have key management.
	// Temporary solution until we have key management
	// WARNING: Only for testing purposes. This is not secure as SigningKey will be stored in the database
	// USE AT YOUR OWN RISK
	SigningKeyOpt *string `json:"SigningKeyOpt"`

	Iden3PaymentRailsRequestV1      *PaymentOptionConfigChainIden3PaymentRailsRequestV1      `json:"Iden3PaymentRailsRequestV1"`
	Iden3PaymentRailsERC20RequestV1 *PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 `json:"Iden3PaymentRailsERC20RequestV1"`
}

// PaymentOptionConfigChainIden3PaymentRailsRequestV1 represents the configuration of a payment option rails
type PaymentOptionConfigChainIden3PaymentRailsRequestV1 struct {
	Amount   float64 `json:"Amount"`
	Currency string  `json:"Currency"`
}

// UnmarshalJSON unmarshals a PaymentOptionConfigChainIden3PaymentRailsRequestV1
// so we can accept amounts with string or float64 types
func (p *PaymentOptionConfigChainIden3PaymentRailsRequestV1) UnmarshalJSON(data []byte) error {
	var temp struct {
		Amount   any    `json:"Amount"`
		Currency string `json:"Currency"`
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}

	switch temp.Amount.(type) {
	case float64:
		p.Amount = temp.Amount.(float64) // nolint: forcetypeassert
	case string:
		amountFloat, err := strconv.ParseFloat(temp.Amount.(string), 64)
		if err != nil {
			return err
		}
		p.Amount = amountFloat
	default:
		return errors.New("unexpected format of amount")
	}

	p.Currency = temp.Currency

	return nil
}

// PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 represents the configuration of a rails ERC20 payment option
type PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1 struct {
	USDT struct {
		Amount float64 `json:"Amount"`
	} `json:"USDT"`
	USDC struct {
		Amount float64 `json:"Amount"`
	} `json:"USDC"`
}

// UnmarshalJSON unmarshals a PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1
// so we can accept amounts with string or float64 types
func (p *PaymentOptionConfigChainIden3PaymentRailsERC20RequestV1) UnmarshalJSON(data []byte) error {
	var temp struct {
		USDT struct {
			Amount any `json:"Amount"`
		}
		USDC struct {
			Amount any `json:"Amount"`
		}
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		return err
	}
	switch temp.USDT.Amount.(type) {
	case float64:
		p.USDT.Amount = temp.USDT.Amount.(float64) // nolint: forcetypeassert
	case string:
		amountFloat, err := strconv.ParseFloat(temp.USDT.Amount.(string), 64)
		if err != nil {
			return err
		}
		p.USDT.Amount = amountFloat
	}
	switch temp.USDC.Amount.(type) {
	case float64:
		p.USDC.Amount = temp.USDC.Amount.(float64) // nolint: forcetypeassert
	case string:
		amountFloat, err := strconv.ParseFloat(temp.USDC.Amount.(string), 64)
		if err != nil {
			return err
		}
		p.USDC.Amount = amountFloat
	}
	return nil
}
