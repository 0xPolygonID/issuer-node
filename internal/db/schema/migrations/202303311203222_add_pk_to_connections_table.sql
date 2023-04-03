-- +goose Up
-- +goose StatementBegin
ALTER TABLE connections
    ADD primary key (id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connections DROP CONSTRAINT connections_pkey;
-- +goose StatementEnd