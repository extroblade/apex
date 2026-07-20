package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// RequestPasswordReset generates a reset token for the account at `email` and
// emails it. It returns nil for ANY email (known or unknown) so the endpoint
// can't be used to enumerate accounts — the email either lands (known) or
// silently doesn't (unknown). The caller should always return 204.
func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	var userID int64
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM users WHERE email = ?`, email).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil // unknown email — return nil to avoid enumeration
	}
	if err != nil {
		return err
	}
	token, err := s.issueToken(ctx, userID, tokenReset)
	if err != nil {
		return err
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		return nil // no mailer configured — token is issued but not delivered (dev/test)
	}
	link := fmt.Sprintf("%s/reset-password/confirm?token=%s", s.baseURL, token)
	body := fmt.Sprintf(
		"Someone requested a password reset for your Apex account.\n\n"+
			"If that was you, reset it here:\n%s\n\n"+
			"This link expires in 24 hours. If you didn't request this, you can safely ignore this email — "+
			"your password is still unchanged.\n",
		link)
	return s.mailer.Send(ctx, email, "Reset your Apex password", body)
}

// ConfirmPasswordReset verifies the token and sets the new password. It revokes
// every session for the user (there's no "current session" to keep — the user
// is resetting because they don't have one).
func (s *Service) ConfirmPasswordReset(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < minPasswordLen {
		return ErrWeakPassword
	}
	userID, err := s.consumeToken(ctx, rawToken, tokenReset)
	if err != nil {
		return err
	}
	hash, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET password_hash = ? WHERE id = ?`, hash, userID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM sessions WHERE user_id = ?`, userID); err != nil {
		return err
	}
	return tx.Commit()
}

// RequestEmailVerification generates (or replaces) a verification token for the
// account at `email` and emails it. No-op (returns nil) if the email is unknown
// or already verified, so the endpoint can't enumerate accounts.
func (s *Service) RequestEmailVerification(ctx context.Context, email string) error {
	email = normalizeEmail(email)
	var (
		userID    int64
		verified  bool
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT id, email_verified FROM users WHERE email = ?`, email).Scan(&userID, &verified)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if verified {
		return nil // already verified — no need to send
	}
	token, err := s.issueToken(ctx, userID, tokenVerify)
	if err != nil {
		return err
	}
	if s.mailer == nil || !s.mailer.Enabled() {
		return nil
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)
	body := fmt.Sprintf(
		"Welcome to Apex! Confirm your email address to finish signing up:\n%s\n\n"+
			"This link expires in 24 hours. If you didn't create an Apex account, you can safely ignore this email.\n",
		link)
	return s.mailer.Send(ctx, email, "Confirm your Apex email", body)
}

// ConfirmEmailVerification validates the token and marks the user's email
// verified.
func (s *Service) ConfirmEmailVerification(ctx context.Context, rawToken string) error {
	userID, err := s.consumeToken(ctx, rawToken, tokenVerify)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx,
		`UPDATE users SET email_verified = 1 WHERE id = ?`, userID); err != nil {
		return err
	}
	return nil
}

// IsEmailVerified reports whether the user has confirmed their email. Returns
// true for users that don't exist (defensive) — the caller has already authed.
func (s *Service) IsEmailVerified(ctx context.Context, userID int64) bool {
	var verified bool
	err := s.db.QueryRowContext(ctx,
		`SELECT email_verified FROM users WHERE id = ?`, userID).Scan(&verified)
	if err != nil {
		return true // defensive: don't block an authed user on a query failure
	}
	return verified
}

// SendWelcomeVerification is called right after Register to email the
// verification link. It's best-effort: a failure here shouldn't fail the
// registration. The token is issued even if the mailer is disabled, so a later
// resend (once SMTP is configured) works.
func (s *Service) SendWelcomeVerification(ctx context.Context, userID int64, email string) {
	if s.mailer == nil || !s.mailer.Enabled() {
		return
	}
	token, err := s.issueToken(ctx, userID, tokenVerify)
	if err != nil {
		return
	}
	link := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)
	body := fmt.Sprintf(
		"Welcome to Apex! Confirm your email address to finish signing up:\n%s\n\n"+
			"This link expires in 24 hours. If you didn't create an Apex account, you can safely ignore this email.\n",
		link)
	_ = s.mailer.Send(ctx, email, "Confirm your Apex email", body)
}