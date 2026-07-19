package middleware

import (
	"net/http"
	"strings"
)

// CORS is a minimal, allowlist-based CORS middleware for the JSON API. `origins`
// is a comma-separated allowlist (e.g. "https://app.example.com,http://localhost:3000").
//
// It reflects the request Origin ONLY when it's on the allowlist, and then sets
// Allow-Credentials (the SPA authenticates with a cookie, so a wildcard origin
// is both unsafe and — with credentials — rejected by browsers). An empty
// allowlist sends no CORS headers at all: same-origin requests (the normal
// setup, where nginx proxies /api) don't need them, and cross-origin ones are
// denied. A literal "*" allows any origin but, per spec, WITHOUT credentials.
func CORS(origins string) func(http.Handler) http.Handler {
	allow := parseOrigins(origins)
	wildcard := len(allow) == 1 && allow[0] == "*"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				h := w.Header()
				h.Add("Vary", "Origin")
				switch {
				case wildcard:
					h.Set("Access-Control-Allow-Origin", "*")
				case allowed(allow, origin):
					h.Set("Access-Control-Allow-Origin", origin)
					h.Set("Access-Control-Allow-Credentials", "true")
				}
				if _, ok := h["Access-Control-Allow-Origin"]; ok {
					h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
					h.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
					h.Set("Access-Control-Max-Age", "600")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func parseOrigins(s string) []string {
	out := make([]string, 0, 4)
	for _, o := range strings.Split(s, ",") {
		if o = strings.TrimSpace(o); o != "" {
			out = append(out, o)
		}
	}
	return out
}

func allowed(list []string, origin string) bool {
	for _, o := range list {
		if strings.EqualFold(o, origin) {
			return true
		}
	}
	return false
}
