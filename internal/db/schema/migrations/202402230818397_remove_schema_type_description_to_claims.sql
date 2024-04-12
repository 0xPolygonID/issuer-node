-- +goose Up
-- +goose StatementBegin
ALTER TABLE claims
    DROP COLUMN schema_type_description;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE claims
    ADD COLUMN schema_type_description text;
-- +goose StatementEnd