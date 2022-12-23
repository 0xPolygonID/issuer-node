
-- +goose Up
ALTER TABLE claims
    ALTER COLUMN mtp_proof TYPE JSON USING mtp_proof::json,
    ALTER COLUMN data TYPE JSON USING data::json;
-- +goose Down
ALTER TABLE claims
    ALTER COLUMN mtp_proof TYPE TEXT,
    ALTER COLUMN data TYPE TEXT;