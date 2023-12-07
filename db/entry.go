package db

import (
	"database/sql"
	"log"
	"time"
)

type Entry struct {
	CreatedAt          time.Time
	UpdatedAt          time.Time
	FirstFeedStartTime time.Time
	LastFeedEndTime    time.Time
	Id                 string
	Notes              string
	Feeds              []Feed
	TotalFeedDuration  time.Duration
	NappyStateDirty    int
	NappyStateWet      bool
}

func (e Entry) GetFeedByFeedId(feedId int) *Feed {
	for _, feed := range e.Feeds {
		if feed.Id == feedId {
			return &feed
		}
	}

	return nil
}

type Feed struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	StartTime time.Time
	EndTime   time.Time
	Side      string
	Duration  time.Duration
	Id        int
}

func UpsertEntry(newEntry Entry) {
	existingEntry := GetEntryById(newEntry.Id)

	if existingEntry == nil {
		newEntry.CreatedAt = time.Now()
	}

	sql := `
    INSERT OR REPLACE INTO entries 
    (id, notes, nappy_state_dirty, nappy_state_wet, updated_at, created_at)
		VALUES 
    (?, ?, ?, ?, ?, ?)
  `

	stmt, err := Db.Prepare(sql)
	if err != nil {
		log.Printf("Failed to parse update entry SQL")
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(
		newEntry.Id,
		newEntry.Notes,
		newEntry.NappyStateDirty,
		newEntry.NappyStateWet,
		time.Now(),
		newEntry.CreatedAt,
	)

	// Exit if we get an error
	if err2 != nil {
		log.Printf("Failed to create or update entry")
		panic(err2)
	}

	log.Printf("Created new entry: " + newEntry.Id)
}

func CreateFeedForEntry(entry *Entry, newFeed Feed) {
	sql := `
    INSERT INTO feeds (entry_id, start_time, end_time, side, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?)
  `

	stmt, err := Db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(entry.Id, newFeed.StartTime, newFeed.EndTime, newFeed.Side, time.Now(), time.Now())
	log.Print("Inserted feed for entry: " + entry.Id)

	if err2 != nil {
		panic(err2)
	}
}

func UpdateFeed(feed *Feed) {
	sql := `
  UPDATE feeds
  SET start_time = ?, end_time = ?, side = ?, updated_at = ?
  WHERE id = ?;
  `

	stmt, err := Db.Prepare(sql)
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	_, err2 := stmt.Exec(feed.StartTime, feed.EndTime, feed.Side, time.Now(), feed.Id)
	log.Printf("Updated feed: %d", feed.Id)

	if err2 != nil {
		panic(err2)
	}
}

func GetEntryById(id string) *Entry {
	sql := `
  SELECT
    e.id,
    e.nappy_state_wet,
    e.nappy_state_dirty,
    e.notes,
    e.created_at,
    e.updated_at,
    f.id,
    f.start_time,
    f.end_time,
    f.side,
    f.duration,
    fg.total_duration,
    fg.started,
    fg.ended,
    f.created_at,
    f.updated_at
  FROM entries e
  LEFT JOIN (
    SELECT *, unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19)) AS duration FROM feeds
  ) f ON e.id = f.entry_id
  LEFT JOIN (
    SELECT
      entry_id, 
      MAX(DATETIME(SUBSTR(end_time, 1, 19))) as ended,
      MIN(DATETIME(SUBSTR(start_time, 1, 19))) as started,
      SUM(unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19))) as total_duration
    FROM feeds
    GROUP BY entry_id
  ) fg on e.id = fg.entry_id
  WHERE e.id = ?
  `

	log.Printf("Finding entity")
	rows, err := Db.Query(sql, id)
	if err != nil {
		log.Printf("Error searching for entity: " + id)
		return nil
	}

	defer rows.Close()
	log.Printf("Converting entities")
	entries := convertRowsToEntries(rows)
	if len(entries) == 0 {
		log.Printf("Could not convert entries")
		return nil
	}

	entry := entries[0]
	log.Printf("Found entity: " + entry.Id)
	return &entry
}

func GetAllEntries(limit int, offset int) []Entry {
	sql := `
  SELECT
    e.id,
    e.nappy_state_wet,
    e.nappy_state_dirty,
    e.notes,
    e.created_at,
    e.updated_at,
    f.id,
    f.start_time,
    f.end_time,
    f.side,
    f.duration,
    fg.total_duration,
    fg.started,
    fg.ended,
    f.created_at,
    f.updated_at
  FROM (
    SELECT * FROM entries ORDER BY created_at DESC LIMIT ? OFFSET ?
  ) AS e
  LEFT JOIN (
    SELECT *, unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19)) AS duration FROM feeds
  ) f ON e.id = f.entry_id
  LEFT JOIN (
    SELECT
      entry_id, 
      MAX(DATETIME(SUBSTR(end_time, 1, 19))) as ended,
      MIN(DATETIME(SUBSTR(start_time, 1, 19))) as started,
      SUM(unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19))) as total_duration
    FROM feeds
    GROUP BY entry_id
  ) fg on e.id = fg.entry_id
  ORDER BY e.created_at DESC, f.start_time DESC;
  `

	rows, err := Db.Query(sql, limit, offset)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	entries := convertRowsToEntries(rows)
	return entries
}

func GetFeedByEntryIdAndFeedId(entryId string, feedId int) *Feed {
	sql := `
  SELECT
    e.id,
    e.nappy_state_wet,
    e.nappy_state_dirty,
    e.notes,
    e.created_at,
    e.updated_at,
    f.id,
    f.start_time,
    f.end_time,
    f.side,
    f.duration,
    fg.total_duration,
    fg.started,
    fg.ended,
    f.created_at,
    f.updated_at
  FROM entries e
  LEFT JOIN (
    SELECT *, unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19)) AS duration FROM feeds
  ) f ON e.id = f.entry_id
  LEFT JOIN (
    SELECT
      entry_id, 
      MAX(DATETIME(SUBSTR(end_time, 1, 19))) as ended,
      MIN(DATETIME(SUBSTR(start_time, 1, 19))) as started,
      SUM(unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19))) as total_duration
    FROM feeds
    GROUP BY entry_id
  ) fg on e.id = fg.entry_id
  WHERE f.entry_id = ? AND f.id = ?
  LIMIT 1
  `

	rows, err := Db.Query(sql, entryId, feedId)
	if err != nil {
		log.Printf("Error searching for entity by feed Id: %d", feedId)
		return nil
	}

	defer rows.Close()
	log.Printf("Converting entities")
	entries := convertRowsToEntries(rows)
	if len(entries) == 0 {
		log.Printf("Could not convert entries with entry id '%s', feed id '%d'", entryId, feedId)
		return nil
	}

	entry := entries[0]
	if len(entry.Feeds) == 0 {
		log.Printf("Could not find feed '%d' for entity Id '%s'", feedId, entry.Id)
		return nil
	}

	feed := entry.Feeds[0]
	log.Printf("Found entity %s with feed Id %d", entry.Id, feed.Id)
	return &feed
}

func convertRowsToEntries(rows *sql.Rows) []Entry {
	entriesMap := make(map[string]*Entry)
	var entryKeys []string

	for rows.Next() {
		var (
			entryCreatedAt, entryUpdatedAt                   time.Time
			id                                               string
			notes                                            string
			nappyStateDirty                                  int
			nappyStateWet                                    bool
			side                                             sql.NullString
			duration, totalFeedDuration                      sql.NullInt64
			rawFirstFeedStartTime, rawLastFeedEndTime        sql.NullString
			feedCreatedAt, feedUpdatedAt, startTime, endTime sql.NullTime
			feedId                                           sql.NullInt64
		)

		err := rows.Scan(
			&id,
			&nappyStateWet,
			&nappyStateDirty,
			&notes,
			&entryCreatedAt,
			&entryUpdatedAt,
			&feedId,
			&startTime,
			&endTime,
			&side,
			&duration,
			&totalFeedDuration,
			&rawFirstFeedStartTime,
			&rawLastFeedEndTime,
			&feedCreatedAt,
			&feedUpdatedAt,
		)
		if err != nil {
			panic(err.Error())
		}

		entry, ok := entriesMap[id]

		if !ok {
			timeLayout := "2006-01-02 15:04:05"
			firstFeedStartTime, _ := time.Parse(timeLayout, rawFirstFeedStartTime.String)
			lastFeedEndTime, _ := time.Parse(timeLayout, rawLastFeedEndTime.String)

			entry = &Entry{
				CreatedAt:          entryCreatedAt,
				UpdatedAt:          entryUpdatedAt,
				Id:                 id,
				Notes:              notes,
				NappyStateDirty:    nappyStateDirty,
				FirstFeedStartTime: firstFeedStartTime,
				LastFeedEndTime:    lastFeedEndTime,
				TotalFeedDuration:  time.Duration(int(totalFeedDuration.Int64) * int(time.Second)),
				NappyStateWet:      nappyStateWet,
			}
			entriesMap[id] = entry
			entryKeys = append(entryKeys, id)
		}

		if feedId.Valid {
			entry.Feeds = append(entry.Feeds, Feed{
				CreatedAt: feedCreatedAt.Time,
				UpdatedAt: feedUpdatedAt.Time,
				Duration:  time.Duration(int(duration.Int64) * int(time.Second)),
				StartTime: startTime.Time,
				EndTime:   endTime.Time,
				Side:      side.String,
				Id:        int(feedId.Int64),
			})
		}
	}

	var entries []Entry
	for _, entryId := range entryKeys {
		entry := *entriesMap[entryId]
		entries = append(entries, entry)
	}

	// log.Printf("Recieved %d entries", len(entries))
	return entries
}

func GetLastEntry() *Entry {
	entries := GetAllEntries(1, 0)
	return &entries[0]
}

func GetLastLoggedSide() string {
	var side string

	sql := `SELECT side FROM feeds ORDER BY start_time DESC LIMIT 1;`
	row := Db.QueryRow(sql)

	err := row.Scan(&side)
	if err != nil {
		return ""
	}

	return side
}

func GetNextSessionStartTime() *time.Time {
	var rawStartTime sql.NullString

	sql := `
  SELECT
    MIN(f.start_time)
  FROM entries e
  LEFT JOIN feeds f ON e.id = f.entry_id
  WHERE f.start_time IS NOT NULL
  GROUP BY e.id
  ORDER BY e.created_at DESC
  LIMIT 1;
  `

	row := Db.QueryRow(sql)

	err := row.Scan(&rawStartTime)
	if err == nil && rawStartTime.Valid {
		log.Printf(rawStartTime.String)
		startTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", rawStartTime.String)
		startTime = startTime.Add(3 * time.Hour)

		if err != nil {
			panic(err)
		}

		return &startTime
	}

	return nil
}

func DeleteEntryById(id string) bool {
	sql := `
    DELETE FROM entries WHERE id = ?;
    DELETE FROM feeds WHERE entry_id = ?;
  `
	_, err := Db.Exec(sql, id, id)
	if err != nil {
		log.Printf("Error deleting entry Id '%s': %s", id, err.Error())
	}
	return err == nil
}

func DeleteFeedByEntryIdAndFeedId(entryId string, feedId int) bool {
	sql := `
    DELETE FROM feeds WHERE entry_id = ? AND id = ?;
  `
	_, err := Db.Exec(sql, entryId, feedId)
	if err != nil {
		log.Printf("Error deleting feed Id '%d' for entry Id '%s': %s", feedId, entryId, err.Error())
	}
	return err == nil
}
