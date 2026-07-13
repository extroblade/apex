package auth

import "context"

// ctxKey is an unexported type so no other package can collide with our context
// key. This is the standard Go idiom for request-scoped context values.
type ctxKey int

const userKey ctxKey = iota

// WithUser returns a copy of ctx carrying the authenticated user.
func WithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, userKey, u)
}

// UserFromContext returns the authenticated user, if any. The bool reports
// whether a user was present (like a map lookup).
func UserFromContext(ctx context.Context) (User, bool) {
	u, ok := ctx.Value(userKey).(User)
	return u, ok
}
