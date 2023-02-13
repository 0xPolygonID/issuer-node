
-- +goose Up
-- +goose StatementBegin
CREATE TABLE claims (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    identifier text NOT NULL,
    issuer text NULL,
    schema_hash text NOT NULL,
    schema_url text NOT NULL,
    schema_type text NOT NULL,
    other_identifier text NULL,
    expiration int8 NULL,
    updatable bool NULL DEFAULT false,
    revoked bool NULL DEFAULT false,
    "version" int8 NULL,
    rev_nonce numeric NULL,
    metadata text NULL,
    core_claim text NULL,
    mtp_proof jsonb NULL,
    "data" jsonb NULL,
    signature_proof jsonb NULL,
    mtp jsonb NULL,
    merkle_root varchar(64) NULL,
    identity_state varchar(64) NULL,
    credential_status jsonb NULL,
    index_hash varchar NULL,
    CONSTRAINT claims_identifier_issuer_index_hash_key UNIQUE (identifier, issuer, index_hash),
    CONSTRAINT claims_pkey PRIMARY KEY (id, identifier)
);

CREATE INDEX claims_iden_other_iden ON claims USING btree (identifier, other_identifier);


CREATE TABLE connections (
    id uuid NULL DEFAULT gen_random_uuid(),
    managed_identifier text NOT NULL,
    third_party_identifier text NOT NULL,
    managed_did_doc jsonb NULL,
    third_party_did_doc jsonb NULL,
    created_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT connections_managed_identifier_third_party_identifier_key UNIQUE (managed_identifier, third_party_identifier)
);

CREATE TABLE identities (
    identifier text NOT NULL,
    relay text NULL,
    "immutable" bool NULL DEFAULT false,
    CONSTRAINT identities_pkey PRIMARY KEY (identifier)
);

CREATE TABLE identity_mts (
     id bigserial NOT NULL,
     identifier text NULL,
     "type" int2 NOT NULL,
     CONSTRAINT identity_mts_identifier_type_key UNIQUE (identifier, type),
     CONSTRAINT identity_mts_pkey PRIMARY KEY (id)
);

CREATE TABLE mt_nodes (
     mt_id int8 NOT NULL,
     "key" bytea NOT NULL,
     "type" int2 NOT NULL,
     child_l bytea NULL,
     child_r bytea NULL,
     entry bytea NULL,
     created_at int8 NULL,
     deleted_at int8 NULL,
     CONSTRAINT mt_nodes_pkey PRIMARY KEY (mt_id, key)
);

CREATE TABLE mt_roots (
     mt_id int8 NOT NULL,
     "key" bytea NULL,
     created_at int8 NULL,
     deleted_at int8 NULL,
     CONSTRAINT mt_roots_pkey PRIMARY KEY (mt_id)
);

CREATE TYPE status AS ENUM ('created', 'transacted', 'confirmed', 'failed');

CREATE TABLE identity_states (
    state_id int4 NOT NULL GENERATED ALWAYS AS IDENTITY,
    identifier text NULL,
    state varchar(64) NOT NULL,
    root_of_roots varchar(64) NULL,
    revocation_tree_root varchar(64) NULL,
    claims_tree_root varchar(64) NULL,
    block_timestamp int4 NULL,
    block_number int4 NULL,
    tx_id varchar(66) NULL,
    previous_state varchar(64) NULL,
    "status" "status" NULL DEFAULT 'created'::status,
    modified_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    created_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT identity_states_identifier_state_key UNIQUE (identifier, state),
    CONSTRAINT identity_states_pkey PRIMARY KEY (state_id),
    CONSTRAINT fk_identity FOREIGN KEY (identifier) REFERENCES public.identities(identifier)
);

CREATE OR REPLACE FUNCTION update_modified_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
RETURN NEW;
END;
$$
language plpgsql;

CREATE TRIGGER update_state_modifiedtime BEFORE UPDATE ON identity_states FOR EACH ROW EXECUTE PROCEDURE  update_modified_at_column();

CREATE TABLE revocation (
   id int8 NOT NULL GENERATED ALWAYS AS IDENTITY,
   identifier text NULL,
   nonce numeric NOT NULL,
   "version" int4 NULL,
   "status" int2 NULL,
   description text NULL,
   modified_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
   created_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
   CONSTRAINT revocation_identifier_nonce_version_key UNIQUE (identifier, nonce, version),
   CONSTRAINT revocation_pkey PRIMARY KEY (id),
   CONSTRAINT fk_identity FOREIGN KEY (identifier) REFERENCES public.identities(identifier)
);

CREATE TRIGGER update_revocation_modifiedtime
    BEFORE UPDATE ON revocation FOR EACH ROW EXECUTE PROCEDURE  update_modified_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS identity_mts;
DROP TABLE IF EXISTS mt_nodes;
DROP TABLE IF EXISTS mt_roots;
DROP TABLE IF EXISTS identities;
DROP TRIGGER IF EXISTS update_state_modifiedtime on identity_states;
DROP TABLE IF EXISTS identity_states;
DROP TYPE IF EXISTS status;
DROP FUNCTION IF EXISTS update_modified_at_column;
DROP TABLE IF EXISTS claims cascade;
DROP TABLE IF EXISTS revocation;
DROP TRIGGER IF EXISTS update_revocation_modifiedtime ON revocation;
ALTER TABLE connections DROP CONSTRAINT connections_managed_identifier_third_party_identifier_key;
DROP TABLE connections;
-- +goose StatementEnd
