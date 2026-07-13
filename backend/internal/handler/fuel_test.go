package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// httptest lets us exercise a handler without a real server or network. The
// fuel handler doesn't use the database, so a nil-DB Handler is fine here.
func TestFuelPlanHandler(t *testing.T) {
	h := New(nil, nil)

	body := `{"raceType":"laps","raceLaps":30,"fuelPerLap":2.5,"tankCapacity":50,"extraLaps":1}`
	req := httptest.NewRequest(http.MethodPost, "/api/fuel/plan", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.FuelPlan(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: want 200, got %d (body: %s)", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"totalLaps":30`) {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestFuelPlanHandler_ValidationError(t *testing.T) {
	h := New(nil, nil)

	// fuelPerLap = 0 is invalid -> 422.
	body := `{"raceType":"laps","raceLaps":30,"fuelPerLap":0,"tankCapacity":50}`
	req := httptest.NewRequest(http.MethodPost, "/api/fuel/plan", strings.NewReader(body))
	rec := httptest.NewRecorder()

	h.FuelPlan(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status: want 422, got %d", rec.Code)
	}
}
