-- +goose Up
-- +goose StatementBegin
ALTER TABLE claims
    ADD COLUMN created_at timestamptz DEFAULT CURRENT_TIMESTAMP     NOT NULL;

UPDATE claims
    SET created_at = (data->>'issuanceDate')::timestamp;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims
    DROP COLUMN created_at;
-- +goose StatementEnd
