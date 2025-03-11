-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_requests RENAME COLUMN recipient_did TO user_did;

ALTER TABLE payment_requests
    ADD COLUMN schema_id uuid,
    ADD CONSTRAINT payment_requests_schema_id_fkey FOREIGN KEY (schema_id) REFERENCES schemas(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_requests RENAME COLUMN user_did TO recipient_did;
ALTER TABLE payment_requests DROP COLUMN schema_id;
-- +goose StatementEnd