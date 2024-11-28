package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/db"
)

// ErrPaymentOptionDoesNotExists error
var ErrPaymentOptionDoesNotExists = errors.New("payment option not found")

// Payment repository
type Payment struct {
	conn db.Storage
}

// NewPayment creates a new Payment repository
func NewPayment(conn db.Storage) *Payment {
	return &Payment{conn}
}

// SavePaymentOption saves a payment option
func (p *Payment) SavePaymentOption(ctx context.Context, opt *domain.PaymentOption) (uuid.UUID, error) {
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
func (p *Payment) GetAllPaymentOptions(ctx context.Context, issuerDID w3c.DID) ([]domain.PaymentOption, error) {
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
func (p *Payment) GetPaymentOptionByID(ctx context.Context, issuerDID *w3c.DID, id uuid.UUID) (*domain.PaymentOption, error) {
	const baseQuery = `
SELECT id, issuer_did, name, description, configuration, created_at, updated_at 
FROM payment_options 
WHERE id = $1
`
	var opt domain.PaymentOption
	var strIssuerDID string

	query := baseQuery
	if issuerDID != nil {
		query += ` AND issuer_did = $2`
	}
	err := p.conn.Pgx.QueryRow(ctx, query, id, issuerDID.String()).Scan(&opt.ID, &strIssuerDID, &opt.Name, &opt.Description, &opt.Config, &opt.CreatedAt, &opt.UpdatedAt)
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
func (p *Payment) DeletePaymentOption(ctx context.Context, issuerDID w3c.DID, id uuid.UUID) error {
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
