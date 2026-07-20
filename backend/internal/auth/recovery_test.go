package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

// fakeMailer captures the last Send for assertions.
type fakeMailer struct {
	enabled  bool
	sentTo   string
	sentSubj string
	sentBody string
	sendErr  error
}

func (f *fakeMailer) Enabled() bool { return f.enabled }
func (f *fakeMailer) Send(_ context.Context, to, subject, body string) error {
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sentTo, f.sentSubj, f.sentBody = to, subject, body
	return nil
}

// RequestPasswordReset must NOT reveal whether an email is registered: it
// returns nil for both known and unknown emails.
func TestRequestPasswordReset_NoEnumeration(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	// Unknown email — no row, no token insert, no mail.
	mock.ExpectQuery("SELECT id FROM users WHERE email = ?").
		WithArgs("nobody@x.com").
		WillReturnError(sql.ErrNoRows)
	// No further expectations: nothing should be written or sent.

	s := &Service{db: db, mailer: &fakeMailer{enabled: true}, baseURL: "https://app"}
	if err := s.RequestPasswordReset(context.Background(), "nobody@x.com"); err != nil {
		t.Errorf("unknown email: err = %v, want nil (no enumeration)", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// RequestPasswordReset for a known email issues a token and sends a link.
func TestRequestPasswordReset_KnownEmailSendsLink(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT id FROM users WHERE email = ?").
		WithArgs("user@x.com").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(42))
	// issueToken: delete previous, insert new.
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM email_tokens WHERE user_id = \\? AND kind = \\?").
		WithArgs(int64(42), "reset").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO email_tokens").
		WithArgs(sqlmock.AnyArg(), int64(42), "reset", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	m := &fakeMailer{enabled: true}
	s := &Service{db: db, mailer: m, baseURL: "https://app"}
	if err := s.RequestPasswordReset(context.Background(), "user@x.com"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if m.sentTo != "user@x.com" {
		t.Errorf("sent to = %q, want user@x.com", m.sentTo)
	}
	if m.sentBody == "" || !contains(m.sentBody, "https://app/reset-password/confirm?token=") {
		t.Errorf("body missing reset link: %q", m.sentBody)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ConfirmPasswordReset rejects an unknown/expired token.
func TestConfirmPasswordReset_BadToken(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT user_id, expires_at FROM email_tokens").
		WithArgs(sqlmock.AnyArg(), "reset").
		WillReturnError(sql.ErrNoRows)

	s := &Service{db: db}
	err := s.ConfirmPasswordReset(context.Background(), "garbage", "newpassword123")
	if !errors.Is(err, ErrTokenInvalid) {
		t.Errorf("err = %v, want ErrTokenInvalid", err)
	}
}

// ConfirmPasswordReset with a valid token updates the password and revokes
// every session.
func TestConfirmPasswordReset_ValidTokenUpdatesAndRevokes(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()

	mock.ExpectQuery("SELECT user_id, expires_at FROM email_tokens").
		WithArgs(sqlmock.AnyArg(), "reset").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).
			AddRow(42, time.Now().Add(24*time.Hour)))
	mock.ExpectExec("DELETE FROM email_tokens WHERE token_hash = \\? AND kind = \\?").
		WithArgs(sqlmock.AnyArg(), "reset").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE users SET password_hash = \\? WHERE id = \\?").
		WithArgs(sqlmock.AnyArg(), int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM sessions WHERE user_id = \\?").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	s := &Service{db: db}
	if err := s.ConfirmPasswordReset(context.Background(), "valid-token", "newpassword123"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ConfirmPasswordReset rejects a too-short new password before hitting the DB.
func TestConfirmPasswordReset_WeakPassword(t *testing.T) {
	s := &Service{}
	if err := s.ConfirmPasswordReset(context.Background(), "tok", "short"); !errors.Is(err, ErrWeakPassword) {
		t.Errorf("err = %v, want ErrWeakPassword", err)
	}
}

// RequestEmailVerification is a no-op for an already-verified email.
func TestRequestEmailVerification_AlreadyVerifiedNoOp(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT id, email_verified FROM users WHERE email = ?").
		WithArgs("user@x.com").
		WillReturnRows(sqlmock.NewRows([]string{"id", "email_verified"}).AddRow(42, true))
	// No token insert, no send.

	m := &fakeMailer{enabled: true}
	s := &Service{db: db, mailer: m, baseURL: "https://app"}
	if err := s.RequestEmailVerification(context.Background(), "user@x.com"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if m.sentTo != "" {
		t.Errorf("already-verified user should not get an email, got %q", m.sentTo)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

// ConfirmEmailVerification marks the user verified.
func TestConfirmEmailVerification_MarksVerified(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery("SELECT user_id, expires_at FROM email_tokens").
		WithArgs(sqlmock.AnyArg(), "verify").
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "expires_at"}).
			AddRow(42, time.Now().Add(24*time.Hour)))
	mock.ExpectExec("DELETE FROM email_tokens WHERE token_hash = \\? AND kind = \\?").
		WithArgs(sqlmock.AnyArg(), "verify").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("UPDATE users SET email_verified = 1 WHERE id = \\?").
		WithArgs(int64(42)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s := &Service{db: db}
	if err := s.ConfirmEmailVerification(context.Background(), "valid-token"); err != nil {
		t.Fatalf("err = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
