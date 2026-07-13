package racing

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"log"
	"strings"
	"unicode"
)

// catalog_seed.json is generated from my-racing-planner's exports of the
// official iRacing data (real ids, prices, free flags, and the current
// season's real weekly schedules). Regenerate with
// backend/scripts/gen-catalog-seed.py; internal/contentsync keeps it fresh at
// runtime between regenerations.
//
//go:embed catalog_seed.json
var catalogSeed []byte

type seedData struct {
	SeasonStart string `json:"seasonStart"`
	Cars        []struct {
		CarID       int     `json:"carId"`
		CarName     string  `json:"carName"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Free        bool    `json:"free"`
		Price       float64 `json:"price"`
	} `json:"cars"`
	Tracks []struct {
		TrackID     int     `json:"trackId"`
		TrackName   string  `json:"trackName"`
		ConfigName  string  `json:"configName"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Free        bool    `json:"free"`
		Price       float64 `json:"price"`
		SkuGroup    int     `json:"skuGroup"`
	} `json:"tracks"`
	Series []struct {
		SeriesID      int    `json:"seriesId"`
		SeriesName    string `json:"seriesName"`
		Category      string `json:"category"`
		CategoryID    int    `json:"categoryId"`
		LicenseNeeded string `json:"licenseNeeded"`
		Description   string `json:"description"`
		Cars          []int  `json:"cars"`
		Weeks         []struct {
			Week    int    `json:"week"`
			TrackID int    `json:"trackId"`
			Date    string `json:"date"`
		} `json:"weeks"`
	} `json:"series"`
	// trackRequirements models combined layouts needing several purchases,
	// e.g. Nürburgring Combined requires the Nordschleife AND the GP circuit.
	TrackRequirements map[string][]int `json:"trackRequirements"`
}

// SeedCatalog upserts the bundled catalog (cars, tracks, series, series→car
// mapping, the real season schedule, and combined-layout requirements) so the
// planner works without any external source. Row updates are NAME-GUARDED: if
// a row's name was changed by the authoritative content sync, the seed leaves
// that row alone. Rows whose ids predate the real-id catalog are migrated by
// name (ownership/favorites/plans follow) and then removed.
func SeedCatalog(ctx context.Context, db *sql.DB) error {
	var data seedData
	if err := json.Unmarshal(catalogSeed, &data); err != nil {
		return err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	for _, c := range data.Cars {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cars (car_id, car_name, category, description, is_free, price)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
				category    = IF(car_name = VALUES(car_name), VALUES(category), category),
				description = IF(car_name = VALUES(car_name) AND VALUES(description) <> '', VALUES(description), description),
				is_free     = IF(car_name = VALUES(car_name), VALUES(is_free), is_free),
				price       = IF(car_name = VALUES(car_name), VALUES(price), price)`,
			c.CarID, c.CarName, c.Category, c.Description, c.Free, c.Price); err != nil {
			return err
		}
	}
	for _, t := range data.Tracks {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tracks (track_id, track_name, config_name, category, description, is_free, price, sku_group)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
				config_name = IF(track_name = VALUES(track_name), VALUES(config_name), config_name),
				category    = IF(track_name = VALUES(track_name), VALUES(category), category),
				description = IF(track_name = VALUES(track_name) AND VALUES(description) <> '', VALUES(description), description),
				is_free     = IF(track_name = VALUES(track_name), VALUES(is_free), is_free),
				price       = IF(track_name = VALUES(track_name), VALUES(price), price),
				sku_group   = IF(track_name = VALUES(track_name), VALUES(sku_group), sku_group)`,
			t.TrackID, t.TrackName, t.ConfigName, t.Category, t.Description, t.Free, t.Price, t.SkuGroup); err != nil {
			return err
		}
	}
	for _, s := range data.Series {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO series (series_id, series_name, category, category_id, license_needed, description)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE
				category       = IF(series_name = VALUES(series_name), VALUES(category), category),
				category_id    = IF(series_name = VALUES(series_name), VALUES(category_id), category_id),
				license_needed = IF(series_name = VALUES(series_name), VALUES(license_needed), license_needed),
				description    = IF(series_name = VALUES(series_name) AND VALUES(description) <> '', VALUES(description), description)`,
			s.SeriesID, s.SeriesName, s.Category, s.CategoryID, s.LicenseNeeded, s.Description); err != nil {
			return err
		}
		for _, carID := range s.Cars {
			if _, err := tx.ExecContext(ctx,
				`INSERT IGNORE INTO series_cars (series_id, car_id) VALUES (?, ?)`,
				s.SeriesID, carID); err != nil {
				return err
			}
		}
		// The REAL season schedule (weeks the series doesn't race stay empty).
		for _, w := range s.Weeks {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO season_schedule (series_id, week, track_id, race_date) VALUES (?, ?, ?, ?)
				ON DUPLICATE KEY UPDATE track_id = VALUES(track_id), race_date = VALUES(race_date)`,
				s.SeriesID, w.Week, w.TrackID, w.Date); err != nil {
				return err
			}
		}
	}
	for trackID, reqs := range data.TrackRequirements {
		for _, req := range reqs {
			if _, err := tx.ExecContext(ctx,
				`INSERT IGNORE INTO track_requirements (track_id, requires_track_id) VALUES (?, ?)`,
				trackID, req); err != nil {
				return err
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return reconcileCatalog(ctx, db, data)
}

// reconcileCatalog migrates rows left over from the pre-real-id seed: any
// car/track/series whose id is NOT in the authoritative set but whose name
// matches an authoritative row gets its user references (ownership, favorites,
// planned races) moved to the real id; the stale rows and their fabricated
// schedule rows are then deleted. Idempotent — after the first run there is
// nothing left to move.
func reconcileCatalog(ctx context.Context, db *sql.DB, data seedData) error {
	carNames := make(map[string]int, len(data.Cars))
	carIDs := make(map[int]bool, len(data.Cars))
	for _, c := range data.Cars {
		carNames[normalizeName(c.CarName)] = c.CarID
		carIDs[c.CarID] = true
	}
	trackNames := make(map[string]int, len(data.Tracks))
	trackIDs := make(map[int]bool, len(data.Tracks))
	for _, t := range data.Tracks {
		trackNames[normalizeName(t.TrackName+" "+t.ConfigName)] = t.TrackID
		if _, ok := trackNames[normalizeName(t.TrackName)]; !ok {
			trackNames[normalizeName(t.TrackName)] = t.TrackID // base-name fallback
		}
		trackIDs[t.TrackID] = true
	}
	seriesNames := make(map[string]int, len(data.Series))
	seriesIDs := make(map[int]bool, len(data.Series))
	for _, s := range data.Series {
		seriesNames[normalizeName(s.SeriesName)] = s.SeriesID
		seriesIDs[s.SeriesID] = true
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	moved, removed := 0, 0
	migrate := func(query string, ids map[int]bool, names map[string]int, refs []string, idCol, table string) error {
		rows, err := tx.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		type stale struct {
			id   int
			name string
		}
		var stales []stale
		for rows.Next() {
			var s stale
			if err := rows.Scan(&s.id, &s.name); err != nil {
				rows.Close()
				return err
			}
			if !ids[s.id] {
				stales = append(stales, s)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		for _, s := range stales {
			if newID, ok := names[normalizeName(s.name)]; ok {
				for _, ref := range refs {
					// UPDATE IGNORE: if the user already has the real id too,
					// the duplicate is skipped and removed by the DELETE below.
					if _, err := tx.ExecContext(ctx,
						"UPDATE IGNORE "+ref+" SET "+idCol+" = ? WHERE "+idCol+" = ?", newID, s.id); err != nil {
						return err
					}
				}
				moved++
			}
			for _, ref := range refs {
				if _, err := tx.ExecContext(ctx, "DELETE FROM "+ref+" WHERE "+idCol+" = ?", s.id); err != nil {
					return err
				}
			}
			if _, err := tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE "+idCol+" = ?", s.id); err != nil {
				return err
			}
			removed++
		}
		return nil
	}

	if err := migrate(`SELECT car_id, car_name FROM cars`, carIDs, carNames,
		[]string{"owned_cars", "series_cars"}, "car_id", "cars"); err != nil {
		return err
	}
	if err := migrate(`SELECT track_id, CONCAT(track_name, ' ', COALESCE(config_name, '')) FROM tracks`,
		trackIDs, trackNames,
		[]string{"owned_tracks", "season_schedule", "track_requirements"}, "track_id", "tracks"); err != nil {
		return err
	}
	if err := migrate(`SELECT series_id, series_name FROM series`, seriesIDs, seriesNames,
		[]string{"favorite_series", "planned_races", "season_schedule", "series_cars"}, "series_id", "series"); err != nil {
		return err
	}

	// The fabricated rotation filled all 13 weeks for every series; the real
	// schedule leaves gaps. Drop schedule rows that aren't part of the real
	// week set for each series we have data for.
	for _, s := range data.Series {
		args := make([]any, 0, len(s.Weeks)+1)
		args = append(args, s.SeriesID)
		ph := make([]string, 0, len(s.Weeks))
		for _, w := range s.Weeks {
			args = append(args, w.Week)
			ph = append(ph, "?")
		}
		q := "DELETE FROM season_schedule WHERE series_id = ?"
		if len(ph) > 0 {
			q += " AND week NOT IN (" + strings.Join(ph, ",") + ")"
		}
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			return err
		}
	}

	if moved > 0 || removed > 0 {
		log.Printf("racing: catalog reconcile — %d stale rows migrated by name, %d removed", moved, removed)
	}
	return tx.Commit()
}

// normalizeName lowercases and strips everything but letters/digits, folding
// common accents, so "Nürburgring Grand-Prix" matches "nurburgring grand prix".
func normalizeName(s string) string {
	s = strings.ToLower(s)
	fold := strings.NewReplacer(
		"ü", "u", "ö", "o", "ä", "a", "é", "e", "è", "e", "á", "a", "ó", "o", "ñ", "n", "ç", "c", "í", "i",
	)
	s = fold.Replace(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
