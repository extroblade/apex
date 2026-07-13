package racing

import (
	"context"
	"database/sql"
)

// GroupStat is one row of a comparator: aggregate performance for a dimension
// value (a category, car, or track) over the user's synced races.
//
// NOTE: "best" here means best *finishing/rating* performance, computed from
// synced race results — not lap-time pace. True fastest-by-lap-time would need
// per-lap telemetry ingestion, a later extension.
type GroupStat struct {
	Key            int     `json:"key"`   // category_id / car_id / track_id
	Label          string  `json:"label"` // human name where available
	Races          int     `json:"races"`
	AvgFinish      float64 `json:"avgFinish"`
	AvgStart       float64 `json:"avgStart"`
	AvgIncidents   float64 `json:"avgIncidents"`
	AvgIRatingGain float64 `json:"avgIRatingGain"`
}

// CompareCategories aggregates the user's races by racing category.
func (s *Service) CompareCategories(ctx context.Context, userID int64) ([]GroupStat, error) {
	return s.groupBy(ctx, userID,
		`SELECT category_id, CAST(category_id AS CHAR)`, "category_id")
}

// CompareCars aggregates the user's races by car.
func (s *Service) CompareCars(ctx context.Context, userID int64) ([]GroupStat, error) {
	return s.groupBy(ctx, userID,
		`SELECT car_id, CAST(car_id AS CHAR)`, "car_id")
}

// CompareTracks aggregates the user's races by track.
func (s *Service) CompareTracks(ctx context.Context, userID int64) ([]GroupStat, error) {
	return s.groupBy(ctx, userID,
		`SELECT track_id, MAX(track_name)`, "track_id")
}

// groupBy runs a shared aggregation. keyExpr provides the "SELECT <key>, <label>"
// prefix and groupCol is the column to GROUP BY / order within. Both come only
// from the trusted callers above (never user input), so string-building the
// query here is safe.
func (s *Service) groupBy(ctx context.Context, userID int64, keyExpr, groupCol string) ([]GroupStat, error) {
	query := keyExpr + ` AS label_out,
		COUNT(*) AS races,
		AVG(finish_position),
		AVG(start_position),
		AVG(incidents),
		AVG(new_irating - old_irating)
		FROM races
		WHERE user_id = ?
		GROUP BY ` + groupCol + `
		ORDER BY AVG(finish_position) ASC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]GroupStat, 0)
	for rows.Next() {
		var (
			g     GroupStat
			label sql.NullString
		)
		if err := rows.Scan(&g.Key, &label, &g.Races,
			&g.AvgFinish, &g.AvgStart, &g.AvgIncidents, &g.AvgIRatingGain); err != nil {
			return nil, err
		}
		g.Label = label.String
		results = append(results, g)
	}
	return results, rows.Err()
}
