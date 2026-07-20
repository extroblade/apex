package handler

import (
	"encoding/json"
	"errors"
	"io"
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
// screen.
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

// BillingCheckout creates a provider checkout session and returns its URL.
func (h *Handler) BillingCheckout(w http.ResponseWriter, r *http.Request) {
	if h.Subscription == nil {
		writeJSON(w, http.StatusInternalServerError, errBody("subscription service unavailable"))
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeJSON(w, http.StatusBadRequest, errBody("invalid JSON body"))
		return
	}
	if req.Plan == "" {
		req.Plan = subscription.PlanPro
	}
	url, err := h.Subscription.CheckoutURL(r.Context(), user.ID, user.Email, req.Plan)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrStripeNotConfigured):
			writeJSON(w, http.StatusServiceUnavailable, errBody(err.Error()))
		case errors.Is(err, subscription.ErrUnsupportedPlan):
			writeJSON(w, http.StatusUnprocessableEntity, errBody(err.Error()))
		default:
			writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// BillingPortal creates a customer-portal session URL for the caller.
func (h *Handler) BillingPortal(w http.ResponseWriter, r *http.Request) {
	if h.Subscription == nil {
		writeJSON(w, http.StatusInternalServerError, errBody("subscription service unavailable"))
		return
	}
	user, ok := auth.UserFromContext(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, errBody("unauthorized"))
		return
	}
	url, err := h.Subscription.PortalURL(r.Context(), user.ID)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrStripeNotConfigured):
			writeJSON(w, http.StatusServiceUnavailable, errBody(err.Error()))
		case errors.Is(err, subscription.ErrNoBillingCustomer):
			writeJSON(w, http.StatusUnprocessableEntity, errBody(err.Error()))
		default:
			writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// StripeWebhook ingests Stripe events and updates subscriptions/entitlements.
func (h *Handler) StripeWebhook(w http.ResponseWriter, r *http.Request) {
	if h.Subscription == nil {
		writeJSON(w, http.StatusServiceUnavailable, errBody("subscription service unavailable"))
		return
	}
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("failed to read body"))
		return
	}
	if err := h.Subscription.HandleStripeWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
		switch {
		case errors.Is(err, subscription.ErrStripeWebhookNotConfigured):
			writeJSON(w, http.StatusNotFound, errBody("not found"))
		case errors.Is(err, subscription.ErrInvalidStripeSignature):
			writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		default:
			writeJSON(w, http.StatusInternalServerError, errBody(err.Error()))
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DevSetTier lets a developer (dev cookie gate) switch their own plan tier.
// This remains as a local/test bridge even after Stripe wiring.
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
