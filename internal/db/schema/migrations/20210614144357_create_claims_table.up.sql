
-- +goose Up
CREATE TABLE IF NOT EXISTS claims (
    id uuid DEFAULT gen_random_uuid(),
    identifier TEXT,
    issuer TEXT,
    schema_hash text NOT NULL,
    schema_url text NOT NULL,
    schema_type text NOT NULL,
    other_identifier TEXT,
    expiration bigint,
    updatable boolean DEFAULT false,
    revoked boolean DEFAULT false,
    version bigint,
    rev_nonce NUMERIC,
    metadata text,
    core_claim text,
    mtp_proof text,
    data text,

    signature_proof json,
    mtp json,
    merkle_root VARCHAR(64),
    identity_state VARCHAR(64) NULL,
    credential_status json,
    index_hash VARCHAR,

    PRIMARY KEY(id, identifier),
    UNIQUE (identifier, issuer, index_hash)
);

CREATE INDEX claims_iden_other_iden ON claims (identifier, other_identifier);

-- +goose Down
DROP TABLE IF EXISTS claims cascade;