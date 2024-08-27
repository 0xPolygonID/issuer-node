-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN credential_status_type text NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
DROP COLUMN credential_status_type;
-- +goose StatementEnd