-- +goose Up
-- +goose StatementBegin
CREATE TABLE requests_for_vc
(
    id             uuid                                       NOT NULL,
    UDID           text                                       NOT NULL,
    issuer_id      text                                       NOT NULL,
    schema_id      text                                       NOT NULL,
    credential_type text                                      NOT NULL,
    request_type   text                                       NOT NULL,
    role_type      text                                       NOT NULL,
    proof_type     text                                       NOT NULL,
    proof_id       text                                       NOT NULL,
    age            text                                       NOT NULL,
    active         bool                                       NOT NULL,
    request_status text                                       NOT NULL,
    verifier_status text                                       NOT NULL,
    wallet_status   text                                       NOT NULL,
    source          text                                       NOT NULL,
    created_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT requests_for_vc_pkey PRIMARY KEY (id)
);

CREATE TABLE requests_for_auth
(
    id         uuid                                      NOT NULL,
    user_id    text                                      NOT NULL,
    authType   text                                      NOT NULL,
    authId     text                                      NOT NULL,
    created_at int8                                      NOT NULL,
    active     bool                                      NOT NULL,
    CONSTRAINT requests_for_auth_pkey PRIMARY KEY (id)
);


CREATE TABLE notifications
(
    id                      uuid                         NOT NULL,
    user_id                 text                         NOT NULL,
    module                  text                         NOT NULL,
    notification_type       text                         NOT NULL,
    notification_title      text                         NOT NULL,
    notification_message    text                         NOT NULL,
    created_at timestamptz NULL DEFAULT CURRENT_TIMESTAMP        ,
    CONSTRAINT notifications_pkey PRIMARY KEY (id)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS requests_for_vc;
DROP TABLE IF EXISTS requests_for_auth;
-- +goose StatementEnd



-- -- +goose Up
-- -- +goose StatementBegin
-- CREATE TABLE requests_for_vc (
--     id uuid NOT NULL,
--     user_id text NOT NULL,
--     schema_id uuid NOT NULL,
--     active bool DEFAULT ture
--     CONSTRAINT requests_for_vc_pkey PRIMARY KEY (id)
-- )

-- -- CREATE TYPE authType AS ENUM ('PAN', 'ADHAR');

-- -- CREATE TABLE requests_for_auth (
-- --     id uuid NOT NULL,
-- --     user_id text NOT NULL,
-- --     authType text NOT NULL,
-- --     authId  text NOT NULL,
-- --     created_at int8 NULL,
-- --     active bool DEFAULT ture
-- --     CONSTRAINT requests_for_auth_pkey PRIMARY KEY (id)
-- -- )
-- -- +goose StatementEnd

-- -- +goose Down

-- -- +goose StatementBegin
-- DROP TABLE IF EXISTS requests_for_vc;
-- DROP TABLE IF EXISTS requests_for_auth;
-- -- +goose StatementEnd