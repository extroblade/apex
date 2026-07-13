package racing

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"apex/internal/iracing"
)

// cacheTTL is how long a cached driver payload stays fresh. Lookups use the
// requesting user's own iRacing session, and results are cached globally by
// cust_id (the data is the same whoever fetches it), which also means a driver
// looked up by one user is cheap for the next.
const cacheTTL = 6 * time.Hour

// DriverProfile is a driver's stats, assembled from cached/live iRacing data.
type DriverProfile struct {
	CustID      int                  `json:"custId"`
	DisplayName string               `json:"displayName"`
	Licenses    []iracing.License    `json:"licenses"`
	Career      []iracing.CareerStat `json:"career"`
	Recent      []iracing.RecentRace `json:"recent"`
	CachedAt    time.Time            `json:"cachedAt"`
}

// SearchDrivers finds drivers by name, using the user's own iRacing session.
func (s *Service) SearchDrivers(ctx context.Context, userID int64, term string) ([]iracing.DriverSearchResult, error) {
	client, _, err := s.clientFor(ctx, userID)
	if err != nil {
		return nil, err
	}
	return client.SearchDrivers(ctx, term)
}

// LookupDriver returns any driver's stats, served from cache when fresh and
// otherwise fetched through the requesting user's own session.
func (s *Service) LookupDriver(ctx context.Context, userID int64, custID int) (DriverProfile, error) {
	// Obtain the user's client lazily — only if something actually needs fetching.
	var client APIClient
	getClient := func() (APIClient, error) {
		if client != nil {
			return client, nil
		}
		c, _, err := s.clientFor(ctx, userID)
		if err != nil {
			return nil, err
		}
		client = c
		return c, nil
	}

	member, memberAt, err := loadCached(ctx, s, custID, "member",
		func() (iracing.Member, error) {
			c, err := getClient()
			if err != nil {
				return iracing.Member{}, err
			}
			return c.Member(ctx, custID)
		})
	if err != nil {
		return DriverProfile{}, err
	}

	career, careerAt, err := loadCached(ctx, s, custID, "career",
		func() ([]iracing.CareerStat, error) {
			c, err := getClient()
			if err != nil {
				return nil, err
			}
			return c.Career(ctx, custID)
		})
	if err != nil {
		return DriverProfile{}, err
	}

	recent, recentAt, err := loadCached(ctx, s, custID, "recent",
		func() ([]iracing.RecentRace, error) {
			c, err := getClient()
			if err != nil {
				return nil, err
			}
			return c.RecentRaces(ctx, custID)
		})
	if err != nil {
		return DriverProfile{}, err
	}

	return DriverProfile{
		CustID:      custID,
		DisplayName: member.DisplayName,
		Licenses:    member.Licenses,
		Career:      career,
		Recent:      recent,
		CachedAt:    oldest(memberAt, careerAt, recentAt),
	}, nil
}

// loadCached returns a value of type T for (custID, kind): from cache if fresh,
// otherwise by calling fetch and writing the result back. On a fetch error it
// falls back to stale cache if any exists. Package-level generic because Go
// methods can't have their own type parameters.
func loadCached[T any](
	ctx context.Context,
	s *Service,
	custID int,
	kind string,
	fetch func() (T, error),
) (T, time.Time, error) {
	var zero T

	payload, fetchedAt, err := s.readCache(ctx, custID, kind)
	if err != nil {
		return zero, time.Time{}, err
	}

	if payload != nil && time.Since(fetchedAt) < cacheTTL {
		var v T
		if json.Unmarshal(payload, &v) == nil {
			return v, fetchedAt, nil
		}
		// Corrupt cache entry: fall through and re-fetch.
	}

	v, ferr := fetch()
	if ferr != nil {
		// Serve stale data rather than failing, if we have any.
		if payload != nil {
			var sv T
			if json.Unmarshal(payload, &sv) == nil {
				return sv, fetchedAt, nil
			}
		}
		return zero, time.Time{}, ferr
	}

	if b, mErr := json.Marshal(v); mErr == nil {
		_ = s.writeCache(ctx, custID, kind, b) // best-effort
	}
	return v, time.Now().UTC(), nil
}

func (s *Service) readCache(ctx context.Context, custID int, kind string) ([]byte, time.Time, error) {
	var (
		payload   []byte
		fetchedAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT payload_json, fetched_at FROM driver_cache WHERE cust_id = ? AND kind = ?`,
		custID, kind).Scan(&payload, &fetchedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, err
	}
	return payload, fetchedAt, nil
}

func (s *Service) writeCache(ctx context.Context, custID int, kind string, payload []byte) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO driver_cache (cust_id, kind, payload_json, fetched_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE payload_json = VALUES(payload_json), fetched_at = VALUES(fetched_at)`,
		custID, kind, payload, time.Now().UTC())
	return err
}

func oldest(times ...time.Time) time.Time {
	var out time.Time
	for _, t := range times {
		if out.IsZero() || (!t.IsZero() && t.Before(out)) {
			out = t
		}
	}
	return out
}
