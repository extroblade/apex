-- Billing foundation for Variant A (freemium + Pro subscription).
--
-- plan_tier is the fast entitlement snapshot used in request-time checks.
-- "free" is the safe default for every existing user.
ALTER TABLE users
    ADD COLUMN plan_tier VARCHAR(16) NOT NULL DEFAULT 'free';

-- Provider-facing subscription records. We keep history (multiple rows per user)
-- so webhook replays, renewals, and cancellations are auditable over time.
CREATE TABLE IF NOT EXISTS billing_subscriptions (
    id                       BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    user_id                  BIGINT UNSIGNED NOT NULL,
    provider                 VARCHAR(32)     NOT NULL,
    provider_customer_id     VARCHAR(191)    NULL,
    provider_subscription_id VARCHAR(191)    NULL,
    plan_tier                VARCHAR(16)     NOT NULL,
    status                   VARCHAR(32)     NOT NULL,
    current_period_start     TIMESTAMP       NULL,
    current_period_end       TIMESTAMP       NULL,
    cancel_at_period_end     TINYINT(1)      NOT NULL DEFAULT 0,
    canceled_at              TIMESTAMP       NULL,
    ended_at                 TIMESTAMP       NULL,
    metadata_json            JSON            NULL,
    created_at               TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at               TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_billing_subscriptions_user (user_id),
    KEY idx_billing_subscriptions_status (status),
    UNIQUE KEY uq_billing_provider_subscription (provider, provider_subscription_id),
    CONSTRAINT fk_billing_subscriptions_user FOREIGN KEY (user_id)
        REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Raw billing events (e.g. Stripe webhooks), stored idempotently by provider+id.
CREATE TABLE IF NOT EXISTS billing_events (
    id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    provider    VARCHAR(32)     NOT NULL,
    event_id    VARCHAR(191)    NOT NULL,
    event_type  VARCHAR(64)     NOT NULL,
    payload     JSON            NOT NULL,
    received_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP      NULL,
    error_text  VARCHAR(255)    NOT NULL DEFAULT '',
    PRIMARY KEY (id),
    UNIQUE KEY uq_billing_events_provider_id (provider, event_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
