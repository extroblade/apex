// Package locales serves the app's i18n bundles from the DB so the set of
// languages is backend-driven: a new language is a row in the `locales` table
// (or a future Cockpit editor) and the frontend picks it up with no rebuild.
//
// `en` is ALSO bundled in the frontend as the instant, offline fallback and the
// source of the Translation type, so the frontend never fetches en — but it is
// seeded here too for completeness and so `/api/locales` lists it.
package locales

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

// seedFS holds the built-in bundles, generated from the frontend source
// (npm run gen:locales writes them here) and embedded at build time.
//
//go:embed data/*.json
var seedFS embed.FS

// ErrNotFound is returned by Bundle for an unknown or disabled locale.
var ErrNotFound = errors.New("locale not found")

// Info is one entry of the language list (GET /api/locales).
type Info struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// builtins name and order the bundled languages; their bundles come from the
// embedded JSON keyed by code. Runtime-added languages use other codes and are
// never touched by the seed.
var builtins = []struct {
	code, name string
	order      int
}{
	{"en", "English", 10},
	{"ru", "Русский", 20},
}

// Seed upserts the built-in locales from the embedded bundles. It refreshes them
// on every startup (so a translation edit ships on deploy) via ON DUPLICATE KEY
// UPDATE, but only for the built-in codes — locales added at runtime survive.
func Seed(ctx context.Context, db *sql.DB) error {
	for _, b := range builtins {
		raw, err := seedFS.ReadFile("data/" + b.code + ".json")
		if err != nil {
			return fmt.Errorf("read embedded bundle %s: %w", b.code, err)
		}
		// Compact away the pretty-print whitespace before storing.
		var buf bytes.Buffer
		if err := json.Compact(&buf, raw); err != nil {
			return fmt.Errorf("bundle %s is not valid JSON: %w", b.code, err)
		}
		if _, err := db.ExecContext(ctx, `
			INSERT INTO locales (code, name, bundle, sort_order)
			VALUES (?, ?, ?, ?)
			ON DUPLICATE KEY UPDATE name = VALUES(name), bundle = VALUES(bundle),
				sort_order = VALUES(sort_order)`,
			b.code, b.name, buf.String(), b.order); err != nil {
			return err
		}
	}
	log.Printf("locales: seeded %d built-in locale(s)", len(builtins))
	return nil
}

// Service reads locales for the HTTP handlers.
type Service struct{ db *sql.DB }

func New(db *sql.DB) *Service { return &Service{db: db} }

// List returns the enabled languages (code + display name), ordered.
func (s *Service) List(ctx context.Context) ([]Info, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT code, name FROM locales WHERE enabled = 1 ORDER BY sort_order, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]Info, 0, 8)
	for rows.Next() {
		var it Info
		if err := rows.Scan(&it.Code, &it.Name); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// Bundle returns the raw JSON translation bundle for a locale, or ErrNotFound.
// The value is stored (and returned) as a JSON document, ready to write straight
// to the response.
func (s *Service) Bundle(ctx context.Context, code string) (string, error) {
	var bundle string
	err := s.db.QueryRowContext(ctx,
		`SELECT bundle FROM locales WHERE code = ? AND enabled = 1`, code).Scan(&bundle)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return bundle, nil
}
