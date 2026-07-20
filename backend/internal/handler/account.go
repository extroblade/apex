package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"apex/internal/auth"
)

// DeleteAccount deletes the authenticated user's account after re-checking
// their current password. The session cookie is cleared. On success the
// caller is fully logged out and all their data is gone (cascade).
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var req struct {
		CurrentPassword string `json:"currentPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if err := h.Auth.DeleteAccount(r.Context(), user.ID, req.CurrentPassword); err != nil {
		writeJSON(w, authStatus(err), errBody(err.Error()))
		return
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// ExportData returns the authenticated user's full data as a JSON download
// (GDPR data portability). The Content-Disposition header makes browsers save
// it as a file rather than render it inline.
func (h *Handler) ExportData(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	data, err := h.Auth.MarshalExport(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	filename := "contentpilot-account-export-" + time.Now().UTC().Format("2006-01-02") + ".json"
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// RequestEmailChange starts the email-change flow: verifies the current
// password, stages the new email as pending, and emails a verification link
// to the new address. Returns 204 even when the mailer is disabled (dev/test)
// — the token is still issued so a later resend works.
func (h *Handler) RequestEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())

	var req struct {
		NewEmail        string `json:"newEmail"`
		CurrentPassword string `json:"currentPassword"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if err := h.Auth.RequestEmailChange(r.Context(), user.ID, req.NewEmail, req.CurrentPassword); err != nil {
		writeJSON(w, emailChangeStatus(err), errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CancelEmailChange clears a staged pending email and discards the outstanding
// verify token. Lets a user undo a mistaken email-change request from the
// profile page without waiting for the token to expire.
func (h *Handler) CancelEmailChange(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	if err := h.Auth.CancelEmailChange(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// emailChangeStatus maps email-change errors to HTTP status codes. Reuses
// authStatus for the shared credential/validation cases.
func emailChangeStatus(err error) int {
	switch {
	case errors.Is(err, auth.ErrEmailSame):
		return http.StatusUnprocessableEntity
	case errors.Is(err, auth.ErrEmailTaken):
		return http.StatusConflict
	default:
		return authStatus(err)
	}
}
