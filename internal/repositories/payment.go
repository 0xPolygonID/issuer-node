package repositories

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrPaymentOptionDoesNotExists error
var ErrPaymentOptionDoesNotExists = errors.New("payment option not found")

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
INTO payment_requests (id, issuer_did, recipient_did, thread_id, payment_option_id, created_at)
VALUES ($1, $2, $3, $4, $5, $6);`
		insertPaymentRequestItem = `
INSERT
INTO payment_request_items (id, nonce, payment_request_id, payment_request_info)
VALUES ($1, $2, $3, $4);`
	)

	tx, err := p.conn.Pgx.Begin(ctx)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not start transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	_, err = tx.Exec(ctx, insertPaymentRequest, req.ID, req.IssuerDID.String(), req.RecipientDID.String(), req.ThreadID, req.PaymentOptionID, req.CreatedAt)
	if err != nil {
		return uuid.Nil, fmt.Errorf("could not insert payment request: %w", err)
	}
	for _, item := range req.Payments {
		_, err = tx.Exec(ctx, insertPaymentRequestItem, item.ID, item.Nonce.String(), item.PaymentRequestID, item.Payment)
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
SELECT pr.id, pr.issuer_did, pr.recipient_did, pr.thread_id, pr.payment_option_id, pr.created_at, pri.id, pri.nonce, pri.payment_request_id, pri.payment_request_info
FROM payment_requests pr
LEFT JOIN payment_request_items pri ON pr.id = pri.payment_request_id
WHERE pr.issuer_did = $1 AND pr.id = $2;`
	rows, err := p.conn.Pgx.Query(ctx, query, issuerDID.String(), id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pr domain.PaymentRequest
	for rows.Next() {
		var item domain.PaymentRequestItem
		var strIssuerDID, strRecipientDID string
		var sNonce string
		var did *w3c.DID
		if err := rows.Scan(
			&pr.ID,
			&strIssuerDID,
			&strRecipientDID,
			&pr.ThreadID,
			&pr.PaymentOptionID,
			&pr.CreatedAt,
			&item.ID,
			&sNonce,
			&item.PaymentRequestID,
			&item.Payment,
		); err != nil {
			return nil, fmt.Errorf("could not scan payment request: %w", err)
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
		if did, err = w3c.ParseDID(strRecipientDID); err != nil {
			return nil, fmt.Errorf("could not parse recipient DID: %w", err)
		}
		pr.RecipientDID = *did
		pr.Payments = append(pr.Payments, item)
	}
	return &pr, nil
}

// GetAllPaymentRequests returns all payment requests
// TODO: Pagination?
func (p *payment) GetAllPaymentRequests(ctx context.Context, issuerDID w3c.DID) ([]domain.PaymentRequest, error) {
	// TODO implement me
	panic("implement me")
}

// GetPaymentRequestItem returns a payment request item
func (p *payment) GetPaymentRequestItem(ctx context.Context, issuerDID w3c.DID, nonce *big.Int) (*domain.PaymentRequestItem, error) {
	const query = `
SELECT id, nonce, payment_request_id, payment_request_info
FROM payment_request_items
LEFT JOIN payment_requests ON payment_requests.id = payment_request_items.payment_request_id
WHERE payment_requests.issuer_did = $1 AND nonce = $2;`
	var item domain.PaymentRequestItem
	err := p.conn.Pgx.QueryRow(ctx, query, issuerDID.String(), nonce).Scan(&item.ID, &item.Nonce, &item.PaymentRequestID, &item.Payment)
	if err != nil {
		return nil, fmt.Errorf("could not get payment request item: %w", err)
	}
	return &item, nil
}

// SavePaymentOption saves a payment option
func (p *payment) SavePaymentOption(ctx context.Context, opt *domain.PaymentOption) (uuid.UUID, error) {
	const query = `
INSERT INTO payment_options (id, issuer_did, name, description, configuration, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);
`

	_, err := p.conn.Pgx.Exec(ctx, query, opt.ID, opt.IssuerDID.String(), opt.Name, opt.Description, opt.Config, opt.CreatedAt, opt.UpdatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "violates foreign key constraint") {
			return uuid.Nil, ErrIdentityNotFound
		}
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
