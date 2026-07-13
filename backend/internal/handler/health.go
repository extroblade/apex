package handler

import (
	"net/http"
)

// Health reports service liveness and database connectivity.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	dbOK := true
	if err := h.DB.PingContext(r.Context()); err != nil {
		dbOK = false
	}

	status := http.StatusOK
	if !dbOK {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]any{
		"status": "ok",
		"db":     dbOK,
	})
}
