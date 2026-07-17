// Command nav is the navigation service: a small standalone HTTP service that
// owns the application's menu configuration (the nav_items table) and serves it
// to the frontend, so the menu is configured on the backend rather than
// hard-coded in the SPA.
//
// It deliberately knows nothing about sessions or feature flags. It serves the
// menu STRUCTURE — including each item's requiresAuth / featureFlag metadata —
// and the frontend, which already holds the viewer and the flag map, decides
// what to actually render. Navigation is not a security boundary (every API
// route enforces its own auth, and hiding a link is cosmetic), so this keeps the
// service dependency-free: it needs MySQL and nothing else.
//
// It owns its schema: it creates and seeds nav_items on startup. The seed uses
// INSERT IGNORE, so defaults appear on a fresh DB and new items appear on
// upgrade, but edits made through the Cockpit are never clobbered.
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Item is one entry of the menu as served to the frontend.
type Item struct {
	Key      string `json:"key"`
	LabelKey string `json:"labelKey"` // i18n key, e.g. "nav.planner"
	Href     string `json:"href"`
	Icon     string `json:"icon"` // lucide name; mapped through a whitelist on the client
	// Placements is where the item may appear: "side" and/or "bottom".
	Placements []string `json:"placements"`
	Order      int      `json:"order"`
	// RequiresAuth / FeatureFlag are metadata the client filters on; empty
	// FeatureFlag means the item is not gated.
	RequiresAuth bool   `json:"requiresAuth"`
	FeatureFlag  string `json:"featureFlag,omitempty"`
}

// store reads the menu. An interface so the handler can be tested with a fake.
type store interface {
	List(ctx context.Context) ([]Item, error)
}

type dbStore struct{ db *sql.DB }

func (s *dbStore) List(ctx context.Context) ([]Item, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT item_key, label_key, href, icon, placements, sort_order,
		       requires_auth, feature_flag
		FROM nav_items
		WHERE enabled = 1
		ORDER BY sort_order, item_key`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Item, 0, 16)
	for rows.Next() {
		var (
			it         Item
			placements string
		)
		if err := rows.Scan(&it.Key, &it.LabelKey, &it.Href, &it.Icon, &placements,
			&it.Order, &it.RequiresAuth, &it.FeatureFlag); err != nil {
			return nil, err
		}
		it.Placements = splitPlacements(placements)
		items = append(items, it)
	}
	return items, rows.Err()
}

// splitPlacements parses the comma-separated placements column into a slice,
// trimming blanks so "side, bottom" and "side,bottom" behave the same.
func splitPlacements(s string) []string {
	out := make([]string, 0, 2)
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// navHandler serves GET /api/nav.
func navHandler(s store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := s.List(r.Context())
		if err != nil {
			log.Printf("nav: list: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "nav unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func healthHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := db.PingContext(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func dsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC&multiStatements=true",
		env("DB_USER", "app"), env("DB_PASSWORD", "app"),
		env("DB_HOST", "localhost"), env("DB_PORT", "3306"), env("DB_NAME", "app"))
}

// connect opens the pool and waits for MySQL to accept connections; the DB
// container may still be starting when this service does.
func connect(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn())
	if err != nil {
		return nil, err
	}
	for attempt := 1; attempt <= 30; attempt++ {
		if err = db.PingContext(ctx); err == nil {
			log.Printf("nav: db connected")
			return db, nil
		}
		log.Printf("nav: waiting for mysql (attempt %d/30): %v", attempt, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
	return nil, fmt.Errorf("db unreachable: %w", err)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := connect(ctx)
	if err != nil {
		log.Fatalf("nav: %v", err)
	}
	defer db.Close()

	if err := migrate(ctx, db); err != nil {
		log.Fatalf("nav: migrate: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/nav", navHandler(&dbStore{db: db}))
	mux.HandleFunc("GET /api/nav/health", healthHandler(db))

	srv := &http.Server{
		Addr:              ":" + env("PORT", "8081"),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("nav: listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("nav: serve: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Printf("nav: stopped")
}
