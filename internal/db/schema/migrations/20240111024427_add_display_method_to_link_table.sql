-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN display_method JSONB NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
    DROP COLUMN display_method;
-- +goose StatementEnd
