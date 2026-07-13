package racing

import (
	"context"
	"database/sql"
	"time"
)

// SeasonWeeks is the length of an iRacing season.
const SeasonWeeks = 13

// currentWeek derives the running season week from the REAL race dates in
// season_schedule: the latest week whose race has started. Falls back to
// ISO-week arithmetic only when no dated schedule exists yet.
func (s *Service) currentWeek(ctx context.Context, now time.Time) int {
	var week sql.NullInt64
	err := s.db.QueryRowContext(ctx, `
		SELECT MAX(week) FROM season_schedule
		WHERE race_date IS NOT NULL AND race_date <= ?`, now.Format("2006-01-02")).Scan(&week)
	if err == nil && week.Valid && week.Int64 >= 1 {
		return int(week.Int64)
	}
	_, isoWeek := now.ISOWeek()
	return (isoWeek-1)%SeasonWeeks + 1
}

// SeasonWeek is one cell of the season grid. TrackAccess is "free" (default
// content, green), "owned" (purchased/unlocked, aquamarine) or "missing" (red);
// TrackOwned means available at all (free or owned). Weeks a series doesn't
// race are simply absent (real schedules have gaps, unlike the old rotation).
type SeasonWeek struct {
	Week        int    `json:"week"`
	TrackID     int    `json:"trackId"`
	TrackName   string `json:"trackName"`
	ConfigName  string `json:"configName"`
	RaceDate    string `json:"raceDate,omitempty"` // YYYY-MM-DD, week start
	TrackOwned  bool   `json:"trackOwned"`
	TrackAccess Access `json:"trackAccess"`
	Planned     bool   `json:"planned"`
}

// SetRacePlanned marks (or unmarks) a series+week as a race the user will run.
func (s *Service) SetRacePlanned(ctx context.Context, userID int64, seriesID, week int, planned bool) error {
	if planned {
		_, err := s.db.ExecContext(ctx,
			`INSERT IGNORE INTO planned_races (user_id, series_id, week) VALUES (?, ?, ?)`,
			userID, seriesID, week)
		return err
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM planned_races WHERE user_id = ? AND series_id = ? AND week = ?`,
		userID, seriesID, week)
	return err
}

// SeasonSeries is one row: a series and its season calendar, annotated with
// what the user owns.
type SeasonSeries struct {
	SeriesID      int          `json:"seriesId"`
	SeriesName    string       `json:"seriesName"`
	Category      string       `json:"category"`
	LicenseNeeded string       `json:"licenseNeeded"`
	Favorite      bool         `json:"favorite"`
	CarOwned      bool         `json:"carOwned"` // owns ≥1 eligible car (series_cars mapping)
	Weeks         []SeasonWeek `json:"weeks"`
}

// Season is the full season view.
type Season struct {
	CurrentWeek int            `json:"currentWeek"`
	TotalWeeks  int            `json:"totalWeeks"`
	Series      []SeasonSeries `json:"series"`
}

// SeasonView assembles the season grid for a user: every series with its
// weekly tracks, ownership flags, and whether the user owns an eligible car.
// Car eligibility uses the precise series_cars mapping when present, falling
// back to the category approximation for series without one.
func (s *Service) SeasonView(ctx context.Context, userID int64) (Season, error) {
	access, err := s.trackAccess(ctx, userID)
	if err != nil {
		return Season{}, err
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT sr.series_id, sr.series_name, sr.category, sr.license_needed,
		       f.series_id IS NOT NULL AS favorite,
		       CASE
		           WHEN EXISTS (SELECT 1 FROM series_cars sc WHERE sc.series_id = sr.series_id) THEN
		               EXISTS (
		                   SELECT 1 FROM series_cars sc
		                   JOIN cars c ON c.car_id = sc.car_id
		                   LEFT JOIN owned_cars oc ON oc.car_id = c.car_id AND oc.user_id = ?
		                   WHERE sc.series_id = sr.series_id AND (c.is_free = 1 OR oc.car_id IS NOT NULL)
		               )
		           ELSE
		               EXISTS (
		                   SELECT 1 FROM cars c
		                   LEFT JOIN owned_cars oc ON oc.car_id = c.car_id AND oc.user_id = ?
		                   WHERE c.category = sr.category AND (c.is_free = 1 OR oc.car_id IS NOT NULL)
		               )
		       END AS car_owned,
		       ss.week, ss.track_id, t.track_name, COALESCE(t.config_name, ''),
		       COALESCE(DATE_FORMAT(ss.race_date, '%Y-%m-%d'), ''),
		       pr.week IS NOT NULL AS planned
		FROM series sr
		JOIN season_schedule ss ON ss.series_id = sr.series_id
		JOIN tracks t ON t.track_id = ss.track_id
		LEFT JOIN favorite_series f ON f.series_id = sr.series_id AND f.user_id = ?
		LEFT JOIN planned_races pr ON pr.series_id = sr.series_id AND pr.week = ss.week AND pr.user_id = ?
		ORDER BY sr.series_name, ss.week`, userID, userID, userID, userID)
	if err != nil {
		return Season{}, err
	}
	defer rows.Close()

	season := Season{CurrentWeek: s.currentWeek(ctx, time.Now()), TotalWeeks: SeasonWeeks}
	index := make(map[int]int) // series_id -> position in season.Series

	for rows.Next() {
		var (
			sr SeasonSeries
			w  SeasonWeek
		)
		if err := rows.Scan(&sr.SeriesID, &sr.SeriesName, &sr.Category, &sr.LicenseNeeded,
			&sr.Favorite, &sr.CarOwned,
			&w.Week, &w.TrackID, &w.TrackName, &w.ConfigName, &w.RaceDate, &w.Planned); err != nil {
			return Season{}, err
		}
		w.TrackAccess = AccessMissing
		if a, ok := access[w.TrackID]; ok {
			w.TrackAccess = a
		}
		w.TrackOwned = w.TrackAccess != AccessMissing
		pos, ok := index[sr.SeriesID]
		if !ok {
			pos = len(season.Series)
			index[sr.SeriesID] = pos
			season.Series = append(season.Series, sr)
		}
		season.Series[pos].Weeks = append(season.Series[pos].Weeks, w)
	}
	return season, rows.Err()
}
