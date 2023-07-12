-- +goose Up
-- +goose StatementBegin
ALTER TABLE schemas DROP COLUMN ts_words;
ALTER TABLE schemas RENAME COLUMN attributes TO words;
UPDATE schemas SET words = schemas.type || ',' || schemas.words;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schemas ADD COLUMN ts_words tsvector DEFAULT to_tsvector(''::text) NOT NULL;
ALTER TABLE schemas RENAME COLUMN words TO attributes;
UPDATE schemas SET ts_words = to_tsvector(schemas.attributes);
-- +goose StatementEnd