// Package subscription provides entitlement checks and lightweight subscription
// reads for Variant A (freemium + Pro).
package subscription

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	PlanFree = "free"
	PlanPro  = "pro"
)

const (
	StatusActive   = "active"
	StatusCanceled = "canceled"
)

var ErrInvalidTier = errors.New("invalid plan tier")

// Service manages subscription reads/writes.
type Service struct {
	db     *sql.DB
	stripe StripeConfig
}

func New(db *sql.DB) *Service { return &Service{db: db} }

// WithStripe configures Stripe-backed checkout/portal/webhook handling.
func (s *Service) WithStripe(cfg StripeConfig) *Service {
	s.stripe = cfg
	return s
}

// Summary is the authenticated user's current billing view.
type Summary struct {
	Tier                string  `json:"tier"`
	Pro                 bool    `json:"pro"`
	Status              string  `json:"status,omitempty"`
	Provider            string  `json:"provider,omitempty"`
	CurrentPeriodEndISO *string `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd   bool    `json:"cancelAtPeriodEnd"`
}

// Current returns the caller's current entitlement and (if present) the latest
// open subscription record.
func (s *Service) Current(ctx context.Context, userID int64) (Summary, error) {
	tier, err := s.tier(ctx, userID)
	if err != nil {
		return Summary{}, err
	}
	out := Summary{Tier: tier, Pro: tier == PlanPro}

	var (
		provider  sql.NullString
		status    sql.NullString
		cancel    bool
		periodEnd sql.NullTime
	)
	err = s.db.QueryRowContext(ctx, `
		SELECT provider, status, cancel_at_period_end, current_period_end
		FROM billing_subscriptions
		WHERE user_id = ? AND ended_at IS NULL
		ORDER BY updated_at DESC, id DESC
		LIMIT 1`, userID).
		Scan(&provider, &status, &cancel, &periodEnd)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Summary{}, err
	}
	if err == nil {
		out.Provider = provider.String
		out.Status = status.String
		out.CancelAtPeriodEnd = cancel
		if periodEnd.Valid {
			v := periodEnd.Time.UTC().Format(time.RFC3339)
			out.CurrentPeriodEndISO = &v
		}
	}
	return out, nil
}

// HasPro reports whether the user has an active Pro entitlement snapshot.
func (s *Service) HasPro(ctx context.Context, userID int64) bool {
	tier, err := s.tier(ctx, userID)
	return err == nil && tier == PlanPro
}

// SetTier is currently used by dev-only tooling to simulate upgrades/downgrades
// before Stripe checkout/webhooks are connected.
func (s *Service) SetTier(ctx context.Context, userID int64, tier, reason string) error {
	tier = strings.ToLower(strings.TrimSpace(tier))
	if tier != PlanFree && tier != PlanPro {
		return ErrInvalidTier
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx,
		`UPDATE users SET plan_tier = ? WHERE id = ?`, tier, userID); err != nil {
		return err
	}

	now := time.Now().UTC()
	if tier == PlanPro {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO billing_subscriptions (
				user_id, provider, plan_tier, status, current_period_start, current_period_end, metadata_json
			) VALUES (?, 'manual', ?, ?, ?, ?, ?)`,
			userID, tier, StatusActive, now, now.Add(30*24*time.Hour), jsonKV(reason)); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE billing_subscriptions
			SET status = ?, cancel_at_period_end = 0, canceled_at = COALESCE(canceled_at, ?), ended_at = COALESCE(ended_at, ?)
			WHERE user_id = ? AND ended_at IS NULL`,
			StatusCanceled, now, now, userID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) tier(ctx context.Context, userID int64) (string, error) {
	var tier string
	if err := s.db.QueryRowContext(ctx,
		`SELECT plan_tier FROM users WHERE id = ?`, userID).Scan(&tier); err != nil {
		return "", err
	}
	tier = strings.ToLower(strings.TrimSpace(tier))
	if tier == "" {
		return PlanFree, nil
	}
	return tier, nil
}

func jsonKV(reason string) []byte {
	if strings.TrimSpace(reason) == "" {
		return []byte(`{"source":"developer"}`)
	}
	return []byte(fmt.Sprintf(`{"source":"developer","reason":%q}`, reason))
}
