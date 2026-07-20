package subscription

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestHandleStripeWebhookRequiresSecret(t *testing.T) {
	svc := New(nil).WithStripe(StripeConfig{})
	err := svc.HandleStripeWebhook(context.Background(), []byte(`{}`), "")
	if !errors.Is(err, ErrStripeWebhookNotConfigured) {
		t.Fatalf("expected ErrStripeWebhookNotConfigured, got %v", err)
	}
}

func TestHandleStripeWebhookInvalidSignature(t *testing.T) {
	svc := New(nil).WithStripe(StripeConfig{WebhookSecret: "whsec_test"})
	err := svc.HandleStripeWebhook(context.Background(), []byte(`{}`), "invalid")
	if !errors.Is(err, ErrInvalidStripeSignature) {
		t.Fatalf("expected ErrInvalidStripeSignature, got %v", err)
	}
}

func TestIsEntitledStripeStatus(t *testing.T) {
	now := time.Date(2026, 7, 20, 10, 0, 0, 0, time.UTC)
	inGrace := sql.NullTime{Valid: true, Time: now.Add(-2 * time.Hour)}
	outOfGrace := sql.NullTime{Valid: true, Time: now.Add(-pastDueGracePeriod - time.Hour)}
	noPeriod := sql.NullTime{}

	tests := []struct {
		name      string
		status    string
		periodEnd sql.NullTime
		want      bool
	}{
		{name: "active", status: "active", periodEnd: noPeriod, want: true},
		{name: "trialing", status: "trialing", periodEnd: noPeriod, want: true},
		{name: "past_due_within_grace", status: "past_due", periodEnd: inGrace, want: true},
		{name: "past_due_outside_grace", status: "past_due", periodEnd: outOfGrace, want: false},
		{name: "past_due_without_period", status: "past_due", periodEnd: noPeriod, want: false},
		{name: "canceled", status: "canceled", periodEnd: noPeriod, want: false},
		{name: "unpaid", status: "unpaid", periodEnd: noPeriod, want: false},
		{name: "unknown", status: "paused", periodEnd: noPeriod, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isEntitledStripeStatus(tt.status, tt.periodEnd, now)
			if got != tt.want {
				t.Fatalf("isEntitledStripeStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}
