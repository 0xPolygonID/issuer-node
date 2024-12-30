-- +goose Up
-- +goose StatementBegin
CREATE TABLE payment_options
(
    id            UUID PRIMARY KEY NOT NULL,
    issuer_did    text             NOT NULL REFERENCES identities (identifier),
    name          text             NOT NULL,
    description   text             NOT NULL,
    configuration jsonb            NOT NULL,
    created_at    timestamptz      NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    timestamptz      NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX payment_options_id_issuer_did_index ON payment_options (id, issuer_did);
CREATE INDEX payment_options_issuer_did_created_at_index ON payment_options (issuer_did, created_at);
CREATE INDEX payment_options_issuer_did ON payment_options (issuer_did);

CREATE TABLE payment_requests
(
    id                UUID PRIMARY KEY NOT NULL,
    credentials       jsonb            NOT NULL, /* []protocol.PaymentRequestCredentials */
    description       text             NOT NULL,
    issuer_did        text             NOT NULL REFERENCES identities (identifier),
    recipient_did     text             NOT NULL,
    payment_option_id UUID REFERENCES payment_options (id),
    created_at        timestamptz      NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX payment_requests_from_nonce_created_at_idx ON payment_requests (issuer_did, created_at);

CREATE TABLE payment_request_items
(
    id                   UUID PRIMARY KEY NOT NULL,
    nonce                numeric          NOT NULL,
    payment_request_id   UUID             NOT NULL REFERENCES payment_requests (id),
    payment_option_id    int              NOT NULL,
    signing_key          text             NOT NULL,
    payment_request_info jsonb            NOT NULL /* protocol.PaymentRequestInfo */
);

CREATE UNIQUE INDEX payment_request_items_nonce_idx ON payment_request_items (nonce);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS payment_requests_items;
DROP TABLE IF EXISTS payment_requests;
DROP TABLE IF EXISTS payment_options;
-- +goose StatementEnd