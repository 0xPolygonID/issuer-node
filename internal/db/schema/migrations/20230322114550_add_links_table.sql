-- +goose Up
-- +goose StatementBegin
CREATE TABLE links
(
    id                              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer_id                       text NOT NULL,
    created_at                      timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    max_issuance                    numeric NULL,
    valid_until                     timestamptz NULL,
    schema_id                       uuid NOT NULL,
    credential_expiration           timestamptz NULL,
    credential_signature_proof      bool NULL DEFAULT false,
    credential_mtp_proof            bool NULL DEFAULT false,
    credential_attributes           jsonb NOT NULL,
    active                          bool NULL DEFAULT true,
    CONSTRAINT links_schemas_id_key foreign key (schema_id) references schemas (id),
    CONSTRAINT links_indetities_id_key foreign key (issuer_id) references identities (identifier)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd
