-- +goose Up
-- +goose StatementBegin
CREATE TABLE requests_for_vc
(
    id         uuid                                      NOT NULL,
    user_id    text                                      NOT NULL,
    issuer_id    text                                      NOT NULL,
    schema_id  text                                      NOT NULL,
    active     bool                                      NOT NULL,
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