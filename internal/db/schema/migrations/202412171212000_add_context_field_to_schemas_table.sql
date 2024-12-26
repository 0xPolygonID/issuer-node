-- +goose Up
-- +goose StatementBegin
ALTER TABLE schemas ADD COLUMN context_url text NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schemas DROP COLUMN context_url;
-- +goose StatementEnd