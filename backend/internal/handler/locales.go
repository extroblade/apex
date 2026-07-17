package handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"apex/internal/locales"
)

// ListLocales returns the available languages (code + display name). The list is
// backend-driven, so a new language appears here — and in the app — with no
// frontend deploy.
func (h *Handler) ListLocales(w http.ResponseWriter, r *http.Request) {
	list, err := h.Locales.List(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody("locales unavailable"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"locales": list})
}

// GetLocale returns one language's translation bundle as JSON. The frontend
// fetches this for any language other than the bundled `en`.
func (h *Handler) GetLocale(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	bundle, err := h.Locales.Bundle(r.Context(), code)
	if errors.Is(err, locales.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, errBody("unknown locale"))
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody("locale unavailable"))
		return
	}
	// The bundle is already a JSON document — write it straight through.
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(bundle))
}
