-- +goose Up
-- +goose StatementBegin
CREATE TABLE display_methods(
    id                              UUID PRIMARY KEY NOT NULL,
    issuer_did                      text NOT NULL,
    name                            text NOT NULL,
    url                             text NOT NULL,
    created_at                      timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    is_default                      bool NOT NULL,
    CONSTRAINT display_methods_identities_id_key foreign key (issuer_did) references identities (identifier),
    UNIQUE (issuer_did, is_default)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS display_methods;
-- +goose StatementEnd
