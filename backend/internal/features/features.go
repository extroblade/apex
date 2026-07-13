// Package features is a small backend feature-toggle service. Flags live in the
// feature_flags table and are cached briefly so reads are cheap. New flags are
// seeded via migrations; code checks them with Enabled.
package features

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

const cacheTTL = 30 * time.Second

// Service reads feature flags with a short-lived cache.
type Service struct {
	db *sql.DB

	mu      sync.Mutex
	cache   map[string]bool
	refresh time.Time
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// All returns every flag and its state.
func (s *Service) All(ctx context.Context) (map[string]bool, error) {
	return s.load(ctx)
}

// Enabled reports whether a flag is on. Unknown or errored flags are off.
func (s *Service) Enabled(ctx context.Context, key string) bool {
	flags, err := s.load(ctx)
	if err != nil {
		return false
	}
	return flags[key]
}

func (s *Service) load(ctx context.Context) (map[string]bool, error) {
	s.mu.Lock()
	if s.cache != nil && time.Now().Before(s.refresh) {
		flags := s.cache
		s.mu.Unlock()
		return flags, nil
	}
	s.mu.Unlock()

	rows, err := s.db.QueryContext(ctx, `SELECT flag_key, enabled FROM feature_flags`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	flags := make(map[string]bool)
	for rows.Next() {
		var (
			key     string
			enabled bool
		)
		if err := rows.Scan(&key, &enabled); err != nil {
			return nil, err
		}
		flags[key] = enabled
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.cache = flags
	s.refresh = time.Now().Add(cacheTTL)
	s.mu.Unlock()
	return flags, nil
}
