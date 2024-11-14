-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_options
(
    id            UUID PRIMARY KEY NOT NULL,
    issuer_did    text             NOT NULL REFERENCES identities (identifier),
    name          text,
    description   text,
    configuration jsonb            NOT NULL,
    created_at    timestamptz      NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamptz      NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX payment_options_id_issuer_did_index ON payment_options (id, issuer_did);
CREATE INDEX payment_options_issuer_did_created_at_index ON payment_options (issuer_did, created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_options;
-- +goose StatementEnd