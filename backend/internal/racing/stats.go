package racing

import (
	"context"

	"apex/internal/iracing"
)

// Dashboard is the combined live snapshot shown on the driver dashboard.
type Dashboard struct {
	Account  Account              `json:"account"`
	Licenses []iracing.License    `json:"licenses"`
	Career   []iracing.CareerStat `json:"career"`
	Recent   []iracing.RecentRace `json:"recent"`
}

// Stats fetches the user's current licenses, career stats, and recent races
// live from iRacing (via the cached session).
func (s *Service) Stats(ctx context.Context, userID int64) (Dashboard, error) {
	client, custID, err := s.clientFor(ctx, userID)
	if err != nil {
		return Dashboard{}, err
	}

	member, err := client.Member(ctx, custID)
	if err != nil {
		return Dashboard{}, err
	}
	career, err := client.Career(ctx, custID)
	if err != nil {
		return Dashboard{}, err
	}
	recent, err := client.RecentRaces(ctx, custID)
	if err != nil {
		return Dashboard{}, err
	}

	account, err := s.Status(ctx, userID)
	if err != nil {
		return Dashboard{}, err
	}

	return Dashboard{
		Account:  account,
		Licenses: member.Licenses,
		Career:   career,
		Recent:   recent,
	}, nil
}
