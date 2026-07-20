package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"apex/internal/auth"
	"apex/internal/subscription"
)

type billingPlan struct {
	Key      string   `json:"key"`
	Name     string   `json:"name"`
	Price    string   `json:"price"`
	Interval string   `json:"interval"`
	Features []string `json:"features"`
}

// BillingPlans returns the productized Free/Pro packaging shown in the upgrade
// screen. Checkout links are wired later when Stripe integration lands.
func (h *Handler) BillingPlans(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"plans": []billingPlan{
			{
				Key:      subscription.PlanFree,
				Name:     "Free",
				Price:    "$0",
				Interval: "month",
				Features: []string{
					"Fuel planner",
					"Season planner",
					"Setups showroom",
					"Goal tracker",
					"Baseline setup generator",
				},
			},
			{
				Key:      subscription.PlanPro,
				Name:     "Pro",
				Price:    "$9",
				Interval: "month",
				Features: []string{
					"Setup pack generator (safe/pro × race/qual/endurance/rain)",
					"Priority access to upcoming planner expansions",
					"All Free features",
				},
			},
		},
	})
}

// MySubscription returns the authenticated user's current tier + billing status.
func (h *Handler) MySubscription(w http.ResponseWriter, r *http.Request) {
	if h.Subscription == nil {
		writeJSON(w, http.StatusInternalServerError, errBody("subscription service unavailable"))
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}
	out, err := h.Subscription.Current(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// DevSetTier lets a developer (dev cookie gate) switch their own plan tier.
// This is a temporary bridge until checkout/webhooks are integrated.
func (h *Handler) DevSetTier(w http.ResponseWriter, r *http.Request) {
	if h.Subscription == nil || !h.devAuth(w, r) {
		writeJSON(w, http.StatusNotFound, errBody("not found"))
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	var req struct {
		Tier   string `json:"tier"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if err := h.Subscription.SetTier(r.Context(), user.ID, req.Tier, req.Reason); err != nil {
		if errors.Is(err, subscription.ErrInvalidTier) {
			writeJSON(w, http.StatusUnprocessableEntity, errBody(err.Error()))
			return
		}
		writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
