package middleware

import (
	"net/http"
)

// MaxBody wraps the request body so a single request can't stream an unbounded
// amount of data into the server (and into json.Decode). On overflow, the
// wrapped Read returns an error *and* the next handler's json.Decode will see
// it — so the handler returns its normal "invalid JSON body" 400 rather than a
// 500 from an OOM. Apply once to the whole API router with a sane default; bump
// it for specific routes (e.g. avatar upload) that legitimately need more.
func MaxBody(max int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't cap GET/DELETE — they have no body to read, and capping
			// would still let the wrapper short-circuit fine.
			if r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, max)
			}
			next.ServeHTTP(w, r)
		})
	}
}
