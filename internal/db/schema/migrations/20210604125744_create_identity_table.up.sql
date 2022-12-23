
-- +goose Up
CREATE TABLE IF NOT EXISTS identities (
   identifier TEXT PRIMARY KEY,
   relay TEXT,
   immutable boolean DEFAULT false
);

-- +goose Down
DROP TABLE IF EXISTS identities;