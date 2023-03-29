-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN issued int not null;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE links DROP COLUMN issued;
-- +goose StatementEnd