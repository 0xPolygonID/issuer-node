-- +goose Up
-- +goose StatementBegin
ALTER TABLE claims DROP COLUMN mtp;
ALTER TABLE claims
    ADD COLUMN mtp bool not null default false;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims DROP COLUMN mtp;
ALTER TABLE claims
    ADD COLUMN mtp jsonb not null;
-- +goose StatementEnd
