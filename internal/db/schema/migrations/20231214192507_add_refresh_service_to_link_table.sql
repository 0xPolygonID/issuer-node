-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN refresh_service JSONB NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
    DROP COLUMN refresh_service;
-- +goose StatementEnd
