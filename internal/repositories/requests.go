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

type Requestresponse struct {
	id        uuid.UUID
	schema_id string
	user_id   string
	issuer_id string
	active    bool
}

func NewRequests() ports.RequestRepository {
	return &requests{}
}

func (i *requests) Save(ctx context.Context, conn db.Querier, request *domain.Request) error {
	fmt.Println("Saving Request Info into DB...", request)
	_, err := conn.Exec(ctx, `INSERT INTO requests_for_vc (id,user_id,issuer_id,schema_id,active) VALUES ($1,$2,$3,$4,$5)`, request.ID, request.User_id, request.Issuer_id, request.Schema_id, request.Active)
	return err
}

func (i *requests) GetByID(ctx context.Context, conn db.Querier, id uuid.UUID) domain.Responce {
	fmt.Println("Getting Dat from id ...", id)
	response := Requestresponse{}
	res := conn.QueryRow(ctx, `SELECT id, user_id ,issuer_id, schema_id, active from requests_for_vc WHERE id = $1`, id).Scan(
		&response.id,
		&response.user_id,
		&response.issuer_id,
		&response.schema_id,
		&response.active,
	)
	fmt.Println("response", response)
	fmt.Println("res", res)
	// si := res.Scan("schema_id")
	// return err;
	return domain.Responce{Id: response.id, SchemaID: response.schema_id, UserDID: response.issuer_id, Issuer_id: response.user_id, Active: response.active}
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
