-- +goose Up
-- +goose StatementBegin
ALTER TABLE identities
    ADD COLUMN address TEXT,
    ADD COLUMN keytype TEXT;

UPDATE identities set keytype = 'BJJ' where keyType is null;

ALTER TABLE identities
    ALTER COLUMN keytype SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identities
DROP COLUMN address,
    DROP COLUMN keyType;
-- +goose StatementEnd