package iracing

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// OAuth support for the iRacing Data API.
//
// iRacing removed legacy email/password auth on 2025-12-09; the Data API now
// requires OAuth 2.x. This app uses the authorization-code flow: the user logs
// in on iRacing's own page and we exchange the returned code for tokens.
//
// ASSUMPTIONS still to verify against iRacing's docs when registering a client:
//   - authorize path is <base>/authorize and token path is <base>/token
//     (base = https://oauth.iracing.com/oauth2)
//   - the scope is "iracing.auth" and the data-server audience is fixed at
//     client-registration time (not a runtime parameter)
//   - client_secret is transmitted SHA-256-masked as a form field (per the
//     docs' "secrets must be masked using SHA-256"); the exact encoding (hex
//     here) may need adjusting.
const (
	defaultOAuthBase = "https://oauth.iracing.com/oauth2"
	OAuthScope       = "iracing.auth"
)

// OAuthConfig holds the registered client's settings.
type OAuthConfig struct {
	ClientID     string
	ClientSecret string // optional — "only issued to some client types"
	RedirectURI  string
	BaseURL      string // override for tests; defaults to defaultOAuthBase
}

func (c OAuthConfig) base() string {
	if c.BaseURL != "" {
		return strings.TrimRight(c.BaseURL, "/")
	}
	return defaultOAuthBase
}

// Token is the /token response.
type Token struct {
	AccessToken           string `json:"access_token"`
	TokenType             string `json:"token_type"`
	ExpiresIn             int    `json:"expires_in"`
	RefreshToken          string `json:"refresh_token"`
	RefreshTokenExpiresIn int    `json:"refresh_token_expires_in"`
	Scope                 string `json:"scope"`
}

// AuthorizeURL builds the URL to send the user's browser to. codeChallenge is
// the PKCE S256 challenge; state is a CSRF/session-binding token.
func (c OAuthConfig) AuthorizeURL(state, codeChallenge string) string {
	v := url.Values{}
	v.Set("response_type", "code")
	v.Set("client_id", c.ClientID)
	v.Set("redirect_uri", c.RedirectURI)
	v.Set("scope", OAuthScope)
	v.Set("state", state)
	if codeChallenge != "" {
		v.Set("code_challenge", codeChallenge)
		v.Set("code_challenge_method", "S256")
	}
	return c.base() + "/authorize?" + v.Encode()
}

// ExchangeCode swaps an authorization code for tokens.
func ExchangeCode(ctx context.Context, httpClient *http.Client, cfg OAuthConfig, code, codeVerifier string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "authorization_code")
	form.Set("client_id", cfg.ClientID)
	form.Set("code", code)
	form.Set("redirect_uri", cfg.RedirectURI)
	if codeVerifier != "" {
		form.Set("code_verifier", codeVerifier)
	}
	if cfg.ClientSecret != "" {
		form.Set("client_secret", maskSecret(cfg.ClientSecret))
	}
	return postToken(ctx, httpClient, cfg, form)
}

// RefreshAccess exchanges a refresh token for a new access token.
func RefreshAccess(ctx context.Context, httpClient *http.Client, cfg OAuthConfig, refreshToken string) (Token, error) {
	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("client_id", cfg.ClientID)
	form.Set("refresh_token", refreshToken)
	if cfg.ClientSecret != "" {
		form.Set("client_secret", maskSecret(cfg.ClientSecret))
	}
	return postToken(ctx, httpClient, cfg, form)
}

func postToken(ctx context.Context, httpClient *http.Client, cfg OAuthConfig, form url.Values) (Token, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.base()+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return Token{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Token{}, err
	}
	if looksLikeHTML(body) {
		return Token{}, ErrUnexpectedResponse
	}
	if resp.StatusCode != http.StatusOK {
		return Token{}, fmt.Errorf("%w: token endpoint returned %d: %s", ErrAuth, resp.StatusCode, snippet(body))
	}

	var t Token
	if err := json.Unmarshal(body, &t); err != nil {
		return Token{}, fmt.Errorf("iracing: decode token response: %w", err)
	}
	if t.AccessToken == "" {
		return Token{}, fmt.Errorf("%w: token response had no access_token", ErrAuth)
	}
	return t, nil
}

// GeneratePKCE returns a random code_verifier and its S256 code_challenge.
func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(sum[:])
	return verifier, challenge, nil
}

// RandomState returns a random CSRF state value.
func RandomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func maskSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

func snippet(b []byte) string {
	const max = 200
	if len(b) > max {
		return string(b[:max])
	}
	return string(b)
}
