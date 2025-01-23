package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/jackc/pgtype"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
	"github.com/polygonid/sh-id-platform/internal/payments"
)

// ErrPaymentOptionDoesNotExists error
var ErrPaymentOptionDoesNotExists = errors.New("payment option not found")

// ErrPaymentOptionAlreadyExists error
var ErrPaymentOptionAlreadyExists = errors.New("payment option already exists")

// ErrPaymentRequestDoesNotExists error
var ErrPaymentRequestDoesNotExists = errors.New("payment request not found")

// payment repository
type payment struct {
	conn db.Storage
}

// NewPayment creates a new payment repository
func NewPayment(conn db.Storage) ports.PaymentRepository {
	return &payment{conn}
}

// SavePaymentRequest saves a payment request
func (p *payment) SavePaymentRequest(ctx context.Context, req *domain.PaymentRequest) (uuid.UUID, error) {
	const (
		insertPaymentRequest = `
INSERT 
INTO payment_requests (id, credentials, description, issuer_did, recipient_did, payment_option_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);`
		insertPaymentRequestItem = `
INSERT
INTO payment_request_items (id, nonce, payment_request_id, payment_option_id, payment_request_info, signing_key)
VALUES ($1, $2, $3, $4, $5, $6);`
	)

	tx, err := p.conn.Pgx.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, insertPaymentRequest,
		req.ID,
		req.Credentials,
		req.Description,
		req.IssuerDID.String(),
		req.UserDID.String(),
		req.PaymentOptionID,
		req.CreatedAt,
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert payment request: %w", err)
	}
	for _, item := range req.Payments {
		_, err = tx.Exec(ctx, insertPaymentRequestItem,
			item.ID,
			item.Nonce.String(),
			item.PaymentRequestID,
			item.PaymentOptionID,
			item.Payment,
			item.SigningKeyID,
		)
		if err != nil {
			return uuid.Nil, fmt.Errorf("could not insert payment request item: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, fmt.Errorf("could not commit transaction: %w", err)
	}
	return req.ID, nil
}

// GetPaymentRequestByID returns a payment request by ID
func (p *payment) GetPaymentRequestByID(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) (*domain.PaymentRequest, error) {
	const query = `
SELECT pr.id, pr.description, pr.credentials, pr.issuer_did, pr.recipient_did,  pr.payment_option_id, pr.created_at, pri.id, pri.nonce, pri.payment_request_id, pri.payment_request_info, pri.payment_option_id, pri.signing_key
FROM payment_requests pr
LEFT JOIN payment_request_items pri ON pr.id = pri.payment_request_id
WHERE pr.issuer_did = $1 AND pr.id = $2;`
	rows, err := p.conn.Pgx.Query(ctx, query, issuerDID.String(), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	requestFound := false
	var pr domain.PaymentRequest
	for rows.Next() {
		requestFound = true
		var item domain.PaymentRequestItem
		var strIssuerDID, strUserDID string
		var sNonce string
		var did *w3c.DID
		var paymentRequestInfoBytes []byte
		var paymentCredentials []byte
		if err := rows.Scan(
			&pr.ID,
			&pr.Description,
			&paymentCredentials,
			&strIssuerDID,
			&strUserDID,
			&pr.PaymentOptionID,
			&pr.CreatedAt,
			&item.ID,
			&sNonce,
			&item.PaymentRequestID,
			&paymentRequestInfoBytes,
			&item.PaymentOptionID,
			&item.SigningKeyID,
		); err != nil {
			return nil, fmt.Errorf("could not scan payment request: %w", err)
		}

		item.Payment, err = p.paymentRequestItem(paymentRequestInfoBytes)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal payment request info: %w", err)
		}

		const base10 = 10
		nonce, ok := new(big.Int).SetString(sNonce, base10)
		if !ok {
			return nil, fmt.Errorf("could not parse nonce: %w", err)
		}
		item.Nonce = *nonce
		if did, err = w3c.ParseDID(strIssuerDID); err != nil {
			return nil, fmt.Errorf("could not parse issuer DID: %w", err)
		}
		pr.IssuerDID = *did
		if did, err = w3c.ParseDID(strUserDID); err != nil {
			return nil, fmt.Errorf("could not parse recipient DID: %w", err)
		}
		pr.UserDID = *did
		pr.Credentials, err = p.paymentRequestCredentials(paymentCredentials)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal payment credentials info: %w", err)
		}

		pr.Payments = append(pr.Payments, item)
	}

	if !requestFound {
		return nil, ErrPaymentRequestDoesNotExists
	}
	return &pr, nil
}

// DeletePaymentRequest deletes a payment request
func (p *payment) DeletePaymentRequest(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) error {
	const removePaymentRequestItemsQuery = `DELETE FROM payment_request_items WHERE payment_request_id = $1;`
	_, err := p.conn.Pgx.Exec(ctx, removePaymentRequestItemsQuery, id)
	if err != nil {
		return err
	}

	const query = `DELETE FROM payment_requests WHERE id = $1 and issuer_did = $2;`
	cmd, err := p.conn.Pgx.Exec(ctx, query, id, issuerDID.String())
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrPaymentRequestDoesNotExists
	}
	return nil
}

// GetAllPaymentRequests returns all payment requests
func (p *payment) GetAllPaymentRequests(ctx context.Context, issuerDID w3c.DID) ([]domain.PaymentRequest, error) {
	const query = `
SELECT pr.id, 
    pr.description, 
    pr.credentials, 
    pr.issuer_did, 
    pr.recipient_did,  
    pr.payment_option_id, 
    pr.created_at, 
    COALESCE(
        JSON_AGG(
            JSON_BUILD_OBJECT(
                'id', pri.id,
                'nc', pri.nonce,
                'rid', pri.payment_request_id,
                'rnfo', pri.payment_request_info,
                'optid', pri.payment_option_id,
                'sk', pri.signing_key
            )
        ) FILTER (WHERE pri.id IS NOT NULL),
        '[]'
    ) AS payment_request_items
FROM payment_requests pr
LEFT JOIN payment_request_items pri ON pr.id = pri.payment_request_id
WHERE pr.issuer_did = $1
GROUP BY pr.id, pr.description, pr.credentials, pr.issuer_did, pr.recipient_did, pr.payment_option_id, pr.created_at
`
	rows, err := p.conn.Pgx.Query(ctx, query, issuerDID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pr domain.PaymentRequest
	var requests []domain.PaymentRequest
	for rows.Next() {
		var strIssuerDID, strUserDID string
		var did *w3c.DID
		var paymentCredentials []byte
		var requestItems pgtype.JSON
		if err := rows.Scan(
			&pr.ID,
			&pr.Description,
			&paymentCredentials,
			&strIssuerDID,
			&strUserDID,
			&pr.PaymentOptionID,
			&pr.CreatedAt,
			&requestItems,
		); err != nil {
			return nil, fmt.Errorf("could not scan payment request: %w", err)
		}
		var itemDtoCol []struct {
			ID                 string                          `json:"id"`
			Nonce              int64                           `json:"nc"`
			PaymentRequestID   string                          `json:"rid"`
			PaymentRequestInfo protocol.PaymentRequestInfoData `json:"rnfo"`
			PaymentOptionID    int                             `json:"optid"`
			SigningKey         string                          `json:"sk"`
		}
		if err := requestItems.AssignTo(&itemDtoCol); err != nil {
			return nil, fmt.Errorf("could not assign to payment request items: %w", err)
		}
		pr.Payments = make([]domain.PaymentRequestItem, len(itemDtoCol))
		for i, itemDto := range itemDtoCol {
			pr.Payments[i].ID, err = uuid.Parse(itemDto.ID)
			if err != nil {
				return nil, fmt.Errorf("could not parse payment request item ID: %w", err)
			}
			pr.Payments[i].PaymentRequestID, err = uuid.Parse(itemDto.PaymentRequestID)
			if err != nil {
				return nil, fmt.Errorf("could not parse payment request ID: %w", err)
			}
			pr.Payments[i].PaymentOptionID = payments.OptionConfigIDType(itemDto.PaymentOptionID)
			pr.Payments[i].SigningKeyID = itemDto.SigningKey
			pr.Payments[i].Payment = itemDto.PaymentRequestInfo[0]
			nonce := big.NewInt(itemDto.Nonce)
			pr.Payments[i].Nonce = *nonce
		}
		if did, err = w3c.ParseDID(strIssuerDID); err != nil {
			return nil, fmt.Errorf("could not parse issuer DID: %w", err)
		}
		pr.IssuerDID = *did
		if did, err = w3c.ParseDID(strUserDID); err != nil {
			return nil, fmt.Errorf("could not parse recipient DID: %w", err)
		}
		pr.UserDID = *did
		pr.Credentials, err = p.paymentRequestCredentials(paymentCredentials)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal payment credentials info: %w", err)
		}

		requests = append(requests, pr)
	}
	return requests, nil
}

// GetPaymentRequestItem returns a payment request item
func (p *payment) GetPaymentRequestItem(ctx context.Context, issuerDID w3c.DID, nonce *big.Int) (*domain.PaymentRequestItem, error) {
	const query = `
SELECT payment_request_items.id, nonce, payment_request_id, payment_request_info, payment_request_items.payment_option_id, payment_request_items.signing_key
FROM payment_request_items 
LEFT JOIN payment_requests ON payment_requests.id = payment_request_items.payment_request_id
WHERE payment_requests.issuer_did = $1 AND nonce = $2;`
	var item domain.PaymentRequestItem
	var sNonce string
	var paymentRequestInfoBytes []byte
	err := p.conn.Pgx.QueryRow(ctx, query, issuerDID.String(), nonce.String()).Scan(
		&item.ID,
		&sNonce,
		&item.PaymentRequestID,
		&paymentRequestInfoBytes,
		&item.PaymentOptionID,
		&item.SigningKeyID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not get payment request item: %w", err)
	}
	item.Payment, err = p.paymentRequestItem(paymentRequestInfoBytes)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal payment request info: %w", err)
	}
	const base10 = 10
	nonceInt, ok := new(big.Int).SetString(sNonce, base10)
	if !ok {
		return nil, fmt.Errorf("could not parse nonce: %w", err)
	}
	item.Nonce = *nonceInt
	return &item, nil
}

// SavePaymentOption saves a payment option
func (p *payment) SavePaymentOption(ctx context.Context, opt *domain.PaymentOption) (uuid.UUID, error) {
	const query = `
		INSERT INTO payment_options (id, issuer_did, name, description, configuration, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
 		ON CONFLICT (id) DO UPDATE SET name=$3, description=$4, configuration=$5, updated_at=NOW()
		RETURNING id;
		`

	_, err := p.conn.Pgx.Exec(ctx, query, opt.ID, opt.IssuerDID.String(), opt.Name, opt.Description, opt.Config, opt.CreatedAt, opt.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return uuid.Nil, ErrIdentityNotFound
		}
		if strings.Contains(err.Error(), "violates unique constraint") {
			return uuid.Nil, ErrPaymentOptionAlreadyExists
		}
		return uuid.Nil, err
	}
	return opt.ID, nil
}

// GetAllPaymentOptions returns all payment options
func (p *payment) GetAllPaymentOptions(ctx context.Context, issuerDID w3c.DID) ([]domain.PaymentOption, error) {
	const query = `
SELECT id, issuer_did, name, description, configuration, created_at, updated_at 
FROM payment_options
WHERE issuer_did=$1
ORDER BY created_at DESC;`

	rows, err := p.conn.Pgx.Query(ctx, query, issuerDID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var opts []domain.PaymentOption
	for rows.Next() {
		var opt domain.PaymentOption
		var strIssuerDID string
		err := rows.Scan(&opt.ID, &strIssuerDID, &opt.Name, &opt.Description, &opt.Config, &opt.CreatedAt, &opt.UpdatedAt)
		if err != nil {
			return nil, err
		}
		did, err := w3c.ParseDID(strIssuerDID)
		if err != nil {
			return nil, fmt.Errorf("could not parse issuer DID: %w", err)
		}
		opt.IssuerDID = *did
		opts = append(opts, opt)
	}
	return opts, nil
}

// GetPaymentOptionByID returns a payment option by ID
func (p *payment) GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error) {
	const baseQuery = `
SELECT id, issuer_did, name, description, configuration, created_at, updated_at 
FROM payment_options 
WHERE id = $1
`
	var opt domain.PaymentOption
	var strIssuerDID string

	query := baseQuery
	queryParams := []interface{}{id}
	if issuerDID != nil {
		query += ` AND issuer_did = $2`
		queryParams = append(queryParams, issuerDID.String())
	}

	err := p.conn.Pgx.QueryRow(ctx, query, queryParams...).Scan(&opt.ID, &strIssuerDID, &opt.Name, &opt.Description, &opt.Config, &opt.CreatedAt, &opt.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, ErrPaymentOptionDoesNotExists
		}
		return nil, err
	}
	did, err := w3c.ParseDID(strIssuerDID)
	if err != nil {
		return nil, fmt.Errorf("could not parse issuer DID: %w", err)
	}
	opt.IssuerDID = *did
	return &opt, nil
}

// DeletePaymentOption deletes a payment option
func (p *payment) DeletePaymentOption(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) error {
	const query = `DELETE FROM payment_options WHERE id = $1 and issuer_did = $2;`

	cmd, err := p.conn.Pgx.Exec(ctx, query, id, issuerDID.String())
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return ErrPaymentOptionDoesNotExists
	}
	return nil
}

// paymentRequestItem extracts the payment request item from the payload
// It uses an intermediate structure of type protocol.PaymentRequestInfoData
// to unmarshal the payment request info.
// This is necessary because PaymentRequestInfoDataItem is an interface and the unmarshal fails
func (p *payment) paymentRequestItem(payload []byte) (protocol.PaymentRequestInfoDataItem, error) {
	var data protocol.PaymentRequestInfoData
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("could not unmarshal payment request info: %w", err)
	}
	if len(data) == 0 {
		return nil, errors.New("payment request info is empty")
	}
	return data[0], nil
}

func (p *payment) paymentRequestCredentials(credentials []byte) ([]protocol.PaymentRequestInfoCredentials, error) {
	var data []protocol.PaymentRequestInfoCredentials
	if err := json.Unmarshal(credentials, &data); err != nil {
		return nil, fmt.Errorf("could not unmarshal payment request info: %w", err)
	}
	return data, nil
}
