package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"apex/internal/auth"
	"apex/internal/goals"
)

// goalBody is the shared create/update payload.
type goalBody struct {
	Title   string  `json:"title"`
	Notes   string  `json:"notes"`
	Unit    string  `json:"unit"`
	Target  float64 `json:"target"`
	Current float64 `json:"current"`
	Done    *bool   `json:"done"`
	DueDate *string `json:"dueDate"`
}

func (b goalBody) input() goals.Input {
	return goals.Input{
		Title: b.Title, Notes: b.Notes, Unit: b.Unit,
		Target: b.Target, Current: b.Current, Done: b.Done, DueDate: b.DueDate,
	}
}

// ListGoals returns the caller's goals.
func (h *Handler) ListGoals(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	items, err := h.Goals.List(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// CreateGoal adds a goal for the caller.
func (h *Handler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	var body goalBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	g, err := h.Goals.Create(r.Context(), user.ID, body.input())
	if err != nil {
		writeJSON(w, goalStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusCreated, g)
}

// UpdateGoal replaces a goal the caller owns (also used for progress updates).
func (h *Handler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errBody("invalid id"))
		return
	}
	var body goalBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	g, err := h.Goals.Update(r.Context(), user.ID, id, body.input())
	if err != nil {
		writeJSON(w, goalStatus(err), errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, g)
}

// DeleteGoal removes a goal the caller owns.
func (h *Handler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	user, _ := auth.UserFromContext(r.Context())
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errBody("invalid id"))
		return
	}
	if err := h.Goals.Delete(r.Context(), user.ID, id); err != nil {
		writeJSON(w, goalStatus(err), errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func goalStatus(err error) int {
	switch {
	case errors.Is(err, goals.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, goals.ErrInvalid):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
