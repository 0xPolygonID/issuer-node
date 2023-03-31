-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN issued_claims int not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links DROP COLUMN issued_claims;
-- +goose StatementEnd