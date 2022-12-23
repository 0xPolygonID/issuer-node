-- +goose Up
CREATE TABLE connections
(
    id                     uuid                     DEFAULT gen_random_uuid(),
    managed_identifier     text NOT NULL,
    third_party_identifier text NOT NULL,
    managed_did_doc        jsonb NULL,
    third_party_did_doc    jsonb NULL,
    created_at             TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    modified_at             TIMESTAMP WITH TIME ZONE default current_timestamp,
    CONSTRAINT connections_managed_identifier_third_party_identifier_key UNIQUE (managed_identifier, third_party_identifier)
);
-- +goose Down
ALTER TABLE connections DROP CONSTRAINT connections_managed_identifier_third_party_identifier_key;
DROP TABLE connections;
