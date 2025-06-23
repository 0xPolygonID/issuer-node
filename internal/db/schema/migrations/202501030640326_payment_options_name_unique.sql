-- +goose Up
-- +goose StatementBegin

-- Update the name column with unique names
WITH duplicates AS (SELECT id, name, ROW_NUMBER() OVER (PARTITION BY name ORDER BY id) AS row_num
                    FROM payment_options)
UPDATE payment_options
SET name =payment_options.name || '_' || duplicates.row_num - 1
FROM duplicates
WHERE payment_options.id = duplicates.id
  AND duplicates.row_num > 1;

ALTER TABLE payment_options
    ADD CONSTRAINT payment_options_name_unique UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_options
    DROP CONSTRAINT payment_options_name_unique;
-- +goose StatementEnd