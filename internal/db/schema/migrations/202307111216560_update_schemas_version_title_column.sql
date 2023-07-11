-- +goose Up
-- +goose StatementBegin
ALTER TABLE schemas
    ALTER COLUMN title DROP NOT NULL;

ALTER TABLE schemas
    ALTER COLUMN title DROP DEFAULT;

ALTER TABLE schemas
    ALTER COLUMN description DROP NOT NULL;

ALTER TABLE schemas
    ALTER COLUMN description DROP DEFAULT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schemas
    ALTER COLUMN title SET NOT NULL DEFAULT '';

ALTER TABLE schemas
    ALTER COLUMN description SET NOT NULL DEFAULT '';
-- +goose StatementEnd
