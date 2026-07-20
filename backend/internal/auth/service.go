package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/mail"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

// defaultAvatarCount is how many default avatars the static service provides.
const defaultAvatarCount = 8

// randomDefaultAvatar returns the path to a random default avatar served by the
// static service.
func randomDefaultAvatar() string {
	return fmt.Sprintf("/media/avatars/avatar-%d.svg", 1+rand.IntN(defaultAvatarCount))
}

// Register creates a new account. It validates input, hashes the password, and
// inserts the row — translating MySQL's duplicate-key error into ErrEmailTaken.
// It also kicks off the welcome/verification email (best-effort, never fails
// the registration).
func (s *Service) Register(ctx context.Context, email, password string) (User, error) {
	email = normalizeEmail(email)
	if !validEmail(email) {
		return User{}, ErrInvalidEmail
	}
	if len(password) < minPasswordLen {
		return User{}, ErrWeakPassword
	}

	hash, err := hashPassword(password)
	if err != nil {
		return User{}, err
	}

	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (email, password_hash, avatar_data_url) VALUES (?, ?, ?)`,
		email, hash, randomDefaultAvatar())
	if err != nil {
		// errors.As unwraps the error chain looking for a *mysql.MySQLError;
		// error 1062 is a duplicate-unique-key violation.
		var myErr *mysql.MySQLError
		if errors.As(err, &myErr) && myErr.Number == 1062 {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	// Best-effort: send the verification email. A failure here (no SMTP
	// configured, send error) must NOT fail the registration — the user can
	// resend from the profile page.
	s.SendWelcomeVerification(ctx, id, email)
	return s.userByID(ctx, id)
}

// Login verifies credentials and, on success, creates a session and returns its
// token (the caller sets it as a cookie).
func (s *Service) Login(ctx context.Context, email, password string) (token string, user User, err error) {
	email = normalizeEmail(email)

	var (
		hash     string
		avatar   sql.NullString
		pending  sql.NullString
		verified bool
	)
	err = s.db.QueryRowContext(ctx,
		`SELECT id, email, pending_email, nickname, avatar_data_url, password_hash, email_verified, created_at FROM users WHERE email = ?`, email).
		Scan(&user.ID, &user.Email, &pending, &user.Nickname, &avatar, &hash, &verified, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", User{}, ErrInvalidCredentials
	}
	user.AvatarURL = avatar.String
	user.PendingEmail = pending.String
	user.EmailVerified = verified
	if err != nil {
		return "", User{}, err
	}

	if !checkPassword(hash, password) {
		return "", User{}, ErrInvalidCredentials
	}

	token, err = s.createSession(ctx, user.ID)
	if err != nil {
		return "", User{}, err
	}
	return token, user, nil
}

// Authenticate resolves a session token to its user, or ErrUnauthorized if the
// token is missing, unknown, or expired.
func (s *Service) Authenticate(ctx context.Context, token string) (User, error) {
	if token == "" {
		return User{}, ErrUnauthorized
	}

	row := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.email, u.pending_email, u.nickname, u.avatar_data_url, u.email_verified, u.created_at
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.expires_at > NOW()`,
		hashToken(token))
	u, err := scanUser(row)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUnauthorized
	}
	if err != nil {
		return User{}, err
	}
	return u, nil
}

// Logout deletes the session so the token can no longer be used.
func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE token_hash = ?`, hashToken(token))
	return err
}

func (s *Service) createSession(ctx context.Context, userID int64) (string, error) {
	token, hash, err := newSessionToken()
	if err != nil {
		return "", err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO sessions (token_hash, user_id, expires_at) VALUES (?, ?, ?)`,
		hash, userID, time.Now().Add(sessionTTL))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (s *Service) userByID(ctx context.Context, id int64) (User, error) {
	return scanUser(s.db.QueryRowContext(ctx,
		`SELECT id, email, pending_email, nickname, avatar_data_url, email_verified, created_at FROM users WHERE id = ?`, id))
}

// scanUser reads the standard user column order, handling the nullable avatar
// and pending_email.
func scanUser(row interface{ Scan(dest ...any) error }) (User, error) {
	var (
		u       User
		avatar  sql.NullString
		pending sql.NullString
	)
	err := row.Scan(&u.ID, &u.Email, &pending, &u.Nickname, &avatar, &u.EmailVerified, &u.CreatedAt)
	u.AvatarURL = avatar.String
	u.PendingEmail = pending.String
	return u, err
}

// UpdateProfile changes the user's nickname and email.
func (s *Service) UpdateProfile(ctx context.Context, userID int64, nickname, email string) (User, error) {
	nickname = strings.TrimSpace(nickname)
	if len([]rune(nickname)) > 50 {
		return User{}, ErrNicknameTooLong
	}
	email = normalizeEmail(email)
	if !validEmail(email) {
		return User{}, ErrInvalidEmail
	}

	_, err := s.db.ExecContext(ctx,
		`UPDATE users SET nickname = ?, email = ? WHERE id = ?`, nickname, email, userID)
	if err != nil {
		var myErr *mysql.MySQLError
		if errors.As(err, &myErr) && myErr.Number == 1062 {
			return User{}, ErrEmailTaken
		}
		return User{}, err
	}
	return s.userByID(ctx, userID)
}

// UpdateAvatar sets (or, with an empty string, clears) the avatar data URL.
func (s *Service) UpdateAvatar(ctx context.Context, userID int64, dataURL string) (User, error) {
	var value any
	if dataURL == "" {
		value = nil
	} else {
		if !strings.HasPrefix(dataURL, "data:image/") {
			return User{}, ErrInvalidAvatar
		}
		if len(dataURL) > maxAvatarBytes {
			return User{}, ErrAvatarTooLarge
		}
		value = dataURL
	}

	if _, err := s.db.ExecContext(ctx,
		`UPDATE users SET avatar_data_url = ? WHERE id = ?`, value, userID); err != nil {
		return User{}, err
	}
	return s.userByID(ctx, userID)
}

// ChangePassword verifies the current password and sets a new one. It also
// revokes every OTHER session for the user (keepToken, the caller's current
// session, survives) so a leaked/old token can't outlive a password change.
func (s *Service) ChangePassword(ctx context.Context, userID int64, current, next, keepToken string) error {
	if len(next) < minPasswordLen {
		return ErrWeakPassword
	}

	var hash string
	if err := s.db.QueryRowContext(ctx,
		`SELECT password_hash FROM users WHERE id = ?`, userID).Scan(&hash); err != nil {
		return err
	}
	if !checkPassword(hash, current) {
		return ErrInvalidCredentials
	}

	newHash, err := hashPassword(next)
	if err != nil {
		return err
	}
	if _, err = s.db.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, userID); err != nil {
		return err
	}
	// Revoke all sessions except the current one.
	_, err = s.db.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = ? AND token_hash <> ?`, userID, hashToken(keepToken))
	return err
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validEmail(email string) bool {
	addr, err := mail.ParseAddress(email)
	return err == nil && addr.Address == email
}
