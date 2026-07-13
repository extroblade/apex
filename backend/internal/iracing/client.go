// Package iracing is a small client for the official iRacing /data API
// (members-ng.iracing.com).
//
// Two things about that API shape the design:
//  1. Auth is OAuth 2.x (see oauth.go). A Client carries an access token and
//     sends it as `Authorization: Bearer`. (iRacing removed the old
//     email/password auth on 2025-12-09.)
//  2. Data endpoints don't return the payload directly — they return a small
//     JSON envelope with a signed link to the real JSON on S3. We follow it.
//
// NOTE: exact JSON field names below are per iRacing's published API. They
// should be re-verified against a live response when credentials are available;
// this package is unit-tested against a simulated server, not the real one.
package iracing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// DefaultBaseURL is the production API host.
const DefaultBaseURL = "https://members-ng.iracing.com"

// userAgent identifies this app to iRacing. iRacing REQUIRES a descriptive
// User-Agent; requests with none (or Go's default) are served a Cloudflare
// challenge/HTML block instead of JSON.
const userAgent = "RacingPlanner/1.0"

// Errors callers can branch on with errors.Is.
var (
	ErrAuth               = errors.New("iracing: authentication failed")
	ErrRateLimited        = errors.New("iracing: rate limited")
	ErrMaintenance        = errors.New("iracing: service in maintenance")
	ErrNotFound           = errors.New("iracing: resource not found")
	ErrUnexpectedResponse = errors.New("iracing: got a non-JSON response (Cloudflare/CAPTCHA challenge or maintenance page) — the iRacing account may need Legacy Authentication enabled, or a browser CAPTCHA is blocking headless login")
)

// Client talks to the iRacing data API. One Client carries one OAuth access
// token (set via SetToken) and sends it as a Bearer token on every request.
type Client struct {
	baseURL string
	http    *http.Client
	token   string
}

// SetToken sets the OAuth access token used for Data API requests.
func (c *Client) SetToken(accessToken string) { c.token = accessToken }

// NewClient returns a Client with its own cookie jar and a sane timeout.
func NewClient() *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		baseURL: DefaultBaseURL,
		http:    &http.Client{Jar: jar, Timeout: 30 * time.Second},
	}
}

// WithBaseURL overrides the API host (used by tests to point at httptest).
func (c *Client) WithBaseURL(u string) *Client {
	c.baseURL = strings.TrimRight(u, "/")
	return c
}

// get calls an authenticated data endpoint and decodes the payload into dest,
// transparently following the S3 link indirection.
func (c *Client) get(ctx context.Context, path string, params url.Values, dest any) error {
	endpoint := c.baseURL + path
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	body, err := c.rawGet(ctx, endpoint)
	if err != nil {
		return err
	}

	// The envelope may carry a signed link; if not, the body IS the payload.
	var env struct {
		Link string `json:"link"`
	}
	if err := json.Unmarshal(body, &env); err == nil && env.Link != "" {
		linked, err := c.rawGet(ctx, env.Link)
		if err != nil {
			return err
		}
		body = linked
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("iracing: decode %s: %w", path, err)
	}
	return nil
}

// rawGet performs a GET and returns the body, mapping HTTP status to errors.
func (c *Client) rawGet(ctx context.Context, rawURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := statusError(resp.StatusCode); err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if looksLikeHTML(body) {
		return nil, ErrUnexpectedResponse
	}
	return body, nil
}

// setHeaders applies the headers iRacing expects on every request, including
// the OAuth Bearer token when one is set.
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
}

// looksLikeHTML reports whether a body is (likely) an HTML page rather than
// JSON — e.g. a Cloudflare challenge or a maintenance page.
func looksLikeHTML(body []byte) bool {
	t := bytes.TrimSpace(body)
	return len(t) > 0 && t[0] == '<'
}

func statusError(code int) error {
	switch code {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrAuth
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusServiceUnavailable:
		return ErrMaintenance
	default:
		return fmt.Errorf("iracing: unexpected status %d", code)
	}
}
