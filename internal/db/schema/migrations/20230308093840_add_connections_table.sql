-- +goose Up
-- +goose StatementBegin
ALTER TABLE connections DROP CONSTRAINT connections_managed_identifier_third_party_identifier_key;
DROP TABLE IF EXISTS connections;
CREATE TABLE connections
(
    id                     uuid,
    issuer_id     text NOT NULL,
    user_id text NOT NULL,
    issuer_doc        jsonb NULL,
    user_doc    jsonb NULL,
    created_at             timestamptz NOT NULL,
    modified_at             timestamptz NOT NULL,
    CONSTRAINT connections_issuer_user_key UNIQUE (issuer_id, user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE connections DROP CONSTRAINT connections_issuer_user_key;
DROP TABLE IF EXISTS connections;
-- +goose StatementEnd
