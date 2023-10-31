-- +goose Up
-- +goose StatementBegin
ALTER TABLE identities
    ADD COLUMN address TEXT,
    ADD COLUMN keyType TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identities
DROP COLUMN address,
    DROP COLUMN keyType;
-- +goose StatementEnd