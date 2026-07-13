// Package features is a small backend feature-toggle service. Flags live in the
// feature_flags table and are cached briefly so reads are cheap. New flags are
// seeded via migrations; code checks them with Enabled.
package features

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"time"

	"apex/internal/cache"
)

const (
	cacheTTL = 30 * time.Second
	// redisKey holds the JSON-encoded flag map shared across processes.
	redisKey = "features:flags"
)

// ErrFlagNotFound is returned by Set when the key does not exist.
var ErrFlagNotFound = errors.New("flag not found")

// Service reads feature flags with a short-lived cache. An optional Redis cache
// (fail-open) is consulted between the in-process cache and the DB so flag reads
// stay cheap and consistent across processes.
type Service struct {
	db    *sql.DB
	redis *cache.Cache

	mu      sync.Mutex
	cache   map[string]bool
	refresh time.Time
}

func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// WithCache attaches a Redis cache used as an L2 for flag reads. A nil cache
// (or one whose Redis is down) simply degrades to DB reads.
func (s *Service) WithCache(c *cache.Cache) *Service {
	s.redis = c
	return s
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

// Set flips a single feature flag in the DB. Returns an error if the key does
// not exist. Callers should follow up with Invalidate to flush the cache.
func (s *Service) Set(ctx context.Context, key string, enabled bool) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE feature_flags SET enabled = ? WHERE flag_key = ?`, enabled, key)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrFlagNotFound
	}
	return nil
}

// Invalidate clears both the in-process and Redis caches so the next read hits
// the DB. Call it after any write (e.g. the Cockpit toggle).
func (s *Service) Invalidate() {
	s.mu.Lock()
	s.cache = nil
	s.mu.Unlock()
	if s.redis != nil {
		s.redis.Delete(context.Background(), redisKey)
	}
}

func (s *Service) load(ctx context.Context) (map[string]bool, error) {
	s.mu.Lock()
	if s.cache != nil && time.Now().Before(s.refresh) {
		flags := s.cache
		s.mu.Unlock()
		return flags, nil
	}
	s.mu.Unlock()

	// L2: Redis (shared across processes). Fail-open — a miss or downed Redis
	// falls through to the DB.
	if s.redis != nil {
		var flags map[string]bool
		if s.redis.GetJSON(ctx, redisKey, &flags) {
			s.store(flags)
			return flags, nil
		}
	}

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

	s.store(flags)
	if s.redis != nil {
		s.redis.SetJSON(ctx, redisKey, flags, cacheTTL)
	}
	return flags, nil
}

// store puts flags in the in-process cache with a fresh TTL.
func (s *Service) store(flags map[string]bool) {
	s.mu.Lock()
	s.cache = flags
	s.refresh = time.Now().Add(cacheTTL)
	s.mu.Unlock()
}
