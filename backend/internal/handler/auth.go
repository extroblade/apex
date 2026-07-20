package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"apex/internal/auth"
	"apex/internal/middleware"
)

// credentials is the shared request body for register and login.
type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Register creates an account and immediately logs the user in.
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req credentials
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}

	if _, err := h.Auth.Register(r.Context(), req.Email, req.Password); err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}

	// Auto-login so the client gets a session cookie without a second request.
	token, user, err := h.Auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, user)
}

// Login verifies credentials and sets the session cookie.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req credentials
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}

	token, user, err := h.Auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}
	h.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, user)
}

// Logout deletes the session and clears the cookie.
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.SessionCookie); err == nil {
		_ = h.Auth.Logout(r.Context(), c.Value)
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// Me returns the currently authenticated user (populated by the Auth middleware).
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}
	writeJSON(w, http.StatusOK, u)
}

// UpdateProfile changes the authenticated user's nickname and email.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var req struct {
		Nickname string `json:"nickname"`
		Email    string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	updated, err := h.Auth.UpdateProfile(r.Context(), user.ID, req.Nickname, req.Email)
	if err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// UpdateAvatar sets or clears the authenticated user's avatar (a data URL).
func (h *Handler) UpdateAvatar(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var req struct {
		Avatar string `json:"avatar"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	updated, err := h.Auth.UpdateAvatar(r.Context(), user.ID, req.Avatar)
	if err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

// ChangePassword updates the password after verifying the current one.
func (h *Handler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	// The current session token — kept alive while all others are revoked.
	var keepToken string
	if c, err := r.Cookie(middleware.SessionCookie); err == nil {
		keepToken = c.Value
	}
	if err := h.Auth.ChangePassword(r.Context(), user.ID, req.CurrentPassword, req.NewPassword, keepToken); err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RequestPasswordReset starts the reset flow: given an email, it issues a
// reset token and emails a link. Always 204 — never reveals whether the email
// belongs to an account (no enumeration).
func (h *Handler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	_ = h.Auth.RequestPasswordReset(r.Context(), req.Email)
	w.WriteHeader(http.StatusNoContent)
}

// ConfirmPasswordReset consumes the reset token and sets the new password.
func (h *Handler) ConfirmPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if err := h.Auth.ConfirmPasswordReset(r.Context(), req.Token, req.NewPassword); err != nil {
		writeJSON(w, recoveryStatus(err), errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RequestEmailVerification starts the verification flow for an email (used by
// the pre-login "resend" form). Always 204 — no enumeration.
func (h *Handler) RequestEmailVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	_ = h.Auth.RequestEmailVerification(r.Context(), req.Email)
	w.WriteHeader(http.StatusNoContent)
}

// ConfirmEmailVerification consumes the verify token (clicked from the email).
// Returns 204 on success so the SPA can show a "verified" state without parsing
// the body.
func (h *Handler) ConfirmEmailVerification(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		writeJSON(w, http.StatusBadRequest, errBody("missing token"))
		return
	}
	if err := h.Auth.ConfirmEmailVerification(r.Context(), token); err != nil {
		writeJSON(w, recoveryStatus(err), errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ResendEmailVerification re-issues and emails the verification token for the
// logged-in user. Used by the "verify your email" banner on the profile page.
func (h *Handler) ResendEmailVerification(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	_ = h.Auth.RequestEmailVerification(r.Context(), user.Email)
	w.WriteHeader(http.StatusNoContent)
}

// recoveryStatus maps the recovery/verification errors to HTTP status codes.
// ErrTokenInvalid → 400 (bad input, not a server fault); ErrWeakPassword → 422.
func recoveryStatus(err error) int {
	switch {
	case errors.Is(err, auth.ErrTokenInvalid):
		return http.StatusBadRequest
	case errors.Is(err, auth.ErrWeakPassword):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// authStatus maps auth errors to HTTP status codes.
func authStatus(err error) int {
	switch {
	case errors.Is(err, auth.ErrEmailTaken):
		return http.StatusConflict
	case errors.Is(err, auth.ErrInvalidCredentials), errors.Is(err, auth.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, auth.ErrWeakPassword), errors.Is(err, auth.ErrInvalidEmail),
		errors.Is(err, auth.ErrNicknameTooLong), errors.Is(err, auth.ErrInvalidAvatar),
		errors.Is(err, auth.ErrAvatarTooLarge):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true, // not readable from JS — mitigates XSS token theft
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		MaxAge:   int((30 * 24 * time.Hour).Seconds()),
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     middleware.SessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1, // delete immediately
	})
}
