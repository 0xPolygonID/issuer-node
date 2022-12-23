
-- +goose Up
ALTER TABLE revocation
    ALTER COLUMN identifier TYPE TEXT;
