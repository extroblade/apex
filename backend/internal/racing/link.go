package racing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"apex/internal/iracing"
)

// pendingTTL bounds how long an authorization request stays valid.
const pendingTTL = 10 * time.Minute

// Account is the public view of a user's linked iRacing account.
type Account struct {
	CustID      int       `json:"custId"`
	DisplayName string    `json:"displayName"`
	LinkedAt    time.Time `json:"linkedAt"`
}

// BeginLink starts the OAuth flow: it generates PKCE + state, remembers them for
// the user, and returns the iRacing authorize URL to redirect the browser to.
func (s *Service) BeginLink(userID int64) (string, error) {
	if !s.oauthReady() {
		return "", ErrOAuthDisabled
	}
	verifier, challenge, err := iracing.GeneratePKCE()
	if err != nil {
		return "", err
	}
	state, err := iracing.RandomState()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	for k, p := range s.pending { // opportunistic cleanup of stale requests
		if time.Since(p.created) > pendingTTL {
			delete(s.pending, k)
		}
	}
	s.pending[state] = pendingLink{userID: userID, verifier: verifier, created: time.Now()}
	s.mu.Unlock()

	return s.oauth.AuthorizeURL(state, challenge), nil
}

// CompleteLink finishes the OAuth flow: it validates the state, exchanges the
// code for tokens, resolves the member, and stores the refresh token encrypted.
func (s *Service) CompleteLink(ctx context.Context, state, code string) (Account, error) {
	s.mu.Lock()
	p, ok := s.pending[state]
	if ok {
		delete(s.pending, state)
	}
	s.mu.Unlock()
	if !ok || time.Since(p.created) > pendingTTL {
		return Account{}, ErrLinkState
	}

	tok, err := iracing.ExchangeCode(ctx, s.http, s.oauth, code, p.verifier)
	if err != nil {
		// Surface infrastructure errors verbatim; treat the rest as a link failure.
		if errors.Is(err, iracing.ErrMaintenance) ||
			errors.Is(err, iracing.ErrRateLimited) ||
			errors.Is(err, iracing.ErrUnexpectedResponse) {
			return Account{}, err
		}
		return Account{}, fmt.Errorf("%w: %v", ErrLinkFailed, err)
	}
	if tok.RefreshToken == "" {
		return Account{}, fmt.Errorf("%w: iRacing returned no refresh token", ErrLinkFailed)
	}

	client := s.factory()
	client.SetToken(tok.AccessToken)
	info, err := client.Info(ctx)
	if err != nil {
		return Account{}, err
	}

	enc, err := s.box.Encrypt([]byte(tok.RefreshToken))
	if err != nil {
		return Account{}, err
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO iracing_accounts (user_id, cust_id, display_name, refresh_token_enc)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			cust_id = VALUES(cust_id),
			display_name = VALUES(display_name),
			refresh_token_enc = VALUES(refresh_token_enc)`,
		p.userID, info.CustID, info.DisplayName, enc)
	if err != nil {
		return Account{}, err
	}

	s.cacheToken(p.userID, tok, info.CustID)
	return s.Status(ctx, p.userID)
}

// Unlink removes the stored account and drops any cached token.
func (s *Service) Unlink(ctx context.Context, userID int64) error {
	s.forgetSession(userID)
	_, err := s.db.ExecContext(ctx, `DELETE FROM iracing_accounts WHERE user_id = ?`, userID)
	return err
}

// Status returns the linked account, or ErrNotLinked.
func (s *Service) Status(ctx context.Context, userID int64) (Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx, `
		SELECT cust_id, display_name, linked_at
		FROM iracing_accounts WHERE user_id = ?`, userID).
		Scan(&a.CustID, &a.DisplayName, &a.LinkedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, ErrNotLinked
	}
	if err != nil {
		return Account{}, err
	}
	return a, nil
}

func (s *Service) loadRefreshToken(ctx context.Context, userID int64) (refreshEnc string, custID int, err error) {
	err = s.db.QueryRowContext(ctx, `
		SELECT refresh_token_enc, cust_id FROM iracing_accounts WHERE user_id = ?`, userID).
		Scan(&refreshEnc, &custID)
	return refreshEnc, custID, err
}
