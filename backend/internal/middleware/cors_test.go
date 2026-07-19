package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORS(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	cases := []struct {
		name      string
		allow     string
		origin    string
		wantACAO  string
		wantCreds bool
	}{
		{"allowlisted origin is reflected with credentials", "http://localhost:3000", "http://localhost:3000", "http://localhost:3000", true},
		{"origin not on the allowlist gets no CORS", "https://app.example.com", "http://evil.test", "", false},
		{"multi-origin allowlist matches the second", "https://a.test, https://b.test", "https://b.test", "https://b.test", true},
		{"wildcard allows any origin without credentials", "*", "http://anything.test", "*", false},
		{"empty allowlist sends nothing", "", "http://localhost:3000", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/x", nil)
			req.Header.Set("Origin", tc.origin)
			CORS(tc.allow)(next).ServeHTTP(rec, req)

			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.wantACAO {
				t.Errorf("Allow-Origin = %q, want %q", got, tc.wantACAO)
			}
			creds := rec.Header().Get("Access-Control-Allow-Credentials") == "true"
			if creds != tc.wantCreds {
				t.Errorf("Allow-Credentials = %v, want %v", creds, tc.wantCreds)
			}
		})
	}
}

func TestCORSPreflightShortCircuits(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true })
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/x", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	CORS("http://localhost:3000")(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
	if called {
		t.Error("preflight should not call the next handler")
	}
}
