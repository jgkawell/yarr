-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE items ADD COLUMN search tsvector
    GENERATED ALWAYS AS
     (
     	setweight(to_tsvector('english', coalesce(title, '')), 'A') ||
     	setweight(to_tsvector('english', coalesce(description, '')), 'B') ||
     	setweight(to_tsvector('english', coalesce(content, '')), 'C')
     ) STORED;

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE items DROP COLUMN IF EXISTS search;

-- +goose StatementEnd
