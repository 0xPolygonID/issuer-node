-- +goose Up
-- +goose StatementBegin
ALTER TABLE schemas
    ADD COLUMN version text NOT NULL DEFAULT '',
    ADD COLUMN title text NOT NULL DEFAULT '',
    ADD COLUMN description text NOT NULL DEFAULT '';

ALTER TABLE schemas DROP CONSTRAINT schemas_issuer_id_url_key;
ALTER TABLE schemas ADD CONSTRAINT schemas_issuer_id_url_version_key UNIQUE (issuer_id, url, version);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schemas
    DROP COLUMN version,
    DROP COLUMN title,
    DROP COLUMN description;

ALTER TABLE schemas DROP CONSTRAINT schemas_issuer_id_url_version_key;
ALTER TABLE schemas ADD CONSTRAINT schemas_issuer_id_url_key UNIQUE (issuer_id, url);
-- +goose StatementEnd
