package racing

import "context"

// Sync pulls the user's recent races from iRacing and upserts them into the
// races table. Returns how many races were processed. Idempotent: re-syncing
// updates existing rows rather than duplicating them (PK is user+subsession).
func (s *Service) Sync(ctx context.Context, userID int64) (int, error) {
	client, custID, err := s.clientFor(ctx, userID)
	if err != nil {
		return 0, err
	}

	races, err := client.RecentRaces(ctx, custID)
	if err != nil {
		return 0, err
	}

	const q = `
		INSERT INTO races
			(user_id, subsession_id, series_id, series_name, category_id, car_id,
			 track_id, track_name, start_position, finish_position, incidents,
			 old_irating, new_irating, laps_complete, session_start_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			series_name = VALUES(series_name),
			finish_position = VALUES(finish_position),
			incidents = VALUES(incidents),
			new_irating = VALUES(new_irating),
			laps_complete = VALUES(laps_complete)`

	// A transaction keeps the batch atomic and is much faster than N autocommits.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after a successful Commit

	for _, r := range races {
		var startTime any
		if !r.SessionStartTime.IsZero() {
			startTime = r.SessionStartTime.UTC()
		}
		if _, err := tx.ExecContext(ctx, q,
			userID, r.SubsessionID, r.SeriesID, r.SeriesName, r.CategoryID, r.CarID,
			r.Track.TrackID, r.Track.TrackName, r.StartPosition, r.FinishPosition,
			r.Incidents, r.OldiRating, r.NewiRating, r.LapsComplete, startTime,
		); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return len(races), nil
}
