-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_authentications
    ADD COLUMN updated_at timestamptz;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_authentications
DROP COLUMN updated_at;
-- +goose StatementEnd