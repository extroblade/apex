// Package setups is the "setups showroom": users save car setups privately and
// optionally publish them for everyone to browse and download. Setups are plain
// text (an exported .sto's contents or a values dump), so no binary handling —
// the store is a single table joined to the catalog for display names.
package setups

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

// Errors surfaced to the handler for status mapping.
var (
	ErrNotFound  = errors.New("setup not found")
	ErrForbidden = errors.New("not your setup")
	ErrInvalid   = errors.New("invalid setup")
)

// Service is the setups store.
type Service struct {
	db *sql.DB
}

// New returns a setups service backed by db.
func New(db *sql.DB) *Service { return &Service{db: db} }

// Setup is a showroom row with catalog display names and per-user flags.
type Setup struct {
	ID        int    `json:"id"`
	CarID     int    `json:"carId"`
	CarName   string `json:"carName"`
	TrackID   int    `json:"trackId"`
	TrackName string `json:"trackName"`
	Name      string `json:"name"`
	Notes     string `json:"notes"`
	Data      string `json:"data,omitempty"` // omitted from list responses
	Category  string `json:"category"`
	Author    string `json:"author"`
	Public    bool   `json:"public"`
	Downloads int    `json:"downloads"`
	Mine      bool   `json:"mine"`
	CreatedAt string `json:"createdAt"`
}

// Input is the create payload.
type Input struct {
	CarID   int
	TrackID int
	Name    string
	Notes   string
	Data    string
	Public  bool
}

// Create validates and inserts a setup owned by userID.
func (s *Service) Create(ctx context.Context, userID int64, in Input) (Setup, error) {
	in.Name = strings.TrimSpace(in.Name)
	if in.Name == "" || in.CarID <= 0 || strings.TrimSpace(in.Data) == "" {
		return Setup{}, ErrInvalid
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO setups (user_id, car_id, track_id, name, notes, data, is_public)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, in.CarID, in.TrackID, in.Name, in.Notes, in.Data, in.Public)
	if err != nil {
		return Setup{}, err
	}
	id, _ := res.LastInsertId()
	return s.Get(ctx, userID, int(id), false)
}

// List returns the showroom: the user's own setups plus everyone's public ones,
// newest first. `mine` restricts it to the user's own. Data is not included.
func (s *Service) List(ctx context.Context, userID int64, mineOnly bool) ([]Setup, error) {
	where := "s.is_public = 1 OR s.user_id = ?"
	if mineOnly {
		where = "s.user_id = ?"
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.car_id, COALESCE(c.car_name, ''), s.track_id, COALESCE(t.track_name, ''),
		       s.name, s.notes, COALESCE(c.category, ''), COALESCE(u.nickname, u.email),
		       s.is_public, s.downloads, s.user_id = ? AS mine, s.created_at
		FROM setups s
		LEFT JOIN cars c   ON c.car_id = s.car_id
		LEFT JOIN tracks t ON t.track_id = s.track_id
		JOIN users u       ON u.id = s.user_id
		WHERE `+where+`
		ORDER BY s.created_at DESC, s.id DESC`, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Setup, 0)
	for rows.Next() {
		var st Setup
		if err := rows.Scan(&st.ID, &st.CarID, &st.CarName, &st.TrackID, &st.TrackName,
			&st.Name, &st.Notes, &st.Category, &st.Author,
			&st.Public, &st.Downloads, &st.Mine, &st.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, st)
	}
	return items, rows.Err()
}

// Get returns one setup if the user may see it (public or owner). When download
// is true and the requester is not the owner, the download counter is bumped and
// the setup's Data is included.
func (s *Service) Get(ctx context.Context, userID int64, id int, download bool) (Setup, error) {
	var st Setup
	var ownerID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.user_id, s.car_id, COALESCE(c.car_name, ''), s.track_id, COALESCE(t.track_name, ''),
		       s.name, s.notes, s.data, COALESCE(c.category, ''), COALESCE(u.nickname, u.email),
		       s.is_public, s.downloads, s.created_at
		FROM setups s
		LEFT JOIN cars c   ON c.car_id = s.car_id
		LEFT JOIN tracks t ON t.track_id = s.track_id
		JOIN users u       ON u.id = s.user_id
		WHERE s.id = ?`, id).Scan(
		&st.ID, &ownerID, &st.CarID, &st.CarName, &st.TrackID, &st.TrackName,
		&st.Name, &st.Notes, &st.Data, &st.Category, &st.Author,
		&st.Public, &st.Downloads, &st.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Setup{}, ErrNotFound
	}
	if err != nil {
		return Setup{}, err
	}
	st.Mine = ownerID == userID
	if !st.Public && !st.Mine {
		return Setup{}, ErrNotFound // don't reveal existence of private setups
	}
	if download && !st.Mine {
		if _, err := s.db.ExecContext(ctx,
			`UPDATE setups SET downloads = downloads + 1 WHERE id = ?`, id); err != nil {
			return Setup{}, err
		}
		st.Downloads++
	}
	return st, nil
}

// SetPublic publishes or unpublishes a setup the user owns.
func (s *Service) SetPublic(ctx context.Context, userID int64, id int, public bool) error {
	return s.own(ctx, userID, id, func() error {
		_, err := s.db.ExecContext(ctx,
			`UPDATE setups SET is_public = ? WHERE id = ?`, public, id)
		return err
	})
}

// Delete removes a setup the user owns.
func (s *Service) Delete(ctx context.Context, userID int64, id int) error {
	return s.own(ctx, userID, id, func() error {
		_, err := s.db.ExecContext(ctx, `DELETE FROM setups WHERE id = ?`, id)
		return err
	})
}

// own runs fn only if userID owns setup id, mapping missing/other rows to errors.
func (s *Service) own(ctx context.Context, userID int64, id int, fn func() error) error {
	var ownerID int64
	err := s.db.QueryRowContext(ctx, `SELECT user_id FROM setups WHERE id = ?`, id).Scan(&ownerID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if ownerID != userID {
		return ErrForbidden
	}
	return fn()
}
