-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

ALTER TABLE items ADD column media_links jsonb[];
UPDATE items SET media_links =
    CASE
        WHEN coalesce(image, '') != '' AND coalesce(podcast_url, '') != ''
            THEN ARRAY[json_build_object('type', 'image', 'url', image), json_build_object('type', 'audio', 'url', podcast_url)]
        WHEN coalesce(image, '') != '' AND coalesce(podcast_url, '') = ''
            THEN ARRAY[json_build_object('type', 'image', 'url', image)]
        WHEN coalesce(podcast_url, '') != '' AND coalesce(image, '') = ''
            THEN ARRAY[json_build_object('type', 'audio', 'url', podcast_url)]
        ELSE
            NULL
        END;

-- TODO: may want to come back later and drop these columns
-- ALTER TABLE items DROP column image;
-- ALTER TABLE items DROP column podcast_url;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

ALTER TABLE items DROP COLUMN IF EXISTS media_links;

-- +goose StatementEnd
