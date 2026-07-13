// Package racing is the application service for iRacing features: linking a
// user's iRacing account (OAuth), caching access tokens, serving dashboard
// stats, driver lookups, syncing race history, and computing comparators.
package racing

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"sync"
	"time"

	"apex/internal/iracing"
	"apex/internal/secretbox"
)

var (
	ErrNotLinked     = errors.New("racing: no linked iRacing account")
	ErrLinkFailed    = errors.New("racing: could not complete iRacing authorization")
	ErrLinkState     = errors.New("racing: invalid or expired authorization request")
	ErrOAuthDisabled = errors.New("racing: iRacing OAuth is not configured")
)

// oauthReady reports whether the OAuth-dependent features can run (encryption
// box present for refresh tokens, and a client id configured).
func (s *Service) oauthReady() bool {
	return s.box != nil && s.oauth.ClientID != ""
}

// APIClient is the slice of the iRacing client this service depends on.
// Programming to this interface lets tests inject a fake with no network.
type APIClient interface {
	SetToken(accessToken string)
	Info(ctx context.Context) (iracing.Member, error)
	Member(ctx context.Context, custID int) (iracing.Member, error)
	Career(ctx context.Context, custID int) ([]iracing.CareerStat, error)
	RecentRaces(ctx context.Context, custID int) ([]iracing.RecentRace, error)
	SearchDrivers(ctx context.Context, term string) ([]iracing.DriverSearchResult, error)
	Cars(ctx context.Context) ([]iracing.Car, error)
	Tracks(ctx context.Context) ([]iracing.CatalogTrack, error)
	Series(ctx context.Context) ([]iracing.CatalogSeries, error)
}

// Compile-time assertion that the real client satisfies the interface.
var _ APIClient = (*iracing.Client)(nil)

// Factory creates a fresh client with no token set.
type Factory func() APIClient

// DefaultFactory returns real iRacing clients.
func DefaultFactory() APIClient { return iracing.NewClient() }

// Service owns the DB, the encryption box, the OAuth config, an access-token
// cache, and the in-flight authorization requests.
type Service struct {
	db      *sql.DB
	box     *secretbox.Box
	factory Factory
	oauth   iracing.OAuthConfig
	http    *http.Client

	mu       sync.Mutex
	sessions map[int64]*cachedToken // access token per app user id
	pending  map[string]pendingLink // OAuth state -> link request
}

type cachedToken struct {
	accessToken string
	custID      int
	expires     time.Time
}

type pendingLink struct {
	userID   int64
	verifier string
	created  time.Time
}

func NewService(db *sql.DB, box *secretbox.Box, factory Factory, oauth iracing.OAuthConfig) *Service {
	return &Service{
		db:       db,
		box:      box,
		factory:  factory,
		oauth:    oauth,
		http:     &http.Client{Timeout: 30 * time.Second},
		sessions: make(map[int64]*cachedToken),
		pending:  make(map[string]pendingLink),
	}
}

// clientFor returns an iRacing client carrying a valid access token for the user.
func (s *Service) clientFor(ctx context.Context, userID int64) (APIClient, int, error) {
	token, custID, err := s.accessToken(ctx, userID)
	if err != nil {
		return nil, 0, err
	}
	c := s.factory()
	c.SetToken(token)
	return c, custID, nil
}

// accessToken returns a cached, unexpired access token, or mints a fresh one
// from the stored refresh token.
func (s *Service) accessToken(ctx context.Context, userID int64) (string, int, error) {
	s.mu.Lock()
	if t, ok := s.sessions[userID]; ok && time.Now().Before(t.expires) {
		token, custID := t.accessToken, t.custID
		s.mu.Unlock()
		return token, custID, nil
	}
	s.mu.Unlock()

	// Minting a fresh token needs the encryption box and OAuth client.
	if !s.oauthReady() {
		return "", 0, ErrOAuthDisabled
	}

	refreshEnc, custID, err := s.loadRefreshToken(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, ErrNotLinked
	}
	if err != nil {
		return "", 0, err
	}
	refresh, err := s.box.Decrypt(refreshEnc)
	if err != nil {
		return "", 0, err
	}

	tok, err := iracing.RefreshAccess(ctx, s.http, s.oauth, string(refresh))
	if err != nil {
		return "", 0, err
	}

	// Refresh tokens may rotate; persist the new one if so.
	if tok.RefreshToken != "" && tok.RefreshToken != string(refresh) {
		if enc, e := s.box.Encrypt([]byte(tok.RefreshToken)); e == nil {
			_, _ = s.db.ExecContext(ctx,
				`UPDATE iracing_accounts SET refresh_token_enc = ? WHERE user_id = ?`, enc, userID)
		}
	}

	s.cacheToken(userID, tok, custID)
	return tok.AccessToken, custID, nil
}

func (s *Service) cacheToken(userID int64, tok iracing.Token, custID int) {
	ttl := time.Duration(tok.ExpiresIn) * time.Second
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	// Refresh a little early so a request never races the expiry.
	expires := time.Now().Add(ttl - 30*time.Second)
	s.mu.Lock()
	s.sessions[userID] = &cachedToken{accessToken: tok.AccessToken, custID: custID, expires: expires}
	s.mu.Unlock()
}

func (s *Service) forgetSession(userID int64) {
	s.mu.Lock()
	delete(s.sessions, userID)
	s.mu.Unlock()
}
