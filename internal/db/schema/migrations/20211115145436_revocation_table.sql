-- +goose Up
CREATE TABLE IF NOT EXISTS revocation (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    identifier VARCHAR (42),
    nonce NUMERIC NOT NULL,
    version INT,
    status SMALLINT,
    description TEXT,
    modified_at TIMESTAMP WITH TIME ZONE default current_timestamp,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
    UNIQUE (identifier, nonce, version),
    PRIMARY KEY(id),
    CONSTRAINT fk_identity
        FOREIGN KEY(identifier)
            REFERENCES identities(identifier)
);

CREATE TRIGGER update_revocation_modifiedtime
    BEFORE UPDATE ON revocation FOR EACH ROW EXECUTE PROCEDURE  update_modified_at_column();



-- +goose Down
DROP TABLE IF EXISTS revocation;
DROP TRIGGER IF EXISTS update_revocation_modifiedtime ON revocation;
