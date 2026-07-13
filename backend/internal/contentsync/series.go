package contentsync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"
)

const defaultSeriesURL = "https://raw.githubusercontent.com/adrianulima/my-racing-planner/main/src/ir-data/series.json"

// seasonWeeks mirrors racing.SeasonWeeks (not imported to keep contentsync
// dependency-free of the app service).
const seasonWeeks = 13

type seriesEntry struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Category string   `json:"category"`
	Cars     []int    `json:"cars"`
	License  struct {
		Letter string `json:"letter"`
	} `json:"license"`
	Weeks []struct {
		WeekNum int    `json:"weekNum"`
		Date    string `json:"date"`
		Track   struct {
			ID int `json:"id"`
		} `json:"track"`
	} `json:"weeks"`
}

var categoryIDs = map[string]int{
	"oval": 1, "road": 2, "dirt_oval": 3, "dirt_road": 4, "sports_car": 5, "formula_car": 6,
}

// syncSeries refreshes series, the series→car mapping, and the REAL season
// schedule from the community series list. Long-running (year-round) series
// are windowed to the current 13-week season, matching the planner grid.
func (s *Syncer) syncSeries(ctx context.Context, url string) error {
	data, hash, fresh, err := s.fetch(ctx, url)
	if err != nil || !fresh {
		return err
	}
	var series map[string]seriesEntry
	if err := json.Unmarshal(data, &series); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	start, ok := seasonStart(series)
	if !ok {
		return fmt.Errorf("no season window derivable from %d series", len(series))
	}
	end := start.AddDate(0, 0, seasonWeeks*7)

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	updated, scheduled := 0, 0
	for _, se := range series {
		if se.ID == 0 || se.Name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO series (series_id, series_name, category, category_id, license_needed)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE series_name = VALUES(series_name),
				category = VALUES(category), category_id = VALUES(category_id),
				license_needed = VALUES(license_needed)`,
			se.ID, se.Name, se.Category, categoryIDs[se.Category], se.License.Letter); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM series_cars WHERE series_id = ?`, se.ID); err != nil {
			return err
		}
		for _, carID := range se.Cars {
			if _, err := tx.ExecContext(ctx,
				`INSERT IGNORE INTO series_cars (series_id, car_id) VALUES (?, ?)`, se.ID, carID); err != nil {
				return err
			}
		}

		// Rewrite the series' schedule from the in-window races.
		if _, err := tx.ExecContext(ctx,
			`DELETE FROM season_schedule WHERE series_id = ?`, se.ID); err != nil {
			return err
		}
		for _, w := range se.Weeks {
			d, err := time.Parse("2006-01-02", w.Date)
			if err != nil || d.Before(start) || !d.Before(end) {
				continue
			}
			week := int(d.Sub(start).Hours()/(24*7)) + 1
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO season_schedule (series_id, week, track_id, race_date) VALUES (?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE track_id = VALUES(track_id), race_date = VALUES(race_date)`,
				se.ID, week, w.Track.ID, w.Date); err != nil {
				return err
			}
			scheduled++
		}
		updated++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("contentsync: %d series upserted, %d races scheduled (season %s) from %s",
		updated, scheduled, start.Format("2006-01-02"), url)
	return s.record(ctx, hash, "series:"+url, len(series), updated)
}

// seasonStart picks the modal first-week date among full-length (13-week)
// series — that's the current season's week 1.
func seasonStart(series map[string]seriesEntry) (time.Time, bool) {
	counts := map[string]int{}
	for _, se := range series {
		if len(se.Weeks) != seasonWeeks {
			continue
		}
		first := se.Weeks[0].Date
		for _, w := range se.Weeks { // weeks may be unordered; find the min date
			if strings.Compare(w.Date, first) < 0 {
				first = w.Date
			}
		}
		counts[first]++
	}
	if len(counts) == 0 {
		return time.Time{}, false
	}
	dates := make([]string, 0, len(counts))
	for d := range counts {
		dates = append(dates, d)
	}
	sort.Slice(dates, func(i, j int) bool {
		if counts[dates[i]] != counts[dates[j]] {
			return counts[dates[i]] > counts[dates[j]]
		}
		return dates[i] > dates[j] // tie → most recent season
	})
	t, err := time.Parse("2006-01-02", dates[0])
	return t, err == nil
}
