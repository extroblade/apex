package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"time"
)

// Token kinds stored in email_tokens.
const (
	tokenVerify = "verify"
	tokenReset  = "reset"
)

// tokenTTL is how long a token is valid. 24h is long enough to survive a user
// checking email the next morning, short enough to limit the window if a link
// leaks.
const tokenTTL = 24 * time.Hour

var (
	ErrTokenInvalid = errors.New("invalid or expired token")
	ErrAlreadyVerified = errors.New("email already verified")
)

// newToken returns a random URL-safe token plus its SHA-256 hash (for storage).
// Same shape as session tokens: 32 random bytes, base64url-encoded.
func newToken() (token, hash string, err error) {
	raw := make([]byte, 32)
	if _, err = rand.Read(raw); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(raw)
	return token, hashToken(token), nil
}

// issueToken writes a fresh (token_hash, user_id, kind) row and returns the raw
// token to send to the user. Any previous token of the same kind for the user
// is deleted first so only one outstanding link per kind exists at a time.
func (s *Service) issueToken(ctx context.Context, userID int64, kind string) (string, error) {
	token, hash, err := newToken()
	if err != nil {
		return "", err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback() //nolint:errcheck
	// One outstanding token per (user, kind) — replace any previous.
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM email_tokens WHERE user_id = ? AND kind = ?`, userID, kind); err != nil {
		return "", err
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO email_tokens (token_hash, user_id, kind, expires_at) VALUES (?, ?, ?, ?)`,
		hash, userID, kind, time.Now().Add(tokenTTL)); err != nil {
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", err
	}
	return token, nil
}

// consumeToken validates the token (right kind, not expired, user exists) and
// deletes it. Returns the user_id on success.
func (s *Service) consumeToken(ctx context.Context, rawToken, kind string) (int64, error) {
	if rawToken == "" {
		return 0, ErrTokenInvalid
	}
	hash := hashToken(rawToken)
	var (
		userID    int64
		expiresAt time.Time
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT user_id, expires_at FROM email_tokens WHERE token_hash = ? AND kind = ?`,
		hash, kind).Scan(&userID, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrTokenInvalid
	}
	if err != nil {
		return 0, err
	}
	if time.Now().After(expiresAt) {
		// Clean up the expired row so the table doesn't accumulate dead tokens.
		_, _ = s.db.ExecContext(ctx, `DELETE FROM email_tokens WHERE token_hash = ? AND kind = ?`, hash, kind)
		return 0, ErrTokenInvalid
	}
	if _, err := s.db.ExecContext(ctx,
		`DELETE FROM email_tokens WHERE token_hash = ? AND kind = ?`, hash, kind); err != nil {
		return 0, err
	}
	return userID, nil
}
