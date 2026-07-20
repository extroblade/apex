package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// DeleteAccount rejects a wrong password without deleting the row.
func TestDeleteAccount_WrongPassword(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).
			AddRow("$2a$10$abcdef")) // not a real bcrypt hash, but CompareHashAndPassword will fail
	// No DELETE expected — the wrong password must short-circuit.

	s := &Service{db: db}
	err := s.DeleteAccount(context.Background(), 42, "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("err = %v, want ErrInvalidCredentials", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// DeleteAccount with the correct password deletes the user row (cascade
// handles the rest).
func TestDeleteAccount_ValidPasswordDeletes(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	hash, _ := hashPassword("correct-password")
	mock.ExpectQuery("SELECT password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow(hash))
	mock.ExpectExec("DELETE FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := &Service{db: db}
	if err := s.DeleteAccount(context.Background(), 42, "correct-password"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// DeleteAccount on a missing user returns ErrUnauthorized.
func TestDeleteAccount_UnknownUser(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnError(sql.ErrNoRows)

	s := &Service{db: db}
	err := s.DeleteAccount(context.Background(), 42, "whatever")
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("err = %v, want ErrUnauthorized", err)
	}
}

// RequestEmailChange refuses the same email (no-op).
func TestRequestEmailChange_SameEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	hash, _ := hashPassword("correct")
	mock.ExpectQuery("SELECT email, password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"email", "password_hash"}).
			AddRow("user@x.com", hash))

	s := &Service{db: db}
	err := s.RequestEmailChange(context.Background(), 42, "user@x.com", "correct")
	if !errors.Is(err, ErrEmailSame) {
		t.Errorf("err = %v, want ErrEmailSame", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// RequestEmailChange refuses a taken email.
func TestRequestEmailChange_TakenEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	hash, _ := hashPassword("correct")
	mock.ExpectQuery("SELECT email, password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"email", "password_hash"}).
			AddRow("user@x.com", hash))
	mock.ExpectQuery("SELECT id FROM users WHERE email = \\?").
		WithArgs("taken@x.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))

	s := &Service{db: db}
	err := s.RequestEmailChange(context.Background(), 42, "taken@x.com", "correct")
	if !errors.Is(err, ErrEmailTaken) {
		t.Errorf("err = %v, want ErrEmailTaken", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// RequestEmailChange with a valid new email stages pending_email and issues
// a verify token with target_email set to the new address.
func TestRequestEmailChange_StagesAndIssuesToken(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	hash, _ := hashPassword("correct")
	mock.ExpectQuery("SELECT email, password_hash FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"email", "password_hash"}).
			AddRow("old@x.com", hash))
	mock.ExpectQuery("SELECT id FROM users WHERE email = \\?").
		WithArgs("new@x.com").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("UPDATE users SET pending_email = \\? WHERE id = \\?").
		WithArgs("new@x.com", int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM email_tokens WHERE user_id = \\? AND kind = \\?").
		WithArgs(int64(42), "verify").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_tokens").
		WithArgs(sqlmock.AnyArg(), int64(42), "verify", "new@x.com", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	m := &fakeMailer{enabled: true}
	s := &Service{db: db, mailer: m, baseURL: "https://app"}
	if err := s.RequestEmailChange(context.Background(), 42, "new@x.com", "correct"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if m.sentTo != "new@x.com" {
		t.Errorf("sent to = %q, want new@x.com", m.sentTo)
	}
	if !strings.Contains(m.sentBody, "https://app/verify-email?token=") {
		t.Errorf("body missing verify link: %q", m.sentBody)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ConfirmEmailVerification with a target_email promotes the pending email.
func TestConfirmEmailVerification_PromotesPendingEmail(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT user_id, expires_at, target_email FROM email_tokens").
		WithArgs(sqlmock.AnyArg(), "verify").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at", "target_email"}).
			AddRow(42, time.Now().Add(24*time.Hour), "new@x.com"))
	mock.ExpectExec("DELETE FROM email_tokens WHERE token_hash = \\? AND kind = \\?").
		WithArgs(sqlmock.AnyArg(), "verify").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET email = \\?, email_verified = 1, pending_email = NULL WHERE id = \\?").
		WithArgs("new@x.com", int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := &Service{db: db}
	if err := s.ConfirmEmailVerification(context.Background(), "valid-token"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// CancelEmailChange clears pending_email and deletes the outstanding token.
func TestCancelEmailChange_ClearsPending(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET pending_email = NULL WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM email_tokens WHERE user_id = \\? AND kind = \\?").
		WithArgs(int64(42), "verify").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := &Service{db: db}
	if err := s.CancelEmailChange(context.Background(), 42); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// PendingEmail returns the staged email (or "" if none).
func TestPendingEmail_ReturnsStaged(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT pending_email FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"pending_email"}).
			AddRow("new@x.com"))

	s := &Service{db: db}
	got, err := s.PendingEmail(context.Background(), 42)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if got != "new@x.com" {
		t.Errorf("pending = %q, want new@x.com", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ExportData assembles the user's data. Verifies the profile section is read
// and the iRacing link is omitted when not linked.
func TestExportData_ProfileAndNoIRacing(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT id, email, pending_email, nickname, avatar_data_url, email_verified, created_at FROM users WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "email", "pending_email", "nickname", "avatar_data_url", "email_verified", "created_at"}).
			AddRow(42, "user@x.com", nil, "Racer", nil, true, time.Now()))
	// Owned cars: empty
	mock.ExpectQuery("SELECT oc.car_id, c.car_name FROM owned_cars").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"car_id", "car_name"}))
	// Owned tracks: empty
	mock.ExpectQuery("SELECT ot.track_id, t.track_name FROM owned_tracks").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"track_id", "track_name"}))
	// Favorite series: empty
	mock.ExpectQuery("SELECT fs.series_id, s.series_name FROM favorite_series").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"series_id", "series_name"}))
	// Planned races: empty
	mock.ExpectQuery("FROM planned_races").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"series_id", "series_name", "week", "race_date", "track_name"}))
	// Setups: empty
	mock.ExpectQuery("FROM setups WHERE user_id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "car_id", "track_id", "notes", "data", "is_public", "downloads", "created_at", "updated_at"}))
	// Goals: empty
	mock.ExpectQuery("FROM goals WHERE user_id = \\?").
		WithArgs(int64(42)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "notes", "unit", "target", "current", "done", "due_date", "created_at", "updated_at"}))
	// iRacing: not linked
	mock.ExpectQuery("SELECT cust_id, display_name, linked_at, updated_at FROM iracing_accounts WHERE user_id = \\?").
		WithArgs(int64(42)).
		WillReturnError(sql.ErrNoRows)

	s := &Service{db: db}
	data, err := s.ExportData(context.Background(), 42)
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if data.Profile.Email != "user@x.com" {
		t.Errorf("email = %q, want user@x.com", data.Profile.Email)
	}
	if data.Profile.Nickname != "Racer" {
		t.Errorf("nickname = %q, want Racer", data.Profile.Nickname)
	}
	if data.IRacing != nil {
		t.Errorf("iRacing = %+v, want nil", data.IRacing)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
