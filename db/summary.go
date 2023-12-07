package db

import (
	"log"
	"time"
)

type DailySummary struct {
	Date                   time.Time
	DirtyNappyMap          map[string]int
	TotalEntries           int
	WetCount               int
	NotWetCount            int
	WetAndDirtyCount       int
	DirtyStreakCount       int
	DirtyStainedCount      int
	DirtyRegularCount      int
	DirtyHeavyCount        int
	DirtyPoonamiCount      int
	DirtyTotalCount        int
	IsToday                bool
	TotalFeedsCount        int
	LeftFeedsCount         int
	RightFeedsCount        int
	TotalFeedDuration      time.Duration
	AvgFeedDuration        time.Duration
	MaxFeedDuration        time.Duration
	MinFeedDuration        time.Duration
	TotalRightFeedDuration time.Duration
	AvgRightFeedDuration   time.Duration
	MaxRightFeedDuration   time.Duration
	MinRightFeedDuration   time.Duration
	TotalLeftFeedDuration  time.Duration
	AvgLeftFeedDuration    time.Duration
	MaxLeftFeedDuration    time.Duration
	MinLeftFeedDuration    time.Duration
}

func GetDailySummaries(daysLimit int, daysOffset int) []DailySummary {
	sql := `
  SELECT
    created_at_date,
    CASE
      WHEN created_at_date IS DATE('now') IS 1 THEN 1
      ELSE 0
    END as is_today,
    COUNT(e.id) as total_entries,
    SUM(
      CASE WHEN e.nappy_state_wet IS 1 
      AND e.nappy_state_dirty > 0 THEN 1 ELSE 0 END
    ) as wet_and_dirty,
    SUM(
      CASE WHEN e.nappy_state_wet IS 1 THEN 1 ELSE 0 END
    ) as wet,
    SUM(
      CASE WHEN e.nappy_state_wet IS 0 THEN 1 ELSE 0 END
    ) as not_wet,
    SUM(
      CASE WHEN e.nappy_state_dirty IS 1 THEN 1 ELSE 0 END
    ) as dirty_streak,
    SUM(
      CASE WHEN e.nappy_state_dirty IS 2 THEN 1 ELSE 0 END
    ) as dirty_stained,
    SUM(
      CASE WHEN e.nappy_state_dirty IS 3 THEN 1 ELSE 0 END
    ) as dirty_regular,
    SUM(
      CASE WHEN e.nappy_state_dirty IS 4 THEN 1 ELSE 0 END
    ) as dirty_heavy,
    SUM(
      CASE WHEN e.nappy_state_dirty IS 5 THEN 1 ELSE 0 END
    ) as dirty_poonami,
    SUM(
      CASE WHEN e.nappy_state_dirty > 0 THEN 1 ELSE 0 END
    ) as dirty_total,
      SUM(feed_count) as feed_count,
      SUM(left_count) as left_count,
      SUM(right_count) as right_count,
    SUM(total_feed_duration) as total_feed_duration,
    ROUND(AVG(avg_feed_duration)) as avg_feed_duration,
    MAX(max_feed_duration) as max_feed_duration,
    MIN(min_feed_duration) as min_feed_duration,
    SUM(total_right_duration) as total_right_duration,
    ROUND(AVG(avg_right_duration)) as avg_right_duration,
    MAX(max_right_duration) as max_right_duration,
    MIN(min_right_duration) as min_right_duration,
    SUM(total_left_duration) as total_left_duration,
    ROUND(AVG(avg_left_duration)) as avg_left_duration,
    MAX(max_left_duration) as max_left_duration,
    MIN(min_left_duration) as min_left_duration
  FROM
    (
    SELECT
      *,
      date(
      SUBSTR(e.created_at, 1, 19)
    ) as created_at_date
    FROM
      entries e
    LEFT JOIN (
      SELECT
        entry_id,
        COUNT(id) as feed_count,
        COUNT(CASE WHEN side = 'right' THEN 1 END) as right_count,
        SUM(right_duration) as total_right_duration,
        ROUND(AVG(right_duration)) as avg_right_duration,
        MAX(right_duration) as max_right_duration,
        MIN(right_duration) as min_right_duration,
        COUNT(CASE WHEN side = 'left' THEN 1 END) as left_count,
        SUM(left_duration) as total_left_duration,
        ROUND(AVG(left_duration)) as avg_left_duration,
        MAX(left_duration) as max_left_duration,
        MIN(left_duration) as min_left_duration,
        SUM(duration) as total_feed_duration,
        ROUND(AVG(duration)) as avg_feed_duration,
        MAX(duration) as max_feed_duration,
        MIN(duration) as min_feed_duration
      FROM
        (
        SELECT
          *,
          CASE
            WHEN side = 'left' THEN duration
            ELSE null
          END as left_duration,
          CASE
            WHEN side = 'right' THEN duration
            ELSE null
          END as right_duration
        FROM
          (
          select
            *,
            unixepoch(SUBSTR(end_time, 1, 19)) - unixepoch(SUBSTR(start_time, 1, 19)) as duration
          FROM
            feeds
      )
    ) f
      GROUP BY
        entry_id
  ) f ON
      e.id = f.entry_id
  ) e
  GROUP BY
    e.created_at_date
  ORDER BY
    e.created_at_date DESC
  LIMIT ?
  OFFSET ?;
  `

	rows, err := Db.Query(sql, daysLimit, daysOffset)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	var dailySummaries []DailySummary
	for rows.Next() {
		var dailySummary DailySummary
		var rawDate string

		var rawTotalFeedDuration int
		var rawAvgFeedDuration int
		var rawMaxFeedDuration int
		var rawMinFeedDuration int
		var rawTotalRightFeedDuration int
		var rawAvgRightFeedDuration int
		var rawMaxRightFeedDuration int
		var rawMinRightFeedDuration int
		var rawTotalLeftFeedDuration int
		var rawAvgLeftFeedDuration int
		var rawMaxLeftFeedDuration int
		var rawMinLeftFeedDuration int

		rows.Scan(
			&rawDate,
			&dailySummary.IsToday,
			&dailySummary.TotalEntries,
			&dailySummary.WetAndDirtyCount,
			&dailySummary.WetCount,
			&dailySummary.NotWetCount,
			&dailySummary.DirtyStreakCount,
			&dailySummary.DirtyStainedCount,
			&dailySummary.DirtyRegularCount,
			&dailySummary.DirtyHeavyCount,
			&dailySummary.DirtyPoonamiCount,
			&dailySummary.DirtyTotalCount,
			&dailySummary.TotalFeedsCount,
			&dailySummary.LeftFeedsCount,
			&dailySummary.RightFeedsCount,
			&rawTotalFeedDuration,
			&rawAvgFeedDuration,
			&rawMaxFeedDuration,
			&rawMinFeedDuration,
			&rawTotalRightFeedDuration,
			&rawAvgRightFeedDuration,
			&rawMaxRightFeedDuration,
			&rawMinRightFeedDuration,
			&rawTotalLeftFeedDuration,
			&rawAvgLeftFeedDuration,
			&rawMaxLeftFeedDuration,
			&rawMinLeftFeedDuration,
		)

		durationFields := []struct {
			dest *time.Duration
			raw  int
		}{
			{&dailySummary.TotalFeedDuration, rawTotalFeedDuration},
			{&dailySummary.AvgFeedDuration, rawAvgFeedDuration},
			{&dailySummary.MaxFeedDuration, rawMaxFeedDuration},
			{&dailySummary.MinFeedDuration, rawMinFeedDuration},
			{&dailySummary.TotalRightFeedDuration, rawTotalRightFeedDuration},
			{&dailySummary.AvgRightFeedDuration, rawAvgRightFeedDuration},
			{&dailySummary.MaxRightFeedDuration, rawMaxRightFeedDuration},
			{&dailySummary.MinRightFeedDuration, rawMinRightFeedDuration},
			{&dailySummary.TotalLeftFeedDuration, rawTotalLeftFeedDuration},
			{&dailySummary.AvgLeftFeedDuration, rawAvgLeftFeedDuration},
			{&dailySummary.MaxLeftFeedDuration, rawMaxLeftFeedDuration},
			{&dailySummary.MinLeftFeedDuration, rawMinLeftFeedDuration},
		}

		for _, field := range durationFields {
			*field.dest = time.Duration(field.raw) * time.Second
		}

		dailySummary.Date, _ = time.Parse("2006-01-02", rawDate)
		log.Printf("Summary date: %s", dailySummary.Date.String())
		dailySummaries = append(dailySummaries, dailySummary)
	}

	return dailySummaries
}
