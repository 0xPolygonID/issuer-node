
-- +goose Up
ALTER TABLE claims
    ALTER COLUMN mtp_proof TYPE JSONB,
    ALTER COLUMN data TYPE JSONB,
    ALTER COLUMN signature_proof TYPE JSONB,
    ALTER COLUMN mtp TYPE JSONB,
    ALTER COLUMN credential_status TYPE JSONB;
-- +goose Down
ALTER TABLE claims
    ALTER COLUMN mtp_proof TYPE JSON,
    ALTER COLUMN data TYPE JSON,
    ALTER COLUMN signature_proof TYPE JSON,
    ALTER COLUMN mtp TYPE JSON,
    ALTER COLUMN credential_status TYPE JSON;
