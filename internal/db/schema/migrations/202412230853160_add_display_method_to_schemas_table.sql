-- +goose Up
-- +goose StatementBegin
ALTER TABLE schemas
    ADD COLUMN display_method_id uuid,
    ADD CONSTRAINT schemas_display_method_id_fkey FOREIGN KEY (display_method_id) REFERENCES display_methods(id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE schemas
    DROP CONSTRAINT schemas_display_method_id_fkey,
    DROP COLUMN display_method_id;
-- +goose StatementEnd
