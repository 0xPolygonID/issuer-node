-- +goose Up
-- +goose StatementBegin
ALTER TABLE links
    DROP COLUMN issued_claims;
ALTER TABLE claims
    ADD COLUMN link_id uuid;
CONSTRAINT claims_links_id_key foreign key (link_id) references links (id),
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE links
    ADD COLUMN issued_claims int not null;
ALTER TABLE claims
    DROP COLUMN lind_id;
-- +goose StatementEnd