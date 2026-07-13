package iracing

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer simulates the Data API: it requires a Bearer token, and data
// endpoints return an envelope pointing at an /s3 payload on the same server.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/data/member/get", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_, _ = w.Write([]byte(`{"link":"http://` + r.Host + `/s3/members"}`))
	})

	mux.HandleFunc("/s3/members", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"members":[{"cust_id":42,"display_name":"Ada Speed",` +
			`"licenses":[{"category_id":2,"category":"road","irating":2500,"safety_rating":3.45}]}]}`))
	})

	return httptest.NewServer(mux)
}

func TestMembersFollowsLinkWithBearer(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient().WithBaseURL(srv.URL)
	c.SetToken("test-token")

	m, err := c.Member(context.Background(), 42)
	if err != nil {
		t.Fatalf("member: %v", err)
	}
	if m.CustID != 42 || m.DisplayName != "Ada Speed" {
		t.Errorf("unexpected member: %+v", m)
	}
	if len(m.Licenses) != 1 || m.Licenses[0].IRating != 2500 {
		t.Errorf("unexpected licenses: %+v", m.Licenses)
	}
}

func TestMissingTokenIsUnauthorized(t *testing.T) {
	srv := newTestServer(t)
	defer srv.Close()

	c := NewClient().WithBaseURL(srv.URL) // no token set
	if _, err := c.Member(context.Background(), 42); !errors.Is(err, ErrAuth) {
		t.Fatalf("want ErrAuth without a token, got %v", err)
	}
}

func TestNonJSONResponseDetected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("<!DOCTYPE html><html><body>Just a moment...</body></html>"))
	}))
	defer srv.Close()

	c := NewClient().WithBaseURL(srv.URL)
	c.SetToken("test-token")
	if _, err := c.Member(context.Background(), 42); !errors.Is(err, ErrUnexpectedResponse) {
		t.Fatalf("want ErrUnexpectedResponse for an HTML body, got %v", err)
	}
}

func TestStatusErrorMapping(t *testing.T) {
	cases := map[int]error{
		http.StatusOK:                 nil,
		http.StatusUnauthorized:       ErrAuth,
		http.StatusForbidden:          ErrAuth,
		http.StatusNotFound:           ErrNotFound,
		http.StatusTooManyRequests:    ErrRateLimited,
		http.StatusServiceUnavailable: ErrMaintenance,
	}
	for code, want := range cases {
		if got := statusError(code); !errors.Is(got, want) {
			t.Errorf("status %d: want %v, got %v", code, want, got)
		}
	}
}
