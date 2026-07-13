package cache

import (
	"context"
	"testing"
	"time"
)

// TestFailOpen verifies that a cache with no reachable Redis never errors and
// always behaves as a miss/no-op, so callers can rely on the DB fallback.
func TestFailOpen(t *testing.T) {
	cases := []struct {
		name string
		c    *Cache
	}{
		{name: "nil cache pointer", c: nil},
		{name: "disabled (blank addr)", c: New("")},
		{name: "downed redis (unreachable addr)", c: New("127.0.0.1:1")},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Use an already-canceled context so a real dial to the unreachable
			// addr returns immediately instead of waiting for the dial timeout.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// GetJSON must report a miss and leave dst untouched.
			var got map[string]bool
			if found := tc.c.GetJSON(ctx, "k", &got); found {
				t.Errorf("GetJSON found = true, want false (fail-open miss)")
			}
			if got != nil {
				t.Errorf("GetJSON populated dst on miss: %v", got)
			}

			// SetJSON / Delete / Ping must not panic and must swallow errors.
			tc.c.SetJSON(ctx, "k", map[string]bool{"a": true}, time.Second)
			tc.c.Delete(ctx, "k")
			if tc.c.Ping(ctx) {
				t.Errorf("Ping = true for unreachable/disabled cache")
			}
			if err := tc.c.Close(); err != nil {
				t.Errorf("Close() error on disabled cache: %v", err)
			}
		})
	}
}

// TestEnabled checks the Enabled predicate for the disabled constructions.
func TestEnabled(t *testing.T) {
	if New("").Enabled() {
		t.Error("blank-addr cache should be disabled")
	}
	var nilCache *Cache
	if nilCache.Enabled() {
		t.Error("nil cache should be disabled")
	}
	if !New("127.0.0.1:6379").Enabled() {
		t.Error("addr cache should report Enabled (regardless of reachability)")
	}
}
