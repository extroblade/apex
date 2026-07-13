// Package contentsync keeps the car/track catalog complete by fetching the
// community-maintained full iRacing content lists (my-racing-planner's data
// exports of the official API) and upserting them: real ids, every legacy car
// and track layout, free flags, prices, and sku purchase groups. The catalog
// barely changes, so the scheduler runs this on a slow (weekly) ticker rather
// than the daily schedule-PDF check; the content-hash guard skips unchanged
// files regardless.
package contentsync

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	defaultCarsURL   = "https://raw.githubusercontent.com/adrianulima/my-racing-planner/main/src/ir-data/cars.json"
	defaultTracksURL = "https://raw.githubusercontent.com/adrianulima/my-racing-planner/main/src/ir-data/tracks.json"
)

type car struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Categories []string `json:"categories"`
	Free       bool     `json:"free"`
	Price      float64  `json:"price"`
}

type track struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Config     string   `json:"config"`
	Categories []string `json:"categories"`
	Free       bool     `json:"free"`
	Price      float64  `json:"price"`
	SKU        int      `json:"sku"`
}

// Syncer fetches and applies the content lists.
type Syncer struct {
	DB   *sql.DB
	HTTP *http.Client
}

func urlOr(env, fallback string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return fallback
}

// Run refreshes the catalog from two complementary sources (deduped by content
// hash via schedule_imports, so unchanged fetches are skipped):
//   - the community JSON lists give the full set of ids incl. legacy content,
//     with prices and sku groups;
//   - the iRacing web catalog adds artwork (image_path) and the authoritative
//     free/included flag, matched onto our rows by name.
func (s *Syncer) Run(ctx context.Context) {
	if err := s.syncCars(ctx, urlOr("CONTENT_CARS_URL", defaultCarsURL)); err != nil {
		log.Printf("contentsync: cars: %v", err)
	}
	if err := s.syncTracks(ctx, urlOr("CONTENT_TRACKS_URL", defaultTracksURL)); err != nil {
		log.Printf("contentsync: tracks: %v", err)
	}
	// Series AFTER cars/tracks so schedule rows reference known catalog ids.
	if err := s.syncSeries(ctx, urlOr("CONTENT_SERIES_URL", defaultSeriesURL)); err != nil {
		log.Printf("contentsync: series: %v", err)
	}
	if err := s.syncWebCatalog(ctx, urlOr("CONTENT_CARS_HTML_URL", defaultCarsHTMLURL), "cars", "car_name"); err != nil {
		log.Printf("contentsync: web cars: %v", err)
	}
	if err := s.syncWebCatalog(ctx, urlOr("CONTENT_TRACKS_HTML_URL", defaultTracksHTMLURL), "tracks", "track_name"); err != nil {
		log.Printf("contentsync: web tracks: %v", err)
	}
}

func (s *Syncer) fetch(ctx context.Context, url string) (data []byte, hash string, fresh bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", false, err
	}
	req.Header.Set("User-Agent", "Apex/1.0 (content sync)")
	resp, err := s.HTTP.Do(req)
	if err != nil {
		return nil, "", false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("status %d", resp.StatusCode)
	}
	data, err = io.ReadAll(io.LimitReader(resp.Body, 16<<20))
	if err != nil {
		return nil, "", false, err
	}

	sum := sha256.Sum256(data)
	hash = hex.EncodeToString(sum[:])
	var seen int
	if err := s.DB.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM schedule_imports WHERE sha256 = ?`, hash).Scan(&seen); err != nil {
		return nil, "", false, err
	}
	return data, hash, seen == 0, nil
}

func (s *Syncer) record(ctx context.Context, hash, source string, found, updated int) error {
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO schedule_imports (sha256, source, series_found, series_matched, weeks_updated)
		VALUES (?, ?, ?, ?, 0)`, hash, source, found, updated)
	return err
}

func (s *Syncer) syncCars(ctx context.Context, url string) error {
	data, hash, fresh, err := s.fetch(ctx, url)
	if err != nil || !fresh {
		return err
	}
	var cars map[string]car
	if err := json.Unmarshal(data, &cars); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	updated := 0
	for _, c := range cars {
		if c.ID == 0 || c.Name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO cars (car_id, car_name, category, is_free, price)
			VALUES (?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE car_name = VALUES(car_name), category = VALUES(category),
				is_free = VALUES(is_free), price = VALUES(price)`,
			c.ID, c.Name, first(c.Categories), c.Free, c.Price); err != nil {
			return err
		}
		updated++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("contentsync: %d cars upserted from %s", updated, url)
	return s.record(ctx, hash, "cars:"+url, len(cars), updated)
}

func (s *Syncer) syncTracks(ctx context.Context, url string) error {
	data, hash, fresh, err := s.fetch(ctx, url)
	if err != nil || !fresh {
		return err
	}
	var tracks map[string]track
	if err := json.Unmarshal(data, &tracks); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	updated := 0
	for _, t := range tracks {
		if t.ID == 0 || t.Name == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO tracks (track_id, track_name, config_name, category, is_free, price, sku_group)
			VALUES (?, ?, ?, ?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE track_name = VALUES(track_name),
				config_name = VALUES(config_name), category = VALUES(category),
				is_free = VALUES(is_free), price = VALUES(price), sku_group = VALUES(sku_group)`,
			t.ID, t.Name, t.Config, first(t.Categories), t.Free, t.Price, t.SKU); err != nil {
			return err
		}
		updated++
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("contentsync: %d track configs upserted from %s", updated, url)
	return s.record(ctx, hash, "tracks:"+url, len(tracks), updated)
}

func first(list []string) string {
	if len(list) == 0 {
		return ""
	}
	return list[0]
}
