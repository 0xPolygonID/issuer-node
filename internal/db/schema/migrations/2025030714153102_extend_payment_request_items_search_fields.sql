-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_requests RENAME COLUMN recipient_did TO user_did;
CREATE INDEX payment_requests_idx_credentials ON payment_requests USING GIN (credentials);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_requests RENAME COLUMN user_did TO recipient_did;
DROP INDEX IF EXISTS payment_requests_idx_credentials;
-- +goose StatementEnd