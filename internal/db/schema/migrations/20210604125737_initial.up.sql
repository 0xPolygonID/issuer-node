-- +goose Up
CREATE TABLE mt_nodes (
    mt_id BIGINT,
    key BYTEA,
    type SMALLINT NOT NULL,
    child_l BYTEA,
    child_r BYTEA,
    entry BYTEA,
    created_at BIGINT,
    deleted_at BIGINT,
    PRIMARY KEY(mt_id, key)
);

CREATE TABLE mt_roots (
    mt_id BIGINT PRIMARY KEY,
    key BYTEA,
    created_at BIGINT,
    deleted_at BIGINT
);

CREATE TABLE identity_mts (
    id BIGSERIAL PRIMARY KEY,
    identifier TEXT,
    type SMALLINT NOT NULL,
    UNIQUE (identifier, type)
);

-- +goose Down
DROP TABLE IF EXISTS identity_mts;
DROP TABLE IF EXISTS mt_nodes;
DROP TABLE IF EXISTS mt_roots;