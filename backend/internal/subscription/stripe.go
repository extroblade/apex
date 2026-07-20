package subscription

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	gomysql "github.com/go-sql-driver/mysql"
	"github.com/stripe/stripe-go/v83"
	billingportalsession "github.com/stripe/stripe-go/v83/billingportal/session"
	checkoutsession "github.com/stripe/stripe-go/v83/checkout/session"
	"github.com/stripe/stripe-go/v83/webhook"
)

const stripeProvider = "stripe"

var (
	ErrStripeNotConfigured        = errors.New("stripe is not configured")
	ErrStripeWebhookNotConfigured = errors.New("stripe webhook is not configured")
	ErrInvalidStripeSignature     = errors.New("invalid stripe signature")
	ErrUnsupportedPlan            = errors.New("unsupported billing plan")
	ErrNoBillingCustomer          = errors.New("no billing customer found")
)

// StripeConfig holds all Stripe-specific runtime settings.
type StripeConfig struct {
	SecretKey     string
	WebhookSecret string
	ProPriceID    string
	SuccessURL    string
	CancelURL     string
	PortalReturn  string
}

func (c StripeConfig) enabled() bool {
	return strings.TrimSpace(c.SecretKey) != "" && strings.TrimSpace(c.ProPriceID) != ""
}

// CheckoutURL creates a Stripe Checkout session URL for the given plan.
func (s *Service) CheckoutURL(ctx context.Context, userID int64, email, plan string) (string, error) {
	if !s.stripe.enabled() {
		return "", ErrStripeNotConfigured
	}
	if strings.ToLower(strings.TrimSpace(plan)) != PlanPro {
		return "", ErrUnsupportedPlan
	}
	customerID, err := s.stripeCustomerID(ctx, userID)
	if err != nil {
		return "", err
	}

	stripe.Key = s.stripe.SecretKey
	userRef := strconv.FormatInt(userID, 10)
	params := &stripe.CheckoutSessionParams{
		SuccessURL:        stripe.String(s.stripe.SuccessURL),
		CancelURL:         stripe.String(s.stripe.CancelURL),
		Mode:              stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		ClientReferenceID: stripe.String(userRef),
		Metadata: map[string]string{
			"apex_user_id": userRef,
		},
		SubscriptionData: &stripe.CheckoutSessionSubscriptionDataParams{
			Metadata: map[string]string{
				"apex_user_id": userRef,
			},
		},
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{Price: stripe.String(s.stripe.ProPriceID), Quantity: stripe.Int64(1)},
		},
	}
	if customerID != "" {
		params.Customer = stripe.String(customerID)
	} else if strings.TrimSpace(email) != "" {
		params.CustomerEmail = stripe.String(email)
	}

	cs, err := checkoutsession.New(params)
	if err != nil {
		return "", err
	}
	return cs.URL, nil
}

// PortalURL creates a Stripe Billing Portal session URL for the user.
func (s *Service) PortalURL(ctx context.Context, userID int64) (string, error) {
	if !s.stripe.enabled() {
		return "", ErrStripeNotConfigured
	}
	customerID, err := s.stripeCustomerID(ctx, userID)
	if err != nil {
		return "", err
	}
	if customerID == "" {
		return "", ErrNoBillingCustomer
	}

	stripe.Key = s.stripe.SecretKey
	params := &stripe.BillingPortalSessionParams{
		Customer:  stripe.String(customerID),
		ReturnURL: stripe.String(s.stripe.PortalReturn),
	}
	ps, err := billingportalsession.New(params)
	if err != nil {
		return "", err
	}
	return ps.URL, nil
}

// HandleStripeWebhook verifies a Stripe webhook signature and applies its state
// changes to local billing tables + user entitlements.
func (s *Service) HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error {
	if strings.TrimSpace(s.stripe.WebhookSecret) == "" {
		return ErrStripeWebhookNotConfigured
	}
	event, err := webhook.ConstructEvent(payload, signature, s.stripe.WebhookSecret)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidStripeSignature, err)
	}

	inserted, err := s.insertBillingEvent(ctx, event, payload)
	if err != nil || !inserted {
		return err
	}

	if err := s.applyStripeEvent(ctx, event); err != nil {
		_ = s.markBillingEventFailure(ctx, event.ID, err)
		return err
	}
	return s.markBillingEventProcessed(ctx, event.ID)
}

func (s *Service) applyStripeEvent(ctx context.Context, event stripe.Event) error {
	switch string(event.Type) {
	case "checkout.session.completed":
		var session stripeCheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			return err
		}
		return s.applyCheckoutSession(ctx, session)
	case "customer.subscription.created", "customer.subscription.updated", "customer.subscription.deleted":
		var sub stripeSubscriptionObject
		if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
			return err
		}
		return s.applyStripeSubscription(ctx, sub)
	default:
		return nil
	}
}

func (s *Service) applyCheckoutSession(ctx context.Context, session stripeCheckoutSession) error {
	userID, err := s.resolveStripeUserID(ctx, session.Metadata, session.Subscription, session.Customer, session.ClientReferenceID)
	if err != nil {
		return err
	}
	if session.Subscription == "" {
		return nil
	}
	now := time.Now().UTC()
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO billing_subscriptions (
			user_id, provider, provider_customer_id, provider_subscription_id, plan_tier, status,
			current_period_start, current_period_end, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			provider_customer_id = VALUES(provider_customer_id),
			plan_tier = VALUES(plan_tier),
			status = VALUES(status),
			current_period_start = VALUES(current_period_start),
			current_period_end = VALUES(current_period_end),
			metadata_json = VALUES(metadata_json)`,
		userID, stripeProvider, nullIfEmpty(session.Customer), session.Subscription, PlanPro, "checkout_completed",
		now, now.Add(30*24*time.Hour), metadataJSON(session.Metadata))
	return err
}

func (s *Service) applyStripeSubscription(ctx context.Context, sub stripeSubscriptionObject) error {
	userID, err := s.resolveStripeUserID(ctx, sub.Metadata, sub.ID, sub.Customer, "")
	if err != nil {
		return err
	}
	status := strings.ToLower(strings.TrimSpace(sub.Status))
	periodStart := unixToNullTime(sub.CurrentPeriodStart)
	periodEnd := unixToNullTime(sub.CurrentPeriodEnd)
	canceledAt := unixToNullTime(sub.CanceledAt)
	endedAt := unixToNullTime(sub.EndedAt)
	if isTerminalSubscriptionStatus(status) && !endedAt.Valid {
		endedAt = sql.NullTime{Time: time.Now().UTC(), Valid: true}
	}
	if !isTerminalSubscriptionStatus(status) {
		endedAt = sql.NullTime{}
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO billing_subscriptions (
			user_id, provider, provider_customer_id, provider_subscription_id, plan_tier, status,
			current_period_start, current_period_end, cancel_at_period_end, canceled_at, ended_at, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			provider_customer_id = VALUES(provider_customer_id),
			plan_tier = VALUES(plan_tier),
			status = VALUES(status),
			current_period_start = VALUES(current_period_start),
			current_period_end = VALUES(current_period_end),
			cancel_at_period_end = VALUES(cancel_at_period_end),
			canceled_at = VALUES(canceled_at),
			ended_at = VALUES(ended_at),
			metadata_json = VALUES(metadata_json)`,
		userID, stripeProvider, nullIfEmpty(sub.Customer), sub.ID, PlanPro, status,
		periodStart, periodEnd, sub.CancelAtPeriodEnd, canceledAt, endedAt, metadataJSON(sub.Metadata))
	if err != nil {
		return err
	}
	return s.syncTierFromSubscriptions(ctx, userID)
}

func (s *Service) syncTierFromSubscriptions(ctx context.Context, userID int64) error {
	var activePro int
	if err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM billing_subscriptions
		WHERE user_id = ? AND provider = ? AND plan_tier = ? AND ended_at IS NULL
		  AND status IN ('active', 'trialing', 'past_due')`,
		userID, stripeProvider, PlanPro).Scan(&activePro); err != nil {
		return err
	}
	next := PlanFree
	if activePro > 0 {
		next = PlanPro
	}
	_, err := s.db.ExecContext(ctx, `UPDATE users SET plan_tier = ? WHERE id = ?`, next, userID)
	return err
}

func (s *Service) stripeCustomerID(ctx context.Context, userID int64) (string, error) {
	var customer sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT provider_customer_id
		FROM billing_subscriptions
		WHERE user_id = ? AND provider = ? AND provider_customer_id IS NOT NULL AND provider_customer_id <> ''
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`, userID, stripeProvider).Scan(&customer)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return customer.String, nil
}

func (s *Service) resolveStripeUserID(ctx context.Context, metadata map[string]string, subscriptionID, customerID, fallbackRef string) (int64, error) {
	if v := strings.TrimSpace(metadata["apex_user_id"]); v != "" {
		id, err := strconv.ParseInt(v, 10, 64)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	if strings.TrimSpace(fallbackRef) != "" {
		id, err := strconv.ParseInt(strings.TrimSpace(fallbackRef), 10, 64)
		if err == nil && id > 0 {
			return id, nil
		}
	}
	var userID int64
	if strings.TrimSpace(subscriptionID) != "" {
		err := s.db.QueryRowContext(ctx, `
			SELECT user_id FROM billing_subscriptions
			WHERE provider = ? AND provider_subscription_id = ?
			ORDER BY updated_at DESC, id DESC
			LIMIT 1`, stripeProvider, subscriptionID).Scan(&userID)
		if err == nil {
			return userID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	if strings.TrimSpace(customerID) != "" {
		err := s.db.QueryRowContext(ctx, `
			SELECT user_id FROM billing_subscriptions
			WHERE provider = ? AND provider_customer_id = ?
			ORDER BY updated_at DESC, id DESC
			LIMIT 1`, stripeProvider, customerID).Scan(&userID)
		if err == nil {
			return userID, nil
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return 0, err
		}
	}
	return 0, errors.New("unable to resolve user for stripe event")
}

func (s *Service) insertBillingEvent(ctx context.Context, event stripe.Event, payload []byte) (bool, error) {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO billing_events (provider, event_id, event_type, payload)
		VALUES (?, ?, ?, ?)`,
		stripeProvider, event.ID, string(event.Type), string(payload))
	if err == nil {
		return true, nil
	}
	var myErr *gomysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == 1062 {
		return false, nil
	}
	return false, err
}

func (s *Service) markBillingEventProcessed(ctx context.Context, eventID string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE billing_events
		SET processed_at = NOW(), error_text = ''
		WHERE provider = ? AND event_id = ?`,
		stripeProvider, eventID)
	return err
}

func (s *Service) markBillingEventFailure(ctx context.Context, eventID string, failure error) error {
	msg := strings.TrimSpace(failure.Error())
	if len(msg) > 255 {
		msg = msg[:255]
	}
	_, err := s.db.ExecContext(ctx, `
		UPDATE billing_events
		SET error_text = ?
		WHERE provider = ? AND event_id = ?`,
		msg, stripeProvider, eventID)
	return err
}

func isTerminalSubscriptionStatus(status string) bool {
	switch status {
	case "canceled", "unpaid", "incomplete_expired":
		return true
	default:
		return false
	}
}

func unixToNullTime(v int64) sql.NullTime {
	if v <= 0 {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: time.Unix(v, 0).UTC(), Valid: true}
}

func nullIfEmpty(v string) any {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	return v
}

func metadataJSON(m map[string]string) []byte {
	if len(m) == 0 {
		return []byte("{}")
	}
	b, err := json.Marshal(m)
	if err != nil {
		return []byte(fmt.Sprintf(`{"marshalError":%q}`, err.Error()))
	}
	return b
}

type stripeCheckoutSession struct {
	ID                string            `json:"id"`
	Customer          string            `json:"customer"`
	Subscription      string            `json:"subscription"`
	ClientReferenceID string            `json:"client_reference_id"`
	Metadata          map[string]string `json:"metadata"`
}

type stripeSubscriptionObject struct {
	ID                 string            `json:"id"`
	Customer           string            `json:"customer"`
	Status             string            `json:"status"`
	CancelAtPeriodEnd  bool              `json:"cancel_at_period_end"`
	CanceledAt         int64             `json:"canceled_at"`
	EndedAt            int64             `json:"ended_at"`
	CurrentPeriodStart int64             `json:"current_period_start"`
	CurrentPeriodEnd   int64             `json:"current_period_end"`
	Metadata           map[string]string `json:"metadata"`
}
