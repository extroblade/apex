package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"apex/internal/auth"
)

// fakeAuth implements the Authenticator interface, so we can test the
// middleware without a database.
type fakeAuth struct {
	user auth.User
	err  error
}

func (f fakeAuth) Authenticate(_ context.Context, _ string) (auth.User, error) {
	return f.user, f.err
}

func TestRequireAuth_BlocksAnonymous(t *testing.T) {
	h := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 for anonymous request, got %d", rec.Code)
	}
}

func TestAuth_SetsUserFromCookie(t *testing.T) {
	fake := fakeAuth{user: auth.User{ID: 7, Email: "x@y.com"}}

	var gotID int64
	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := auth.UserFromContext(r.Context()); ok {
			gotID = u.ID
		}
		w.WriteHeader(http.StatusOK)
	})

	// Chain the two middlewares like the real router does.
	h := Auth(fake)(RequireAuth(final))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookie, Value: "any-token"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("want 200 for authenticated request, got %d", rec.Code)
	}
	if gotID != 7 {
		t.Errorf("want user id 7 in context, got %d", gotID)
	}
}
