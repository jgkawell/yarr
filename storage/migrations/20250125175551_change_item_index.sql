-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

drop index if exists idx_item_status;
create index if not exists idx_item__date_id_status on items(date,id,status);

-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

drop index if exists idx_item__date_id_status;
create index if not exists idx_item_status  on items(status);

-- +goose StatementEnd
