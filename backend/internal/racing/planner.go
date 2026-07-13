package racing

import "context"

// --- DTOs returned to the client (catalog items carry per-user flags) ---

type CarItem struct {
	CarID       int    `json:"carId"`
	CarName     string `json:"carName"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Free        bool   `json:"free"`
	ImagePath   string `json:"imagePath"`
	Owned       bool   `json:"owned"`
}

type TrackItem struct {
	TrackID     int    `json:"trackId"`
	TrackName   string `json:"trackName"`
	ConfigName  string `json:"configName"`
	Category    string `json:"category"`
	Description string `json:"description"`
	Free        bool   `json:"free"`
	ImagePath   string `json:"imagePath"`
	Owned       bool   `json:"owned"`
}

type SeriesItem struct {
	SeriesID      int    `json:"seriesId"`
	SeriesName    string `json:"seriesName"`
	Category      string `json:"category"`
	Description   string `json:"description"`
	LicenseNeeded string `json:"licenseNeeded"`
	ImagePath     string `json:"imagePath"`
	Favorite      bool   `json:"favorite"`
}

// CatalogCounts reports how many items were synced.
type CatalogCounts struct {
	Cars   int `json:"cars"`
	Tracks int `json:"tracks"`
	Series int `json:"series"`
}

// --- Catalog sync (needs a linked iRacing session) ---

// SyncCatalog fetches the car/track/series catalog from iRacing and upserts it.
// The catalog is global; any linked user can refresh it.
func (s *Service) SyncCatalog(ctx context.Context, userID int64) (CatalogCounts, error) {
	client, _, err := s.clientFor(ctx, userID)
	if err != nil {
		return CatalogCounts{}, err
	}

	cars, err := client.Cars(ctx)
	if err != nil {
		return CatalogCounts{}, err
	}
	tracks, err := client.Tracks(ctx)
	if err != nil {
		return CatalogCounts{}, err
	}
	series, err := client.Series(ctx)
	if err != nil {
		return CatalogCounts{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CatalogCounts{}, err
	}
	defer tx.Rollback() //nolint:errcheck // no-op after Commit

	for _, c := range cars {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO cars (car_id, car_name) VALUES (?, ?)
			 ON DUPLICATE KEY UPDATE car_name = VALUES(car_name)`,
			c.CarID, c.CarName); err != nil {
			return CatalogCounts{}, err
		}
	}
	for _, t := range tracks {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO tracks (track_id, track_name, config_name, category) VALUES (?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE track_name = VALUES(track_name),
				config_name = VALUES(config_name), category = VALUES(category)`,
			t.TrackID, t.TrackName, t.ConfigName, t.Category); err != nil {
			return CatalogCounts{}, err
		}
	}
	for _, sr := range series {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO series (series_id, series_name, category, category_id) VALUES (?, ?, ?, ?)
			 ON DUPLICATE KEY UPDATE series_name = VALUES(series_name),
				category = VALUES(category), category_id = VALUES(category_id)`,
			sr.SeriesID, sr.SeriesName, sr.Category, sr.CategoryID); err != nil {
			return CatalogCounts{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return CatalogCounts{}, err
	}
	return CatalogCounts{Cars: len(cars), Tracks: len(tracks), Series: len(series)}, nil
}

// --- Catalog reads with per-user flags ---

func (s *Service) Cars(ctx context.Context, userID int64) ([]CarItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.car_id, c.car_name, c.category, c.description, c.is_free, c.image_path,
		       o.car_id IS NOT NULL AS owned
		FROM cars c
		LEFT JOIN owned_cars o ON o.car_id = c.car_id AND o.user_id = ?
		ORDER BY c.car_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]CarItem, 0)
	for rows.Next() {
		var it CarItem
		if err := rows.Scan(&it.CarID, &it.CarName, &it.Category, &it.Description, &it.Free, &it.ImagePath, &it.Owned); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *Service) Tracks(ctx context.Context, userID int64) ([]TrackItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.track_id, t.track_name, t.config_name, t.category, t.description, t.is_free,
		       t.image_path, o.track_id IS NOT NULL AS owned
		FROM tracks t
		LEFT JOIN owned_tracks o ON o.track_id = t.track_id AND o.user_id = ?
		ORDER BY t.track_name, t.config_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]TrackItem, 0)
	for rows.Next() {
		var it TrackItem
		if err := rows.Scan(&it.TrackID, &it.TrackName, &it.ConfigName, &it.Category, &it.Description, &it.Free, &it.ImagePath, &it.Owned); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *Service) SeriesList(ctx context.Context, userID int64) ([]SeriesItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.series_id, s.series_name, s.category, s.description, s.license_needed,
		       s.image_path, f.series_id IS NOT NULL AS favorite
		FROM series s
		LEFT JOIN favorite_series f ON f.series_id = s.series_id AND f.user_id = ?
		ORDER BY s.series_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]SeriesItem, 0)
	for rows.Next() {
		var it SeriesItem
		if err := rows.Scan(&it.SeriesID, &it.SeriesName, &it.Category, &it.Description, &it.LicenseNeeded, &it.ImagePath, &it.Favorite); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// --- Ownership / favorite toggles ---

func (s *Service) SetCarOwned(ctx context.Context, userID int64, carID int, owned bool) error {
	return s.toggle(ctx, "owned_cars", "car_id", userID, carID, owned)
}

func (s *Service) SetTrackOwned(ctx context.Context, userID int64, trackID int, owned bool) error {
	return s.toggle(ctx, "owned_tracks", "track_id", userID, trackID, owned)
}

func (s *Service) SetSeriesFavorite(ctx context.Context, userID int64, seriesID int, favorite bool) error {
	return s.toggle(ctx, "favorite_series", "series_id", userID, seriesID, favorite)
}

// toggle inserts or deletes a (user_id, itemID) row. table/col come only from
// the trusted callers above, never from user input.
func (s *Service) toggle(ctx context.Context, table, col string, userID int64, itemID int, on bool) error {
	if on {
		_, err := s.db.ExecContext(ctx,
			`INSERT IGNORE INTO `+table+` (user_id, `+col+`) VALUES (?, ?)`, userID, itemID)
		return err
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM `+table+` WHERE user_id = ? AND `+col+` = ?`, userID, itemID)
	return err
}
