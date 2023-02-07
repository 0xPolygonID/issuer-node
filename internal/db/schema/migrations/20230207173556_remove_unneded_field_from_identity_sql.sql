-- +goose Up
-- +goose StatementBegin
ALTER TABLE identities DROP COLUMN relay;
ALTER TABLE identities DROP COLUMN "immutable";
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE identities ADD COLUMN relay text NULL;
ALTER TABLE identities ADD COLUMN "immutable" bool NULL default false;
-- +goose StatementEnd
