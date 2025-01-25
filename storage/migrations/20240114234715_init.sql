-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';

create collation nocase (
	provider = icu,
	locale = 'und-u-ks-level2',
	deterministic = false
);

-- m01_initial -> m08_normalize_datetime (from upstream)

create table if not exists folders (
	id             serial primary key,
	title          text not null,
	is_expanded    boolean not null default false
);
create unique index if not exists idx_folder_title on folders(title);

create table if not exists feeds (
	id             serial primary key,
	folder_id      integer references folders(id) on delete set null,
	title          text not null,
	description    text,
	link           text,
	feed_link      text not null,
	icon           bytea
);
create index if not exists idx_feed_folder_id on feeds(folder_id);
create unique index if not exists idx_feed_feed_link on feeds(feed_link);

create table if not exists items (
	id             serial primary key,
	guid           text not null,
	feed_id        integer references feeds(id) on delete cascade,
	title          text,
	link           text,
	description    text,
	content        text,
	author         text,
	date           timestamp,
	date_updated   timestamp,
	date_arrived   timestamp,
	status         integer,
	image          text,
	search_rowid   integer,
	podcast_url    text
);
create index if not exists idx_item_feed_id on items(feed_id);
create index if not exists idx_item_status  on items(status);
create index if not exists idx_item_search_rowid on items(search_rowid);
create unique index if not exists idx_item_guid on items(feed_id, guid);

create table if not exists settings (
	key            text primary key,
	val            bytea
);

-- // TODO: replicate this behavior outside of the database
-- // create virtual table if not exists search using fts4(title, description, content);

-- // create trigger if not exists del_item_search after delete on items begin
-- //   delete from search where rowid = old.search_rowid;
-- // end;

create table if not exists http_states (
	feed_id        integer references feeds(id) on delete cascade unique,
	last_refreshed timestamp not null,
	last_modified  text not null,
	etag           text not null
);

create table if not exists feed_errors (
	feed_id        integer references feeds(id) on delete cascade unique,
	error          text
);

create table if not exists feed_sizes (
	feed_id        integer references feeds(id) on delete cascade unique,
	size           integer not null default 0
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';

drop collation if exists nocase;
drop table if exists folders;
drop table if exists feeds;
drop table if exists items;
drop table if exists settings;
drop table if exists http_states;
drop table if exists feed_errors;
drop table if exists feed_sizes;

-- +goose StatementEnd
