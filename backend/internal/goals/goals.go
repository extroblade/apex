// Package goals is the goal tracker: users set numeric personal goals (wins,
// iRating, a license class) and update their progress. A goal auto-completes
// when current reaches target, but can also be toggled done manually.
package goals

import (
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrNotFound = errors.New("goal not found")
	ErrInvalid  = errors.New("invalid goal")
)

// Service is the goals store.
type Service struct {
	db *sql.DB
}

// New returns a goals service backed by db.
func New(db *sql.DB) *Service { return &Service{db: db} }

// Goal is a tracked goal with derived progress.
type Goal struct {
	ID       int     `json:"id"`
	Title    string  `json:"title"`
	Notes    string  `json:"notes"`
	Unit     string  `json:"unit"`
	Target   float64 `json:"target"`
	Current  float64 `json:"current"`
	Done     bool    `json:"done"`
	DueDate  *string `json:"dueDate"`  // YYYY-MM-DD or null
	Progress float64 `json:"progress"` // 0..1, derived
	Created  string  `json:"createdAt"`
}

// Input is the create/update payload.
type Input struct {
	Title   string
	Notes   string
	Unit    string
	Target  float64
	Current float64
	Done    *bool
	DueDate *string
}

func validate(in Input) (Input, error) {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return in, ErrInvalid
	}
	if in.Target < 0 || in.Current < 0 {
		return in, ErrInvalid
	}
	if in.DueDate != nil && strings.TrimSpace(*in.DueDate) == "" {
		in.DueDate = nil // treat empty string as "no date"
	}
	return in, nil
}

// autoDone reports the stored done flag: an explicit value wins, otherwise it's
// derived from reaching the target (target 0 means "no numeric target").
func autoDone(in Input) bool {
	if in.Done != nil {
		return *in.Done
	}
	return in.Target > 0 && in.Current >= in.Target
}

// Create inserts a goal owned by userID.
func (s *Service) Create(ctx context.Context, userID int64, in Input) (Goal, error) {
	in, err := validate(in)
	if err != nil {
		return Goal{}, err
	}
	res, err := s.db.ExecContext(ctx, `
		INSERT INTO goals (user_id, title, notes, unit, target, current, done, due_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		userID, in.Title, in.Notes, in.Unit, in.Target, in.Current, autoDone(in), in.DueDate)
	if err != nil {
		return Goal{}, err
	}
	id, _ := res.LastInsertId()
	return s.get(ctx, userID, int(id))
}

// Update replaces a goal the user owns.
func (s *Service) Update(ctx context.Context, userID int64, id int, in Input) (Goal, error) {
	in, err := validate(in)
	if err != nil {
		return Goal{}, err
	}
	res, err := s.db.ExecContext(ctx, `
		UPDATE goals SET title = ?, notes = ?, unit = ?, target = ?, current = ?, done = ?, due_date = ?
		WHERE id = ? AND user_id = ?`,
		in.Title, in.Notes, in.Unit, in.Target, in.Current, autoDone(in), in.DueDate, id, userID)
	if err != nil {
		return Goal{}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		// Either the goal doesn't exist or isn't the user's; both look "not found".
		if !s.exists(ctx, userID, id) {
			return Goal{}, ErrNotFound
		}
	}
	return s.get(ctx, userID, id)
}

// List returns the user's goals, open ones first then completed, newest first.
func (s *Service) List(ctx context.Context, userID int64) ([]Goal, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, title, notes, unit, target, current, done, due_date, created_at
		FROM goals WHERE user_id = ?
		ORDER BY done ASC, created_at DESC, id DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]Goal, 0)
	for rows.Next() {
		g, err := scanGoal(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, g)
	}
	return items, rows.Err()
}

// Delete removes a goal the user owns.
func (s *Service) Delete(ctx context.Context, userID int64, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM goals WHERE id = ? AND user_id = ?`, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Service) get(ctx context.Context, userID int64, id int) (Goal, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, title, notes, unit, target, current, done, due_date, created_at
		FROM goals WHERE id = ? AND user_id = ?`, id, userID)
	g, err := scanGoal(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Goal{}, ErrNotFound
	}
	return g, err
}

func (s *Service) exists(ctx context.Context, userID int64, id int) bool {
	var one int
	err := s.db.QueryRowContext(ctx,
		`SELECT 1 FROM goals WHERE id = ? AND user_id = ?`, id, userID).Scan(&one)
	return err == nil
}

// scanner is implemented by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanGoal(sc scanner) (Goal, error) {
	var (
		g   Goal
		due sql.NullString
	)
	if err := sc.Scan(&g.ID, &g.Title, &g.Notes, &g.Unit, &g.Target, &g.Current,
		&g.Done, &due, &g.Created); err != nil {
		return Goal{}, err
	}
	if due.Valid {
		// DATE comes back as YYYY-MM-DD (possibly with a time suffix); keep the date.
		d := due.String
		if len(d) >= 10 {
			d = d[:10]
		}
		g.DueDate = &d
	}
	g.Progress = progress(g.Current, g.Target, g.Done)
	return g, nil
}

// progress is a 0..1 completion ratio; a done goal is always 1.
func progress(current, target float64, done bool) float64 {
	if done {
		return 1
	}
	if target <= 0 {
		return 0
	}
	p := current / target
	if p < 0 {
		return 0
	}
	if p > 1 {
		return 1
	}
	return p
}
