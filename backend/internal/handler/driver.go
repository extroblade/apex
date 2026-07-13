package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// SearchDrivers handles GET /api/drivers/search?q=... — search for any driver
// using the logged-in user's own iRacing session.
func (h *Handler) SearchDrivers(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	term := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(term) < 2 {
		writeJSON(w, http.StatusUnprocessableEntity, errBody("search term must be at least 2 characters"))
		return
	}

	results, err := h.Racing.SearchDrivers(r.Context(), user.ID, term)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, results)
}

// DriverProfile handles GET /api/drivers/{custId} — any driver's stats, fetched
// through the logged-in user's session and cached.
func (h *Handler) DriverProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	custID, err := strconv.Atoi(chi.URLParam(r, "custId"))
	if err != nil || custID <= 0 {
		writeJSON(w, http.StatusBadRequest, errBody("invalid cust id"))
		return
	}

	profile, err := h.Racing.LookupDriver(r.Context(), user.ID, custID)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, profile)
}
