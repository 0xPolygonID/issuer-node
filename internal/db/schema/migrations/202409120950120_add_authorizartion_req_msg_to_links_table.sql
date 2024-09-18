
-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN authorization_request_message jsonb NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
DROP COLUMN authorization_request_message;
-- +goose StatementEnd
