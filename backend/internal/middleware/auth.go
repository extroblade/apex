package middleware

import (
	"context"
	"net/http"

	"apex/internal/auth"
)

// SessionCookie is the cookie name holding the opaque session token.
const SessionCookie = "session"

// Authenticator is the small slice of *auth.Service this middleware needs.
// Depending on an interface (not the concrete type) keeps the packages loosely
// coupled and makes the middleware trivial to test with a fake.
type Authenticator interface {
	Authenticate(ctx context.Context, token string) (auth.User, error)
}

// Auth reads the session cookie and, when valid, attaches the user to the
// request context. It does NOT reject anonymous requests — that's RequireAuth's
// job — so it can wrap both public and protected routes.
func Auth(a Authenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if c, err := r.Cookie(SessionCookie); err == nil {
				if u, err := a.Authenticate(r.Context(), c.Value); err == nil {
					r = r.WithContext(auth.WithUser(r.Context(), u))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth rejects requests that have no authenticated user in context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := auth.UserFromContext(r.Context()); !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":"unauthorized"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
