package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"apex/internal/auth"
	"apex/internal/cache"
	"apex/internal/features"
	"apex/internal/goals"
	"apex/internal/locales"
	"apex/internal/racing"
	"apex/internal/setups"
)

// Handler carries dependencies shared by the HTTP handlers.
type Handler struct {
	DB           *sql.DB
	Auth         *auth.Service
	Racing       *racing.Service // always set; OAuth features gate internally
	Features     *features.Service
	Setups       *setups.Service
	Goals        *goals.Service
	Locales      *locales.Service
	Cache        *cache.Cache // optional; nil-safe (fail-open)
	CookieSecure bool
	DeveloperKey string
}

func New(db *sql.DB, authSvc *auth.Service) *Handler {
	return &Handler{DB: db, Auth: authSvc}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
