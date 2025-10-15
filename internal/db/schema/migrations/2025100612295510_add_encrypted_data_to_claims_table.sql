-- +goose Up
-- +goose StatementBegin
ALTER TABLE claims
    ADD COLUMN encrypted_data TEXT,
    ADD COLUMN context_url TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims
    DROP COLUMN encrypted_data,
    DROP  COLUMN context_url;
-- +goose StatementEnd