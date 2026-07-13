package handler

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"apex/internal/auth"
	"apex/internal/iracing"
	"apex/internal/racing"
)

// postLinkRedirect is where the browser lands after the OAuth callback. Relative
// so it resolves to the SPA origin (nginx/dev-proxy both serve the app there).
const postLinkRedirect = "/dashboard"

// racingReady guards handlers that need the iRacing integration; it returns the
// authenticated user and reports whether the request can proceed.
func (h *Handler) racingReady(w http.ResponseWriter, r *http.Request) (auth.User, bool) {
	if h.Racing == nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody("iRacing integration is not configured"))
		return auth.User{}, false
	}
	// RequireAuth runs before these handlers, so the user is guaranteed present.
	user, _ := auth.UserFromContext(r.Context())
	return user, true
}

// AuthorizeIRacing starts the OAuth flow, redirecting the browser to iRacing's
// login/consent page. Hit as a full-page navigation (carries the app session).
func (h *Handler) AuthorizeIRacing(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	authURL, err := h.Racing.BeginLink(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	http.Redirect(w, r, authURL, http.StatusFound)
}

// CallbackIRacing is iRacing's OAuth redirect target. It authenticates the user
// via the OAuth `state` (not the app session), so it isn't behind RequireAuth.
// It always redirects the browser back to the SPA, with a success or error flag.
func (h *Handler) CallbackIRacing(w http.ResponseWriter, r *http.Request) {
	if h.Racing == nil {
		redirectWithError(w, r, "iRacing integration is not configured")
		return
	}

	q := r.URL.Query()
	if oauthErr := q.Get("error"); oauthErr != "" {
		redirectWithError(w, r, oauthErr)
		return
	}
	code, state := q.Get("code"), q.Get("state")
	if code == "" || state == "" {
		redirectWithError(w, r, "missing code or state")
		return
	}

	if _, err := h.Racing.CompleteLink(r.Context(), state, code); err != nil {
		redirectWithError(w, r, err.Error())
		return
	}
	http.Redirect(w, r, postLinkRedirect+"?iracing=linked", http.StatusFound)
}

func redirectWithError(w http.ResponseWriter, r *http.Request, msg string) {
	http.Redirect(w, r, postLinkRedirect+"?iracing_error="+url.QueryEscape(msg), http.StatusFound)
}

// UnlinkIRacing removes a user's stored iRacing credentials.
func (h *Handler) UnlinkIRacing(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	if err := h.Racing.Unlink(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// IRacingStatus reports whether an account is linked.
func (h *Handler) IRacingStatus(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	account, err := h.Racing.Status(r.Context(), user.ID)
	if errors.Is(err, racing.ErrNotLinked) {
		writeJSON(w, http.StatusOK, map[string]any{"linked": false})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"linked": true, "account": account})
}

// IRacingStats returns the live driver dashboard.
func (h *Handler) IRacingStats(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	dash, err := h.Racing.Stats(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, dash)
}

// IRacingSync pulls recent races into the local store.
func (h *Handler) IRacingSync(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	n, err := h.Racing.Sync(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"synced": n})
}

// Comparators.

func (h *Handler) CompareCategories(w http.ResponseWriter, r *http.Request) {
	h.compare(w, r, (*racing.Service).CompareCategories)
}

func (h *Handler) CompareCars(w http.ResponseWriter, r *http.Request) {
	h.compare(w, r, (*racing.Service).CompareCars)
}

func (h *Handler) CompareTracks(w http.ResponseWriter, r *http.Request) {
	h.compare(w, r, (*racing.Service).CompareTracks)
}

// compare runs one of the comparator methods (passed as a method value) and
// writes the result. This avoids three near-identical handler bodies.
func (h *Handler) compare(
	w http.ResponseWriter,
	r *http.Request,
	fn func(*racing.Service, context.Context, int64) ([]racing.GroupStat, error),
) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	stats, err := fn(h.Racing, r.Context(), user.ID)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// racingStatus maps racing/iracing errors to HTTP status codes.
func racingStatus(err error) int {
	switch {
	case errors.Is(err, racing.ErrNotLinked):
		return http.StatusConflict
	case errors.Is(err, racing.ErrOAuthDisabled):
		return http.StatusServiceUnavailable
	case errors.Is(err, iracing.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, racing.ErrLinkFailed), errors.Is(err, iracing.ErrAuth):
		return http.StatusBadGateway
	case errors.Is(err, iracing.ErrRateLimited):
		return http.StatusTooManyRequests
	case errors.Is(err, iracing.ErrMaintenance), errors.Is(err, iracing.ErrUnexpectedResponse):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
