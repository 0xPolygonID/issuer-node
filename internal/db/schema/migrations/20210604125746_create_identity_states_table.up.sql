
-- +goose Up
CREATE TYPE status AS ENUM ('created', 'transacted', 'confirmed', 'failed');

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_modified_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.modified_at = NOW();
    RETURN NEW;
END;
$$
language plpgsql;
-- +goose StatementEnd

CREATE TABLE IF NOT EXISTS identity_states (
   state_id INT GENERATED ALWAYS AS IDENTITY,
   identifier TEXT,
   state VARCHAR(64) NOT NULL,
   root_of_roots VARCHAR(64),
   revocation_tree_root VARCHAR(64),
   claims_tree_root VARCHAR(64),
   block_timestamp INT,
   block_number INT,
   tx_id VARCHAR(66),
   previous_state VARCHAR(64),
   status  status default 'created',
   modified_at TIMESTAMP WITH TIME ZONE default current_timestamp,
   created_at TIMESTAMP WITH TIME ZONE DEFAULT current_timestamp,
   UNIQUE (identifier, state),
   PRIMARY KEY(state_id),
   CONSTRAINT fk_identity
      FOREIGN KEY(identifier) 
	  REFERENCES identities(identifier)
);

CREATE TRIGGER update_state_modifiedtime BEFORE UPDATE ON identity_states FOR EACH ROW EXECUTE PROCEDURE  update_modified_at_column();


-- +goose Down
DROP TRIGGER IF EXISTS update_state_modifiedtime on identity_states;
DROP TABLE IF EXISTS identity_states;
DROP TYPE IF EXISTS status;
DROP FUNCTION IF EXISTS update_modified_at_column;