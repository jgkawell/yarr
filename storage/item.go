package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/jgkawell/yarr/content/htmlutil"
)

type ItemStatus int

const (
	UNREAD  ItemStatus = 0
	READ    ItemStatus = 1
	STARRED ItemStatus = 2
)

var StatusRepresentations = map[ItemStatus]string{
	UNREAD:  "unread",
	READ:    "read",
	STARRED: "starred",
}

var StatusValues = map[string]ItemStatus{
	"unread":  UNREAD,
	"read":    READ,
	"starred": STARRED,
}

func (s ItemStatus) MarshalJSON() ([]byte, error) {
	return json.Marshal(StatusRepresentations[s])
}

func (s *ItemStatus) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}
	*s = StatusValues[str]
	return nil
}

type Item struct {
	Id       int64      `json:"id"`
	GUID     string     `json:"guid"`
	FeedId   int64      `json:"feed_id"`
	Title    string     `json:"title"`
	Link     string     `json:"link"`
	Content  string     `json:"content,omitempty"`
	Date     time.Time  `json:"date"`
	Status   ItemStatus `json:"status"`
	ImageURL *string    `json:"image"`
	AudioURL *string    `json:"podcast_url"`
}

type ItemFilter struct {
	FolderID *int64
	FeedID   *int64
	Status   *ItemStatus
	Search   *string
	After    *int64
	IDs      *[]int64
	SinceID  *int64
	MaxID    *int64
	Before   *time.Time
}

type MarkFilter struct {
	FolderID *int64
	FeedID   *int64

	Before *time.Time
}

type ItemList []Item

func (list ItemList) Len() int {
	return len(list)
}

func (list ItemList) SortKey(i int) string {
	return list[i].Date.Format(time.RFC3339) + "::" + list[i].GUID
}

func (list ItemList) Less(i, j int) bool {
	return list.SortKey(i) < list.SortKey(j)
}

func (list ItemList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (s *Storage) CreateItems(items []Item) bool {
	tx, err := s.db.Begin()
	if err != nil {
		log.Print(err)
		return false
	}

	now := time.Now().UTC()

	itemsSorted := ItemList(items)
	sort.Sort(itemsSorted)

	for _, item := range itemsSorted {
		_, err = tx.Exec(`
			insert into items (
				guid, feed_id, title, link, date,
				content, image, podcast_url,
				date_arrived, status
			)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			on conflict (feed_id, guid) do nothing`,
			item.GUID, item.FeedId, item.Title, item.Link, item.Date,
			item.Content, item.ImageURL, item.AudioURL,
			now, UNREAD,
		)
		if err != nil {
			log.Print(err)
			if err = tx.Rollback(); err != nil {
				log.Print(err)
				return false
			}
			return false
		}
	}
	if err = tx.Commit(); err != nil {
		log.Print(err)
		return false
	}
	return true
}

func listQueryPredicate(filter ItemFilter, newestFirst bool) (string, []interface{}) {
	cond := make([]string, 0)
	args := make([]interface{}, 0)
	index := 1
	if filter.FolderID != nil {
		cond = append(cond, fmt.Sprintf("i.feed_id in (select id from feeds where folder_id = $%d)", index))
		args = append(args, *filter.FolderID)
		index++
	}
	if filter.FeedID != nil {
		cond = append(cond, fmt.Sprintf("i.feed_id = $%d", index))
		args = append(args, *filter.FeedID)
		index++
	}
	if filter.Status != nil {
		cond = append(cond, fmt.Sprintf("i.status = $%d", index))
		args = append(args, *filter.Status)
		index++
	}
	if filter.Search != nil {
		words := strings.Fields(*filter.Search)
		terms := make([]string, len(words))
		for idx, word := range words {
			terms[idx] = word + "*"
		}

		cond = append(cond, fmt.Sprintf("i.search_rowid in (select rowid from search where search match $%d)", index))
		args = append(args, strings.Join(terms, " "))
		index++
	}
	if filter.After != nil {
		compare := ">"
		if newestFirst {
			compare = "<"
		}
		cond = append(cond, fmt.Sprintf("(i.date, i.id) %s (select date, id from items where id = $%d)", compare, index))
		args = append(args, *filter.After)
		index++
	}
	if filter.IDs != nil && len(*filter.IDs) > 0 {
		qmarks := make([]string, len(*filter.IDs))
		idargs := make([]interface{}, len(*filter.IDs))
		for i, id := range *filter.IDs {
			qmarks[i] = fmt.Sprintf("$%d", index)
			idargs[i] = id
			index++
		}
		cond = append(cond, "i.id in ("+strings.Join(qmarks, ",")+")")
		args = append(args, idargs...)
	}
	if filter.SinceID != nil {
		cond = append(cond, fmt.Sprintf("i.id > $%d)", index))
		args = append(args, filter.SinceID)
		index++
	}
	if filter.MaxID != nil {
		cond = append(cond, fmt.Sprintf("i.id < $%d", index))
		args = append(args, filter.MaxID)
		index++
	}
	if filter.Before != nil {
		cond = append(cond, fmt.Sprintf("i.date < $%d", index))
		args = append(args, filter.Before)
		index++
	}

	predicate := "true"
	if len(cond) > 0 {
		predicate = strings.Join(cond, " and ")
	}

	return predicate, args
}

func (s *Storage) CountItems(filter ItemFilter) int {
	predicate, args := listQueryPredicate(filter, false)

	var count int
	query := fmt.Sprintf(`
		select count(*)
		from items
		where %s
		`, predicate)
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		log.Print(err)
		return 0
	}
	return count
}

func (s *Storage) ListItems(filter ItemFilter, limit int, newestFirst bool, withContent bool) []Item {
	predicate, args := listQueryPredicate(filter, newestFirst)
	result := make([]Item, 0, 0)

	order := "date desc, id desc"
	if !newestFirst {
		order = "date asc, id asc"
	}
	if filter.IDs != nil || filter.SinceID != nil {
		order = "i.id asc"
	}
	if filter.MaxID != nil {
		order = "i.id desc"
	}

	selectCols := "i.id, i.guid, i.feed_id, i.title, i.link, i.date, i.status, i.image, i.podcast_url"
	if withContent {
		selectCols += ", i.content"
	} else {
		selectCols += ", '' as content"
	}
	query := fmt.Sprintf(`
		select %s
		from items i
		where %s
		order by %s
		limit %d
		`, selectCols, predicate, order, limit)
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Print(err)
		return result
	}
	for rows.Next() {
		var x Item
		err = rows.Scan(
			&x.Id, &x.GUID, &x.FeedId,
			&x.Title, &x.Link, &x.Date,
			&x.Status, &x.ImageURL, &x.AudioURL, &x.Content,
		)
		if err != nil {
			log.Print(err)
			return result
		}
		result = append(result, x)
	}
	return result
}

func (s *Storage) GetItem(id int64) *Item {
	i := &Item{}
	err := s.db.QueryRow(`
		select
			i.id, i.guid, i.feed_id, i.title, i.link, i.content,
			i.date, i.status, i.image, i.podcast_url
		from items i
		where i.id = $1
	`, id).Scan(
		&i.Id, &i.GUID, &i.FeedId, &i.Title, &i.Link, &i.Content,
		&i.Date, &i.Status, &i.ImageURL, &i.AudioURL,
	)
	if err != nil {
		log.Print(err)
		return nil
	}
	return i
}

func (s *Storage) UpdateItemStatus(item_id int64, status ItemStatus) bool {
	_, err := s.db.Exec(`update items set status = $1 where id = $2`, status, item_id)
	return err == nil
}

func (s *Storage) MarkItemsRead(filter MarkFilter) bool {
	predicate, args := listQueryPredicate(ItemFilter{
		FolderID: filter.FolderID,
		FeedID:   filter.FeedID,
		Before:   filter.Before,
	}, false)
	query := fmt.Sprintf(`
		update items as i set status = %d
		where %s and i.status != %d
		`, READ, predicate, STARRED)
	_, err := s.db.Exec(query, args...)
	if err != nil {
		log.Print(err)
	}
	return err == nil
}

type FeedStat struct {
	FeedId       int64 `json:"feed_id"`
	UnreadCount  int64 `json:"unread"`
	StarredCount int64 `json:"starred"`
}

func (s *Storage) FeedStats() []FeedStat {
	result := make([]FeedStat, 0)
	rows, err := s.db.Query(fmt.Sprintf(`
		select
			feed_id,
			sum(case status when %d then 1 else 0 end),
			sum(case status when %d then 1 else 0 end)
		from items
		group by feed_id
	`, UNREAD, STARRED))
	if err != nil {
		log.Print(err)
		return result
	}
	for rows.Next() {
		stat := FeedStat{}
		rows.Scan(&stat.FeedId, &stat.UnreadCount, &stat.StarredCount)
		result = append(result, stat)
	}
	return result
}

func (s *Storage) SyncSearch() {
	rows, err := s.db.Query(`
		select id, title, content
		from items
		where search_rowid is null;
	`)
	if err != nil {
		log.Print(err)
		return
	}

	items := make([]Item, 0)
	for rows.Next() {
		var item Item
		rows.Scan(&item.Id, &item.Title, &item.Content)
		items = append(items, item)
	}

	for _, item := range items {
		result, err := s.db.Exec(`
			insert into search (title, description, content) values ($1, '', $2)`,
			item.Title, htmlutil.ExtractText(item.Content),
		)
		if err != nil {
			log.Print(err)
			return
		}
		if numrows, err := result.RowsAffected(); err == nil && numrows == 1 {
			if rowId, err := result.LastInsertId(); err == nil {
				s.db.Exec(
					`update items set search_rowid = $1 where id = $2`,
					rowId, item.Id,
				)
			}
		}
	}
}

var (
	itemsKeepSize = 50
	itemsKeepDays = 90
)

// Delete old articles from the database to cleanup space.
//
// The rules:
//   - Never delete starred entries.
//   - Keep at least the same amount of articles the feed provides (default: 50).
//     This prevents from deleting items for rarely updated and/or ever-growing
//     feeds which might eventually reappear as unread.
//   - Keep entries for a certain period (default: 90 days).
func (s *Storage) DeleteOldItems() {
	rows, err := s.db.Query(`
		select
			i.feed_id,
			max(coalesce(s.size, 0)) as max_items,
			count(*) as num_items
		from items i
		left outer join feed_sizes s on s.feed_id = i.feed_id
		where status != $1
		group by i.feed_id
	`, STARRED)

	if err != nil {
		log.Print(err)
		return
	}

	feedLimits := make(map[int64]int64, 0)
	for rows.Next() {
		var feedId, limit int64
		rows.Scan(&feedId, &limit, nil)
		// set default limit if not set
		if limit == 0 {
			limit = int64(itemsKeepSize)
		}
		feedLimits[feedId] = limit
	}

	for feedId, limit := range feedLimits {
		result, err := s.db.Exec(`
			delete from items
			where id in (
				select i.id
				from items i
				where i.feed_id = $1 and status != $2
				order by date desc
				offset $3
			) and date_arrived < $4
			`,
			feedId,
			STARRED,
			limit,
			time.Now().UTC().Add(-time.Hour*time.Duration(24*itemsKeepDays)),
		)
		if err != nil {
			log.Print(err)
			return
		}
		numDeleted, err := result.RowsAffected()
		if err != nil {
			log.Print(err)
			return
		}
		if numDeleted > 0 {
			log.Printf("Deleted %d old items (feed: %d)", numDeleted, feedId)
		}
	}
}
