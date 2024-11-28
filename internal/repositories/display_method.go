package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"

	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

var (
	// DisplayMethodNotFoundErr is the error returned when the display method is not found
	DisplayMethodNotFoundErr = errors.New("display method not found")
	// DisplayMethodDuplicateNameError is the error returned when trying to save a display method with the same name
	DisplayMethodDuplicateNameError = errors.New("display method with the same name already exists")
)

// DisplayMethod represents the display method repository
type DisplayMethod struct {
	conn db.Storage
}

// NewDisplayMethod creates a new display method repository
func NewDisplayMethod(conn db.Storage) ports.DisplayMethodRepository {
	return &DisplayMethod{
		conn,
	}
}

// Save stores in the database the given display method and updates the modified at in case already exists
func (d DisplayMethod) Save(ctx context.Context, displayMethod domain.DisplayMethod) (*uuid.UUID, error) {
	var id uuid.UUID
	sql := `INSERT INTO display_methods (id, name, url, issuer_did)
			VALUES($1, $2, $3, $4) ON CONFLICT (id) DO
			UPDATE SET name=$2, url=$3, issuer_did=$4
			RETURNING id`
	err := d.conn.Pgx.QueryRow(ctx, sql, displayMethod.ID, displayMethod.Name, displayMethod.URL, displayMethod.IssuerCoreDID().String()).Scan(&id)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			return nil, DisplayMethodDuplicateNameError
		}
		return nil, err
	}
	return &id, err
}

// GetAll returns all display methods for the given identity
func (d DisplayMethod) GetAll(ctx context.Context, identityDID w3c.DID, filter ports.DisplayMethodFilter) ([]domain.DisplayMethod, uint, error) {
	var displayMethods []domain.DisplayMethod
	sql := `SELECT id, name, url, issuer_did FROM display_methods WHERE issuer_did=$1`

	orderStr := " ORDER BY created_at DESC"
	if len(filter.OrderBy) > 0 {
		orderStr = " ORDER BY " + filter.OrderBy.String()
	}
	sql += orderStr

	sqlArgs := make([]interface{}, 0)
	sqlArgs = append(sqlArgs, identityDID.String())
	var count uint
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM (SELECT id FROM (%s) as innerquery) as count", sql)
	if err := d.conn.Pgx.QueryRow(ctx, countQuery, sqlArgs...).Scan(&count); err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, 0, nil
		}
		return nil, 0, err
	}

	sql += fmt.Sprintf(" OFFSET %d LIMIT %d;", (filter.Page-1)*filter.MaxResults, filter.MaxResults)
	rows, err := d.conn.Pgx.Query(ctx, sql, identityDID.String())
	if err != nil {
		return nil, 0, err
	}
	for rows.Next() {
		var displayMethod domain.DisplayMethod
		err = rows.Scan(&displayMethod.ID, &displayMethod.Name, &displayMethod.URL, &displayMethod.IssuerDID)
		if err != nil {
			return nil, 0, err
		}
		displayMethods = append(displayMethods, displayMethod)
	}
	return displayMethods, count, nil
}

// GetByID returns the display method with the given id
func (d DisplayMethod) GetByID(ctx context.Context, identityDID w3c.DID, id uuid.UUID) (*domain.DisplayMethod, error) {
	var displayMethod domain.DisplayMethod
	sql := `SELECT id, name, url, issuer_did FROM display_methods WHERE issuer_did=$1 and id=$2`
	err := d.conn.Pgx.QueryRow(ctx, sql, identityDID.String(), id).Scan(&displayMethod.ID, &displayMethod.Name, &displayMethod.URL, &displayMethod.IssuerDID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			return nil, DisplayMethodNotFoundErr
		}
		return nil, err
	}
	return &displayMethod, nil
}

// Delete removes the display method with the given id
func (d DisplayMethod) Delete(ctx context.Context, identityDID w3c.DID, id uuid.UUID) error {
	sql := `DELETE FROM display_methods WHERE issuer_did=$1 AND id=$2`
	tag, err := d.conn.Pgx.Exec(ctx, sql, identityDID.String(), id)
	if tag.RowsAffected() == 0 {
		return DisplayMethodNotFoundErr
	}
	return err
}
