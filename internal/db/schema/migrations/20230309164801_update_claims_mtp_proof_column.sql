-- +goose Up
-- +goose StatementBegin
UPDATE claims SET mtp_proof = jsonb_set(mtp_proof, '{type}', '"Iden3SparseMerkleTreeProof"');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE claims SET mtp_proof = jsonb_set(mtp_proof, '{type}', '"Iden3SparseMerkleProof"');
-- +goose StatementEnd
