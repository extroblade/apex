package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ErrEmailSame is returned when the new email matches the current one — no
// change needed.
var ErrEmailSame = errors.New("new email is the same as the current email")

// DeleteAccount verifies the current password and deletes the user's account.
// All related rows (sessions, garage, plans, setups, goals, iRacing link,
// email tokens) are removed by ON DELETE CASCADE foreign keys.
func (s *Service) DeleteAccount(ctx context.Context, userID int64, currentPassword string) error {
	var hash string
	if err := s.db.QueryRowContext(ctx,
		`SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrUnauthorized
		}
		return err
	}
	if !checkPassword(hash, currentPassword) {
		return ErrInvalidCredentials
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM users WHERE id = ?`, userID)
	return err
}

// RequestEmailChange verifies the current password, stages `newEmail` as the
// user's pending_email, and sends a verification link to the NEW address. The
// old email stays valid for login until the new one is confirmed — a typo'd
// change can't lock the user out. The new email is not yet stored on the user
// row as `email` (only as `pending_email` and on the token), so a duplicate-
// email check happens here against other users; the actual swap on confirm
// re-checks uniqueness via the unique index on users.email.
func (s *Service) RequestEmailChange(ctx context.Context, userID int64, newEmail, currentPassword string) error {
	newEmail = normalizeEmail(newEmail)
	if !validEmail(newEmail) {
		return ErrInvalidEmail
	}

	var (
		hash         string
		currentEmail string
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT email, password_hash FROM users WHERE id = ?`, userID).Scan(&currentEmail, &hash)
	if err != nil {
		return err
	}
	if !checkPassword(hash, currentPassword) {
		return ErrInvalidCredentials
	}
	if newEmail == currentEmail {
		return ErrEmailSame
	}
	// Refuse if the new email is already taken by another account.
	var existing int64
	err = s.db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE email = ?`, newEmail).Scan(&existing)
	if err == nil {
		return ErrEmailTaken
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Stage the pending email so the UI can show "verification pending for X".
	if _, err := s.db.ExecContext(ctx,
		`UPDATE users SET pending_email = ? WHERE id = ?`, newEmail, userID); err != nil {
		return err
	}

	token, err := s.issueToken(ctx, userID, tokenVerify, newEmail)
	if err != nil {
		return err
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		return nil // token issued but not delivered (dev/test)
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)
	body := fmt.Sprintf(
		"Someone requested changing the email on your Apex account to this address.\n\n"+
			"If that was you, confirm it here:\n%s\n\n"+
			"This link expires in 24 hours. Until you confirm, your current email stays active. "+
			"If you didn't request this, you can safely ignore this email.\n",
		link)
	return s.mailer.Send(ctx, newEmail, "Confirm your new Apex email", body)
}

// CancelEmailChange clears any staged pending_email and discards any
// outstanding verify token for the user. Lets a user undo a mistaken email
// change request from the profile page without waiting for the token to
// expire.
func (s *Service) CancelEmailChange(ctx context.Context, userID int64) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET pending_email = NULL WHERE id = ?`, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM email_tokens WHERE user_id = ? AND kind = ?`, userID, tokenVerify); err != nil {
		return err
	}
	return tx.Commit()
}

// PendingEmail returns the staged new email for the user ("" if none).
func (s *Service) PendingEmail(ctx context.Context, userID int64) (string, error) {
	var pending sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT pending_email FROM users WHERE id = ?`, userID).Scan(&pending)
	if err != nil {
		return "", err
	}
	return pending.String, nil
}

// AccountData is the user's full data export (GDPR). All collections are
// user-owned rows only; nothing from other users is included. The struct is
// JSON-marshaled as the export response body.
type AccountData struct {
	ExportedAt   time.Time      `json:"exportedAt"`
	Profile      ProfileExport  `json:"profile"`
	Garage       GarageExport   `json:"garage"`
	PlannedRaces []PlannedRace  `json:"plannedRaces"`
	Setups       []SetupExport  `json:"setups"`
	Goals        []GoalExport   `json:"goals"`
	IRacing      *IRacingExport `json:"iracing,omitempty"`
}

type ProfileExport struct {
	ID            int64     `json:"id"`
	Email         string    `json:"email"`
	PendingEmail  string    `json:"pendingEmail,omitempty"`
	Nickname      string    `json:"nickname"`
	EmailVerified bool      `json:"emailVerified"`
	CreatedAt     time.Time `json:"createdAt"`
}

type GarageExport struct {
	Cars           []OwnedCar       `json:"cars"`
	Tracks         []OwnedTrack     `json:"tracks"`
	FavoriteSeries []FavoriteSeries `json:"favoriteSeries"`
}

type OwnedCar struct {
	CarID int64  `json:"carId"`
	Name  string `json:"name"`
}

type OwnedTrack struct {
	TrackID int64  `json:"trackId"`
	Name    string `json:"name"`
}

type FavoriteSeries struct {
	SeriesID int64  `json:"seriesId"`
	Name     string `json:"name"`
}

type PlannedRace struct {
	SeriesID   int64  `json:"seriesId"`
	SeriesName string `json:"seriesName"`
	Week       int    `json:"week"`
	RaceDate   string `json:"raceDate"`
	TrackName  string `json:"trackName"`
}

type SetupExport struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CarID     int64     `json:"carId"`
	TrackID   int64     `json:"trackId"`
	Notes     string    `json:"notes"`
	Data      string    `json:"data"`
	Public    bool      `json:"public"`
	Downloads int       `json:"downloads"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GoalExport struct {
	ID        int64      `json:"id"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes"`
	Unit      string     `json:"unit"`
	Target    float64    `json:"target"`
	Current   float64    `json:"current"`
	Done      bool       `json:"done"`
	DueDate   *time.Time `json:"dueDate,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type IRacingExport struct {
	CustID      int64     `json:"custId"`
	DisplayName string    `json:"displayName"`
	LinkedAt    time.Time `json:"linkedAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ExportData assembles the user's full data export. It's read-only and never
// includes secrets (no password hash, no OAuth tokens, no session tokens).
// The iRacing section is nil if the user hasn't linked an account.
func (s *Service) ExportData(ctx context.Context, userID int64) (AccountData, error) {
	out := AccountData{ExportedAt: time.Now().UTC()}

	// Profile
	var (
		avatar  sql.NullString
		pending sql.NullString
	)
	p := ProfileExport{ID: userID}
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email, pending_email, nickname, avatar_data_url, email_verified, created_at
		 FROM users WHERE id = ?`, userID).
		Scan(&p.ID, &p.Email, &pending, &p.Nickname, &avatar, &p.EmailVerified, &p.CreatedAt)
	if err != nil {
		return out, err
	}
	p.PendingEmail = pending.String
	_ = avatar // avatar is intentionally not in the export (it's a data URL, large)
	out.Profile = p

	// Garage: owned cars
	carRows, err := s.db.QueryContext(ctx,
		`SELECT oc.car_id, c.car_name FROM owned_cars oc
		 JOIN cars c ON c.car_id = oc.car_id
		 WHERE oc.user_id = ? ORDER BY c.car_name`, userID)
	if err != nil {
		return out, err
	}
	for carRows.Next() {
		var c OwnedCar
		if err := carRows.Scan(&c.CarID, &c.Name); err != nil {
			carRows.Close()
			return out, err
		}
		out.Garage.Cars = append(out.Garage.Cars, c)
	}
	carRows.Close()

	// Garage: owned tracks
	trackRows, err := s.db.QueryContext(ctx,
		`SELECT ot.track_id, t.track_name FROM owned_tracks ot
		 JOIN tracks t ON t.track_id = ot.track_id
		 WHERE ot.user_id = ? ORDER BY t.track_name`, userID)
	if err != nil {
		return out, err
	}
	for trackRows.Next() {
		var t OwnedTrack
		if err := trackRows.Scan(&t.TrackID, &t.Name); err != nil {
			trackRows.Close()
			return out, err
		}
		out.Garage.Tracks = append(out.Garage.Tracks, t)
	}
	trackRows.Close()

	// Garage: favorite series
	favRows, err := s.db.QueryContext(ctx,
		`SELECT fs.series_id, s.series_name FROM favorite_series fs
		 JOIN series s ON s.series_id = fs.series_id
		 WHERE fs.user_id = ? ORDER BY s.series_name`, userID)
	if err != nil {
		return out, err
	}
	for favRows.Next() {
		var f FavoriteSeries
		if err := favRows.Scan(&f.SeriesID, &f.Name); err != nil {
			favRows.Close()
			return out, err
		}
		out.Garage.FavoriteSeries = append(out.Garage.FavoriteSeries, f)
	}
	favRows.Close()

	// Planned races (joined to season_schedule for track + race_date)
	planRows, err := s.db.QueryContext(ctx,
		`SELECT pr.series_id, s.series_name, pr.week,
		        COALESCE(DATE_FORMAT(ss.race_date, '%Y-%m-%d'), ''),
		        t.track_name
		 FROM planned_races pr
		 JOIN series s ON s.series_id = pr.series_id
		 LEFT JOIN season_schedule ss ON ss.series_id = pr.series_id AND ss.week = pr.week
		 LEFT JOIN tracks t ON t.track_id = ss.track_id
		 WHERE pr.user_id = ? ORDER BY ss.race_date, s.series_name`, userID)
	if err != nil {
		return out, err
	}
	for planRows.Next() {
		var pr PlannedRace
		if err := planRows.Scan(&pr.SeriesID, &pr.SeriesName, &pr.Week, &pr.RaceDate, &pr.TrackName); err != nil {
			planRows.Close()
			return out, err
		}
		out.PlannedRaces = append(out.PlannedRaces, pr)
	}
	planRows.Close()

	// Setups
	setupRows, err := s.db.QueryContext(ctx,
		`SELECT id, name, car_id, track_id, notes, data, is_public, downloads, created_at, updated_at
		 FROM setups WHERE user_id = ? ORDER BY created_at`, userID)
	if err != nil {
		return out, err
	}
	for setupRows.Next() {
		var su SetupExport
		if err := setupRows.Scan(&su.ID, &su.Name, &su.CarID, &su.TrackID, &su.Notes, &su.Data,
			&su.Public, &su.Downloads, &su.CreatedAt, &su.UpdatedAt); err != nil {
			setupRows.Close()
			return out, err
		}
		out.Setups = append(out.Setups, su)
	}
	setupRows.Close()

	// Goals
	goalRows, err := s.db.QueryContext(ctx,
		`SELECT id, title, notes, unit, target, current, done, due_date, created_at, updated_at
		 FROM goals WHERE user_id = ? ORDER BY created_at`, userID)
	if err != nil {
		return out, err
	}
	for goalRows.Next() {
		var g GoalExport
		var due sql.NullString
		if err := goalRows.Scan(&g.ID, &g.Title, &g.Notes, &g.Unit, &g.Target, &g.Current,
			&g.Done, &due, &g.CreatedAt, &g.UpdatedAt); err != nil {
			goalRows.Close()
			return out, err
		}
		if due.Valid {
			// DATE comes back as YYYY-MM-DD (possibly with a time suffix); parse the date part.
			d := due.String
			if len(d) >= 10 {
				d = d[:10]
			}
			if t, err := time.Parse("2006-01-02", d); err == nil {
				g.DueDate = &t
			}
		}
		out.Goals = append(out.Goals, g)
	}
	goalRows.Close()

	// iRacing link (no tokens, just public metadata)
	var ir IRacingExport
	err = s.db.QueryRowContext(ctx,
		`SELECT cust_id, display_name, linked_at, updated_at FROM iracing_accounts WHERE user_id = ?`,
		userID).Scan(&ir.CustID, &ir.DisplayName, &ir.LinkedAt, &ir.UpdatedAt)
	if err == nil {
		out.IRacing = &ir
	} else if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}

	return out, nil
}

// MarshalExport returns the user's data as pretty-printed JSON. Used by the
// handler to write the download response.
func (s *Service) MarshalExport(ctx context.Context, userID int64) ([]byte, error) {
	data, err := s.ExportData(ctx, userID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(data, "", "  ")
}
