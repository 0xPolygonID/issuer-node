-- +goose Up
-- +goose StatementBegin
CREATE TABLE schemas
(
    id         uuid                                      NOT NULL,
    issuer_id  text                                      NOT NULL,
    url        text                                      NOT NULL,
    type       text                                      NOT NULL,
    attributes text                                      NOT NULL,
    hash       text                                      NOT NULL,
    ts_words   tsvector    DEFAULT to_tsvector(''::text) NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP     NOT NULL,
    CONSTRAINT schemas_issuer_id_url_key UNIQUE (issuer_id, url),
    CONSTRAINT schemas_pkey PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS schemas;
-- +goose StatementEnd
