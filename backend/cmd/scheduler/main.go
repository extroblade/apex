// The scheduler is a small companion service that keeps catalog data fresh:
// once a day (and at startup) it looks for a new iRacing season-schedule PDF —
// probing known URLs and scanning a drop directory — and parses any new file
// into the season_schedule table. The API service owns migrations; this
// process waits for the schema instead of migrating itself.
package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"apex/internal/config"
	"apex/internal/contentsync"
	"apex/internal/db"
	"apex/internal/schedulepdf"
)

func main() {
	cfg := config.Load()

	database, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()

	waitForSchema(database)

	ingestor := &schedulepdf.Ingestor{
		DB:   database,
		HTTP: &http.Client{Timeout: 60 * time.Second},
		URLs:     splitList(os.Getenv("SCHEDULE_PDF_URLS")),
		Dir:      os.Getenv("SCHEDULE_PDF_DIR"),
		ImageDir: os.Getenv("SCHEDULE_IMAGE_DIR"),
	}

	// The season schedule can appear any day once a new season opens, so we check
	// for a new PDF daily.
	interval := durationEnv("SCHEDULE_CHECK_INTERVAL", 24*time.Hour, time.Minute)
	// The full car/track catalog barely moves (a handful of new cars/tracks per
	// season), so re-fetching the internet list daily is wasteful. Sync it weekly
	// — plus once at startup — and rely on the deduped content-hash guard to skip
	// unchanged files anyway.
	contentInterval := durationEnv("CONTENT_SYNC_INTERVAL", 7*24*time.Hour, time.Hour)

	syncer := &contentsync.Syncer{
		DB:   database,
		HTTP: &http.Client{Timeout: 60 * time.Second},
	}

	log.Printf("scheduler: season PDFs every %s, content list every %s (dir=%q, %d extra urls)",
		interval, contentInterval, ingestor.Dir, len(ingestor.URLs))

	// Content sync runs on its own slow ticker so a fresh schedule check never
	// waits on (or drags in) a full catalog re-fetch.
	go func() {
		for {
			syncer.Run(context.Background())
			time.Sleep(contentInterval)
		}
	}()

	for {
		ingestor.Run(context.Background())
		time.Sleep(interval)
	}
}

// durationEnv parses a duration from env, falling back to def and clamping to a
// sensible minimum so a typo can't turn the loop into a busy-wait.
func durationEnv(key string, def, min time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= min {
			return d
		}
	}
	return def
}

// waitForSchema blocks until the API's migrations have created our tables
// (the API service owns migrations; racing it would double-apply DDL).
func waitForSchema(database *sql.DB) {
	for attempt := 1; ; attempt++ {
		var one int
		err := database.QueryRow(`SELECT 1 FROM schedule_imports LIMIT 1`).Scan(&one)
		if err == nil || errors.Is(err, sql.ErrNoRows) {
			return
		}
		if attempt >= 60 {
			log.Fatalf("schema not ready after %d attempts: %v", attempt, err)
		}
		log.Printf("waiting for schema (attempt %d): %v", attempt, err)
		time.Sleep(5 * time.Second)
	}
}

func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
