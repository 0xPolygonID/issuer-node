-- +goose Up
-- +goose StatementBegin
CREATE TABLE keys(
    id                              UUID PRIMARY KEY NOT NULL,
    issuer_did                      text NOT NULL,
    name                            text NOT NULL,
    public_key                      text NOT NULL,
    created_at                      timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                      timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT keys_unique_name UNIQUE (issuer_did, name),
    CONSTRAINT keys_identities_id_key foreign key (issuer_did) references identities (identifier)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS keys;
-- +goose StatementEnd