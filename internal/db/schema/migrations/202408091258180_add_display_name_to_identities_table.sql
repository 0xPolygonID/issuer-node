-- +goose Up
-- +goose StatementBegin
ALTER TABLE identities
    ADD COLUMN display_name text NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identities
    DROP COLUMN display_name;
-- +goose StatementEnd
