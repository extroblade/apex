package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"apex/internal/features"
)

// devCookie is the cookie name the frontend sets via ?dev=KEY.
const devCookie = "developer"

// devKey is the developer key from the environment. Set via server wiring.
func (h *Handler) devKey() string { return h.DeveloperKey }

// devAuth checks that the developer cookie matches the configured key. Returns
// false (and the usual 404 fallback) when DEVELOPER_KEY is empty or the cookie
// doesn't match — so the endpoints silently disappear in production.
func (h *Handler) devAuth(w http.ResponseWriter, r *http.Request) bool {
	key := h.devKey()
	if key == "" {
		return false
	}
	c, err := r.Cookie(devCookie)
	if err != nil || c.Value != key {
		return false
	}
	return true
}

// AllFeatures returns every feature flag and its state. Only visible when the
// "cockpit" feature flag is on AND the developer cookie matches DEVELOPER_KEY.
func (h *Handler) AllFeatures(w http.ResponseWriter, r *http.Request) {
	if h.Features == nil || !h.devAuth(w, r) {
		writeJSON(w, http.StatusNotFound, errBody("not found"))
		return
	}
	flags, err := h.Features.All(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, flags)
}

// toggleFlag is the request body for PUT /api/features/{key}.
type toggleFlag struct {
	Enabled *bool `json:"enabled"`
}

// ToggleFeature flips a single feature flag. Only visible when the "cockpit"
// feature flag is on AND the developer cookie matches DEVELOPER_KEY. Invalidates
// the feature-flag cache so the new state propagates immediately.
func (h *Handler) ToggleFeature(w http.ResponseWriter, r *http.Request) {
	if h.Features == nil || !h.devAuth(w, r) {
		writeJSON(w, http.StatusNotFound, errBody("not found"))
		return
	}
	// Extract the flag key from the URL path. chi populates URLParams after
	// stripping the matched route prefix, so for /api/features/{key} the
	// "key" param is available via chi.URLParam.
	key := strings.TrimPrefix(r.URL.Path, "/api/features/")
	key = strings.TrimRight(key, "/")
	if key == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing flag key"))
		return
	}

	var req toggleFlag
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if req.Enabled == nil {
		writeJSON(w, http.StatusBadRequest, errBody("missing enabled field"))
		return
	}

	if err := h.Features.Set(r.Context(), key, *req.Enabled); err != nil {
		if errors.Is(err, features.ErrFlagNotFound) {
			writeJSON(w, http.StatusNotFound, errBody("unknown flag"))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	h.Features.Invalidate() // flush cache so /api/features reflects the change
	w.WriteHeader(http.StatusNoContent)
}

// HealthCockpit returns a summary for the cockpit overlay: DB ping + redis ping
// (nil-safe so it works before the Redis task is wired up).
func (h *Handler) HealthCockpit(w http.ResponseWriter, r *http.Request) {
	if !h.devAuth(w, r) {
		writeJSON(w, http.StatusNotFound, errBody("not found"))
		return
	}
	dbOK := true
	if err := h.DB.PingContext(r.Context()); err != nil {
		dbOK = false
	}
	// redisEnabled distinguishes "no cache configured" from "configured but down"
	// (redisOK). Nil-safe: h.Cache may be nil before Redis is wired up.
	redisEnabled := h.Cache != nil && h.Cache.Enabled()
	redisOK := redisEnabled && h.Cache.Ping(r.Context())
	writeJSON(w, http.StatusOK, map[string]any{
		"db":           dbOK,
		"redisEnabled": redisEnabled,
		"redis":        redisOK,
	})
}
