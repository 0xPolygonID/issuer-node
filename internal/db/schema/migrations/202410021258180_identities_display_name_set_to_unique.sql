-- +goose Up
-- +goose StatementBegin
WITH cte AS (SELECT identifier,
                    display_name,
                    ROW_NUMBER() OVER (PARTITION BY display_name ORDER BY identifier) AS rn
             FROM identities
             WHERE display_name IS NOT NULL)
UPDATE identities
SET display_name = identities.display_name || '_' || rn
FROM cte
WHERE identities.identifier = cte.identifier
  AND cte.rn > 1;

ALTER TABLE identities
    ADD CONSTRAINT identities_display_name_unique UNIQUE (display_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE identities
    DROP CONSTRAINT identities_display_name_unique;
-- +goose StatementEnd
