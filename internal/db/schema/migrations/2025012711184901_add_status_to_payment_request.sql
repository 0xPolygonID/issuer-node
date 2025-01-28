-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_requests
    ADD COLUMN modified_at timestamptz;

ALTER TABLE payment_requests
    ADD COLUMN status TEXT;

ALTER TABLE payment_requests
    ADD COLUMN paid_nonce numeric NULL;

UPDATE payment_requests set modified_at = created_at where modified_at is null;
UPDATE payment_requests set status = 'not-verified' where status is null;

ALTER TABLE payment_requests
    ALTER COLUMN modified_at SET NOT NULL;

ALTER TABLE payment_requests
    ALTER COLUMN status SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_requests
    DROP COLUMN modified_at;

ALTER TABLE payment_requests
    DROP COLUMN status;

ALTER TABLE payment_requests
    DROP COLUMN paid_nonce;
-- +goose StatementEnd