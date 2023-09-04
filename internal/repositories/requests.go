package repositories

import (
	"context"
	// "errors"
	"fmt"
	// "time"

	// "github.com/google/uuid"
	// core "github.com/iden3/go-iden3-core"
	// "github.com/jackc/pgtype"
	// "github.com/jackc/pgx/v4"

	"github.com/google/uuid"
	"github.com/polygonid/sh-id-platform/internal/core/domain"
	"github.com/polygonid/sh-id-platform/internal/core/ports"
	"github.com/polygonid/sh-id-platform/internal/db"
)

type requests struct{}

func NewRequests() ports.RequestRepository {
	return &requests{}
}

func (i *requests) Save(ctx context.Context, conn db.Querier, request *domain.Request) error {
	fmt.Println("Saving Request Info into DB...", request)
	_, err := conn.Exec(ctx, `INSERT INTO request_for_vc (id,user_id,issuer_id,schema_id,active) VALUES ($1,$2,$3,$4,$5)`, request.ID, request.User_id, request.Issuer_id, request.Schema_id, request.Active)
	return err
}

func (i *requests) GetByID(ctx context.Context, conn db.Querier, id uuid.UUID) {
	fmt.Println("Getting Dat from id ...", id)
	err := conn.QueryRow(ctx, `SELECT id,user_id,issuer_id,schema_id,active from request_for_vc WHERE id = $1`, id)

	fmt.Println("res", err)
	// return err;
}

// func (c *connections) Delete(ctx context.Context, conn db.Querier, id uuid.UUID, issuerDID core.DID) error {
// 	sql := `DELETE FROM connections WHERE id = $1 AND issuer_id = $2`
// 	cmd, err := conn.Exec(ctx, sql, id.String(), issuerDID.String())
// 	if err != nil {
// 		return err
// 	}

// 	if cmd.RowsAffected() == 0 {
// 		return ErrConnectionDoesNotExist
// 	}

// 	return nil
// }
