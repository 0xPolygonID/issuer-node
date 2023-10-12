package repositories

import (
	"context"
	"errors"
	"time"

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
	id              uuid.UUID
	schema_id       uuid.UUID
	user_id         string
	issuer_id       string
	active          bool
	credential_type string
	request_type    string
	role_type       string
	proof_type	  string
	proof_id	  string
	age 		   string
	request_status  string
	verifier_status string
	wallet_status   string
	source 			string
	created_at      time.Time
	modifed_at      time.Time
}

type NotificationReponse struct{
	id              uuid.UUID
	user_id         string
	module          string
	notification_type string
	notification_title string
	notification_message string
	created_at      time.Time
}



var ErrNoRequestExist = errors.New("Request does not exist")

func NewRequests() ports.RequestRepository {
	return &requests{}
}

func (i *requests) Save(ctx context.Context, conn db.Querier, request *domain.Request) error {
	fmt.Println("Saving Request Info into DB...", request)
	_, err := conn.Exec(ctx, `INSERT INTO requests_for_vc (id,UDID,issuer_id,schema_id,credential_type,request_type,role_type,proof_type,proof_id,age,active,request_status,verifier_status,wallet_status,source) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`, 
	request.ID, 
	request.User_id,
	request.Issuer_id,
	request.Schema_id,
	request.CredentialType,
	request.Type,
	request.RoleType,
	request.ProofType,
	request.ProofId,
	request.Age,
	request.Active, 
	request.Status, 
	request.Verify_Status, 
	request.Wallet_Status,
	request.Source)
	return err
}

func (i *requests) GetByID(ctx context.Context, conn db.Querier, id uuid.UUID) (domain.Responce, error) {
	fmt.Println("Getting Dat from id ...", id)
	response := Requestresponse{}
	var errres domain.Responce
	res := conn.QueryRow(ctx, `SELECT id,UDID,issuer_id,schema_id,credential_type,request_type,role_type,proof_type,proof_id,age,active,request_status,verifier_status,wallet_status,source,created_at,modified_at from requests_for_vc WHERE id = $1`, id).Scan(
		&response.id,
		&response.user_id,
		&response.issuer_id,
		&response.schema_id,
		&response.credential_type,
		&response.request_type,
		&response.role_type,
		&response.proof_type,
		&response.proof_id,
		&response.age,
		&response.active,
		&response.request_status,
		&response.verifier_status,
		&response.wallet_status,
		&response.source,
		&response.created_at,
		&response.modifed_at,
	)
	if res != nil {
		fmt.Println("ERR", res)
		return errres, ErrNoRequestExist
	}
	fmt.Println("response", response)
	// si := res.Scan("schema_id")
	// return err;
	return domain.Responce{
		Id: response.id,
		UserDID: response.user_id,
		Issuer_id: response.issuer_id,
		SchemaID: response.schema_id,
		CredentialType: response.credential_type,
		RequestType:response.request_type,
		RoleType: response.role_type,
		ProofType: response.proof_type,
		ProofId: response.proof_id,
		Age: response.age,
		RequestStatus: response.request_status,
		WalletStatus: response.wallet_status,
		VerifyStatus: response.verifier_status,
		Active: response.active,
		Source: response.source,
		CreatedAt: response.created_at,
		ModifiedAt: response.modifed_at,
		}, nil
}

func (i *requests) Get(ctx context.Context, conn db.Querier) ([]*domain.Responce, error) {
	fmt.Println("Getting Dat from id ...")
	response := Requestresponse{}
	domainResponce := make([]*domain.Responce, 0)
	rows, err := conn.Query(ctx, `SELECT id,UDID,issuer_id,schema_id,credential_type,request_type,role_type,proof_type,proof_id,age,active,request_status,verifier_status,wallet_status,source,created_at,modified_at from requests_for_vc`)

	fmt.Println("Get Rquests", rows)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		fmt.Println("Get Rquests in loop", rows)
		if err := rows.Scan(
			&response.id,
			&response.user_id,
			&response.issuer_id,
			&response.schema_id,
			&response.credential_type,
			&response.request_type,
			&response.role_type,
			&response.proof_type,
			&response.proof_id,
			&response.age,
			&response.active,
			&response.request_status,
			&response.verifier_status,
			&response.wallet_status,
			&response.source,
			&response.created_at,
			&response.modifed_at,
		); err != nil {
			return nil, err
		}
		temp := &domain.Responce{
			Id: response.id,
			UserDID: response.user_id,
			Issuer_id: response.issuer_id,
			SchemaID: response.schema_id,
			CredentialType: response.credential_type,
			RequestType:response.request_type,
			RoleType: response.role_type,
			ProofType: response.proof_type,
			ProofId: response.proof_id,
			Age: response.age,
			RequestStatus: response.request_status,
			WalletStatus: response.wallet_status,
			VerifyStatus: response.verifier_status,
			Active: response.active,
			Source: response.source,
			CreatedAt: response.created_at,
			ModifiedAt: response.modifed_at,
			}

		domainResponce = append(domainResponce, temp)
	}
	fmt.Println("response", response)
	// si := res.Scan("schema_id")
	// return err;
	return domainResponce, nil
}

func (i *requests) UpdateStatus(ctx context.Context, conn db.Querier, id uuid.UUID) (int64, error) {
	query := "UPDATE requests_for_vc SET   = $1 , verifier_status = $2 , wallet_status =$3  WHERE id = $4"
	query = "UPDATE requests_for_vc SET request_status = $1 , verifier_status = $2 , wallet_status =$3  WHERE id = $4"
	res, err := conn.Exec(ctx, query, "VC_Issued", "VC_Issued", "Issued", id)
	if err != nil {
		return 0, err
	}
	fmt.Println("Updating Status", res)
	return res.RowsAffected(), nil
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

func (i *requests) NewNotification(ctx context.Context, conn db.Querier, req *domain.NotificationData) (bool, error) {
	query := "INSERT INTO notifications ( id, user_id, module, notification_type, notification_title, notification_message) VALUES ($1,$2,$3,$4,$5,$6)"
	_, err := conn.Exec(ctx, query, req.ID, req.User_id, req.Module, req.NotificationType, req.NotificationTitle, req.NotificationMessage)
	if err != nil {
		return false, err
	}
	return true, nil
}


func (i *requests)GetNotifications(ctx context.Context, conn db.Querier)([]*domain.NotificationReponse,error){

	response := NotificationReponse{}
	domainResponce := make([]*domain.NotificationReponse, 0)
	rows, err := conn.Query(ctx, `SELECT id,user_id,module,notification_type,notification_title,notification_message,created_at from notifications`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&response.id,
			&response.user_id,
			&response.module,
			&response.notification_type,
			&response.notification_title,
			&response.notification_message,
			&response.created_at,
		); err != nil {
			return nil, err
		}
		temp := &domain.NotificationReponse{
			ID: response.id,
			User_id: response.user_id,
			Module: response.module,
			NotificationType: response.notification_type,
			NotificationTitle: response.notification_title,
			NotificationMessage: response.notification_message,
			CreatedAt: response.created_at,
		}
		domainResponce = append(domainResponce, temp)
	}
	return domainResponce, nil
}
