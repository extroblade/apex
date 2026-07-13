package handler

import "net/http"

// ListFeatures returns the current feature-flag states so the UI can adapt.
func (h *Handler) ListFeatures(w http.ResponseWriter, r *http.Request) {
	if h.Features == nil {
		writeJSON(w, http.StatusOK, map[string]bool{})
		return
	}
	flags, err := h.Features.All(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, flags)
}
