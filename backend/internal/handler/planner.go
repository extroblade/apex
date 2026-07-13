package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"apex/internal/racing"
)

// SyncCatalog refreshes the global iRacing catalog (cars/tracks/series).
func (h *Handler) SyncCatalog(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	counts, err := h.Racing.SyncCatalog(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, racingStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, counts)
}

func (h *Handler) ListCars(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	items, err := h.Racing.Cars(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ListTracks(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	items, err := h.Racing.Tracks(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ListSeries(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	items, err := h.Racing.SeriesList(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// SeasonView returns the season grid: every series' 13-week track calendar,
// annotated with the user's owned tracks/cars and favorites.
func (h *Handler) SeasonView(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	season, err := h.Racing.SeasonView(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, season)
}

// SetRacePlanned toggles a (series, week) race in the user's plan.
func (h *Handler) SetRacePlanned(w http.ResponseWriter, r *http.Request) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	var req struct {
		SeriesID int   `json:"seriesId"`
		Week     int   `json:"week"`
		Planned  *bool `json:"planned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if req.Planned == nil || req.SeriesID <= 0 || req.Week < 1 || req.Week > racing.SeasonWeeks {
		writeJSON(w, http.StatusUnprocessableEntity, errBody("seriesId, week (1-13) and planned are required"))
		return
	}
	if err := h.Racing.SetRacePlanned(r.Context(), user.ID, req.SeriesID, req.Week, *req.Planned); err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// flagBody is the toggle payload for ownership/favorite endpoints.
type flagBody struct {
	Owned    *bool `json:"owned"`
	Favorite *bool `json:"favorite"`
}

func (h *Handler) SetCarOwned(w http.ResponseWriter, r *http.Request) {
	h.toggle(w, r, func(s *racing.Service, uid int64, id int, on bool) error {
		return s.SetCarOwned(r.Context(), uid, id, on)
	}, flagOwned)
}

func (h *Handler) SetTrackOwned(w http.ResponseWriter, r *http.Request) {
	h.toggle(w, r, func(s *racing.Service, uid int64, id int, on bool) error {
		return s.SetTrackOwned(r.Context(), uid, id, on)
	}, flagOwned)
}

func (h *Handler) SetSeriesFavorite(w http.ResponseWriter, r *http.Request) {
	h.toggle(w, r, func(s *racing.Service, uid int64, id int, on bool) error {
		return s.SetSeriesFavorite(r.Context(), uid, id, on)
	}, flagFavorite)
}

type flagKind int

const (
	flagOwned flagKind = iota
	flagFavorite
)

// toggle shares the parse-id / decode-flag / call-service / respond logic.
func (h *Handler) toggle(
	w http.ResponseWriter,
	r *http.Request,
	set func(s *racing.Service, uid int64, id int, on bool) error,
	kind flagKind,
) {
	user, ok := h.racingReady(w, r)
	if !ok {
		return
	}
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid id"))
		return
	}
	var body flagBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	on := body.Owned
	if kind == flagFavorite {
		on = body.Favorite
	}
	if on == nil {
		writeJSON(w, http.StatusUnprocessableEntity, errBody("missing boolean flag"))
		return
	}
	if err := set(h.Racing, user.ID, id, *on); err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
