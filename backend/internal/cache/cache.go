// Package cache is a fail-open wrapper around Redis. Every method degrades to a
// miss/no-op when no client is configured or Redis is unreachable, so callers
// can always fall back to their source of truth (the DB) without ever handling a
// cache error. A nil *Cache and a Cache with a nil client both behave as "cache
// disabled — always a miss".
package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// opTimeout caps every cache operation so a downed/slow Redis fails open fast
// instead of blocking the request while go-redis retries the dial. Without it a
// single op against an unreachable server can hang for over a minute.
const opTimeout = 300 * time.Millisecond

// Cache holds an optional Redis client. Construct with New.
type Cache struct {
	rdb *redis.Client
}

// New returns a Cache backed by the Redis server at addr. A blank addr yields a
// disabled cache (nil client) that always misses — this is the fail-open default
// when REDIS_ADDR is unset. The client is tuned to fail fast (short timeouts, no
// retries) so an unreachable Redis never stalls a request.
func New(addr string) *Cache {
	if addr == "" {
		return &Cache{}
	}
	return &Cache{rdb: redis.NewClient(&redis.Options{
		Addr:         addr,
		DialTimeout:  opTimeout,
		ReadTimeout:  opTimeout,
		WriteTimeout: opTimeout,
		PoolTimeout:  opTimeout,
		MaxRetries:   -1, // -1 disables retries; a miss falls through to the DB
	})}
}

// op derives a short-deadline context so a single Redis call can't block longer
// than opTimeout regardless of the caller's context.
func op(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, opTimeout)
}

// Enabled reports whether a Redis client is configured. It says nothing about
// whether Redis is currently reachable — reads/writes fail open regardless.
func (c *Cache) Enabled() bool { return c != nil && c.rdb != nil }

// GetJSON unmarshals the value stored at key into dst. It returns found=false on
// any miss, disabled cache, unmarshal error, or Redis error — none of which are
// surfaced, so the caller simply proceeds to its source of truth.
func (c *Cache) GetJSON(ctx context.Context, key string, dst any) (found bool) {
	if !c.Enabled() {
		return false
	}
	ctx, cancel := op(ctx)
	defer cancel()
	data, err := c.rdb.Get(ctx, key).Bytes()
	if err != nil {
		// redis.Nil (plain miss) or a connection error — both are just a miss.
		return false
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return false
	}
	return true
}

// SetJSON stores v (JSON-encoded) at key with the given TTL. Caching is
// best-effort: a disabled cache, a marshal error, or a downed Redis is silently
// ignored.
func (c *Cache) SetJSON(ctx context.Context, key string, v any, ttl time.Duration) {
	if !c.Enabled() {
		return
	}
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	ctx, cancel := op(ctx)
	defer cancel()
	_ = c.rdb.Set(ctx, key, data, ttl).Err()
}

// Delete removes keys. Errors (including a downed Redis) are swallowed.
func (c *Cache) Delete(ctx context.Context, keys ...string) {
	if !c.Enabled() || len(keys) == 0 {
		return
	}
	ctx, cancel := op(ctx)
	defer cancel()
	_ = c.rdb.Del(ctx, keys...).Err()
}

// Ping reports whether Redis currently answers. A disabled cache or any error
// reports false without side effects; use it only for diagnostics (health).
func (c *Cache) Ping(ctx context.Context) bool {
	if !c.Enabled() {
		return false
	}
	ctx, cancel := op(ctx)
	defer cancel()
	return c.rdb.Ping(ctx).Err() == nil
}

// Close releases the underlying client. Safe to call on a disabled cache.
func (c *Cache) Close() error {
	if !c.Enabled() {
		return nil
	}
	return c.rdb.Close()
}
