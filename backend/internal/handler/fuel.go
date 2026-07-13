package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"apex/internal/fuel"
)

// FuelPlan handles POST /api/fuel/plan: decode JSON body, compute a plan,
// return it (or a 422 with the validation message).
func (h *Handler) FuelPlan(w http.ResponseWriter, r *http.Request) {
	var req fuel.Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid JSON body"})
		return
	}

	plan, err := fuel.Calculate(req)
	if err != nil {
		// Our known validation errors map to 422; anything else is a 500.
		status := http.StatusInternalServerError
		if errors.Is(err, fuel.ErrFuelPerLap) ||
			errors.Is(err, fuel.ErrTankCapacity) ||
			errors.Is(err, fuel.ErrRaceLength) ||
			errors.Is(err, fuel.ErrLapTime) ||
			errors.Is(err, fuel.ErrWindowLapTime) ||
			errors.Is(err, fuel.ErrRulesInfeasible) ||
			errors.Is(err, fuel.ErrTooManyStops) {
			status = http.StatusUnprocessableEntity
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, plan)
}
