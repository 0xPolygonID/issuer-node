-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS verification_queries (
    id uuid NOT NULL PRIMARY KEY,
    issuer_id text NOT NULL,
    chain_id integer NOT NULL,
    skip_check_revocation boolean NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT verification_queries_indentities_id_fk foreign key (issuer_id) references identities (identifier)
);

CREATE TABLE IF NOT EXISTS verification_scopes (
    id uuid NOT NULL PRIMARY KEY,
    verification_query_id uuid NOT NULL,
    scope_id integer NOT NULL,
    circuit_id text NOT NULL,
    context text NOT NULL,
    allowed_issuers text[] NOT NULL,
    credential_type text NOT NULL,
    credential_subject jsonb NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT verification_scopes_verification_query_id_fk foreign key (verification_query_id) references verification_queries (id)
);

CREATE TABLE IF NOT EXISTS verification_responses (
    id uuid NOT NULL PRIMARY KEY,
    verification_scope_id uuid NOT NULL,
    user_did text NOT NULL,
    response jsonb NOT NULL,
    pass boolean NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NOT NULL,
    CONSTRAINT verification_responses_verification_scopes_id_fk foreign key (verification_scope_id) references verification_scopes (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE verification_responses DROP CONSTRAINT verification_responses_verification_scopes_id_fk;
ALTER TABLE verification_scopes DROP CONSTRAINT verification_scopes_verification_query_id_fk;
ALTER TABLE verification_queries DROP CONSTRAINT verification_queries_indentities_id_fk;
DROP TABLE IF EXISTS verification_queries;
DROP TABLE IF EXISTS verification_responses;
DROP TABLE IF EXISTS verification_scopes;
-- +goose StatementEnd