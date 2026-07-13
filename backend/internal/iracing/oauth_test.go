package iracing

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthorizeURL(t *testing.T) {
	cfg := OAuthConfig{ClientID: "abc", RedirectURI: "https://app/cb", BaseURL: "https://oauth.example/oauth2"}
	raw := cfg.AuthorizeURL("state123", "chal456")

	u, err := url.Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	checks := map[string]string{
		"response_type":         "code",
		"client_id":             "abc",
		"redirect_uri":          "https://app/cb",
		"scope":                 OAuthScope,
		"state":                 "state123",
		"code_challenge":        "chal456",
		"code_challenge_method": "S256",
	}
	for k, want := range checks {
		if q.Get(k) != want {
			t.Errorf("param %s: want %q, got %q", k, want, q.Get(k))
		}
	}
	if !strings.HasPrefix(raw, "https://oauth.example/oauth2/authorize?") {
		t.Errorf("unexpected authorize URL: %s", raw)
	}
}

func TestPKCEChallengeMatchesVerifier(t *testing.T) {
	verifier, challenge, err := GeneratePKCE()
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256([]byte(verifier))
	want := base64.RawURLEncoding.EncodeToString(sum[:])
	if challenge != want {
		t.Errorf("challenge does not match S256(verifier)")
	}
}

func TestExchangeCode(t *testing.T) {
	var gotForm url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotForm = r.PostForm
		_, _ = w.Write([]byte(`{"access_token":"at","token_type":"Bearer","expires_in":600,"refresh_token":"rt"}`))
	}))
	defer srv.Close()

	cfg := OAuthConfig{ClientID: "cid", ClientSecret: "sek", RedirectURI: "https://app/cb", BaseURL: srv.URL}
	tok, err := ExchangeCode(context.Background(), srv.Client(), cfg, "the-code", "the-verifier")
	if err != nil {
		t.Fatalf("exchange: %v", err)
	}
	if tok.AccessToken != "at" || tok.RefreshToken != "rt" || tok.ExpiresIn != 600 {
		t.Errorf("unexpected token: %+v", tok)
	}
	if gotForm.Get("grant_type") != "authorization_code" {
		t.Errorf("grant_type: %q", gotForm.Get("grant_type"))
	}
	if gotForm.Get("code") != "the-code" || gotForm.Get("code_verifier") != "the-verifier" {
		t.Errorf("code/verifier not sent: %v", gotForm)
	}
	// client_secret must be masked, not sent in the clear.
	if gotForm.Get("client_secret") == "sek" || gotForm.Get("client_secret") == "" {
		t.Errorf("client_secret should be masked, got %q", gotForm.Get("client_secret"))
	}
}

func TestRefreshAccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		if r.PostForm.Get("grant_type") != "refresh_token" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"at2","token_type":"Bearer","expires_in":600}`))
	}))
	defer srv.Close()

	cfg := OAuthConfig{ClientID: "cid", BaseURL: srv.URL}
	tok, err := RefreshAccess(context.Background(), srv.Client(), cfg, "the-refresh")
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if tok.AccessToken != "at2" {
		t.Errorf("unexpected token: %+v", tok)
	}
}
