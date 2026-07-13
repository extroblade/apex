package iracing

import "context"

// Car is a catalog car entry (from /data/car/get).
type Car struct {
	CarID   int    `json:"car_id"`
	CarName string `json:"car_name"`
}

// CatalogTrack is a catalog track configuration (from /data/track/get).
type CatalogTrack struct {
	TrackID    int    `json:"track_id"`
	TrackName  string `json:"track_name"`
	ConfigName string `json:"config_name"`
	Category   string `json:"category"`
}

// CatalogSeries is a catalog series entry (from /data/series/get).
type CatalogSeries struct {
	SeriesID   int    `json:"series_id"`
	SeriesName string `json:"series_name"`
	Category   string `json:"category"`
	CategoryID int    `json:"category_id"`
}

// Cars returns the full car catalog. These lookup endpoints return a bare JSON
// array (via the S3 link).
func (c *Client) Cars(ctx context.Context) ([]Car, error) {
	var out []Car
	if err := c.get(ctx, "/data/car/get", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Tracks returns the full track catalog (one entry per track configuration).
func (c *Client) Tracks(ctx context.Context) ([]CatalogTrack, error) {
	var out []CatalogTrack
	if err := c.get(ctx, "/data/track/get", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// Series returns the full series catalog.
func (c *Client) Series(ctx context.Context) ([]CatalogSeries, error) {
	var out []CatalogSeries
	if err := c.get(ctx, "/data/series/get", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
