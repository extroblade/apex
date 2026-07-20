package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime configuration, sourced from environment variables.
type Config struct {
	Port       string
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	CORSOrigin string
	// CookieSecure marks the session cookie Secure (HTTPS-only). Keep false for
	// local http development; set true in production behind TLS.
	CookieSecure bool
	// EncryptionKey is a base64-encoded 32-byte key used to encrypt stored
	// iRacing refresh tokens at rest. Empty disables the iRacing integration.
	EncryptionKey string
	// iRacing OAuth client settings (register an app with iRacing to get these).
	// ClientID + RedirectURI are required to enable iRacing; ClientSecret is
	// optional (only some client types get one).
	IRacingClientID     string
	IRacingClientSecret string
	IRacingRedirectURI  string
	// DeveloperKey activates the Cockpit dev-overlay when non-empty. The frontend
	// sets a "developer" cookie matching this key; only matching requests see the
	// all-features list and the toggle endpoint. Empty disables Cockpit entirely.
	DeveloperKey string
	// RedisAddr is the host:port of the Redis cache (e.g. "redis:6379"). Empty
	// disables caching — the app reads straight from the DB (fail-open).
	RedisAddr string
	// AuthRateLimit is the max register/login attempts per IP per minute.
	// <= 0 disables the limiter (used in tests/e2e where all traffic is one IP).
	AuthRateLimit int
	// SMTP settings for transactional email (password reset, email
	// verification). Empty Host disables the mailer — the auth flows still
	// run but no mail is delivered (dev/test without an SMTP server).
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	// AppBaseURL is the public origin used to build links in transactional
	// emails (e.g. "https://apex.app"). Defaults to http://localhost:3000 for
	// local dev so the links land on the dev frontend.
	AppBaseURL string
	// Stripe billing settings for Variant A subscriptions.
	StripeSecretKey     string
	StripeWebhookSecret string
	StripeProPriceID    string
	StripeSuccessURL    string
	StripeCancelURL     string
	StripePortalReturn  string
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() *Config {
	cfg := &Config{
		Port:       env("PORT", "8080"),
		DBHost:     env("DB_HOST", "localhost"),
		DBPort:     env("DB_PORT", "3306"),
		DBUser:     env("DB_USER", "app"),
		DBPassword: env("DB_PASSWORD", "app"),
		DBName:     env("DB_NAME", "app"),
		// Empty = deny cross-origin (the SPA is same-origin behind nginx, so it
		// needs no CORS). Set to a comma-separated allowlist for direct API access.
		CORSOrigin:          env("CORS_ORIGIN", ""),
		CookieSecure:        env("COOKIE_SECURE", "false") == "true",
		EncryptionKey:       env("APP_ENCRYPTION_KEY", ""),
		IRacingClientID:     env("IRACING_CLIENT_ID", ""),
		IRacingClientSecret: env("IRACING_CLIENT_SECRET", ""),
		IRacingRedirectURI:  env("IRACING_OAUTH_REDIRECT_URI", ""),
		DeveloperKey:        env("DEVELOPER_KEY", ""),
		RedisAddr:           env("REDIS_ADDR", ""),
		AuthRateLimit:       envInt("AUTH_RATE_LIMIT", 20),
		SMTPHost:            env("SMTP_HOST", ""),
		SMTPPort:            env("SMTP_PORT", "587"),
		SMTPUser:            env("SMTP_USER", ""),
		SMTPPassword:        env("SMTP_PASSWORD", ""),
		SMTPFrom:            env("SMTP_FROM", ""),
		AppBaseURL:          env("APP_BASE_URL", "http://localhost:3000"),
		StripeSecretKey:     env("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: env("STRIPE_WEBHOOK_SECRET", ""),
		StripeProPriceID:    env("STRIPE_PRO_PRICE_ID", ""),
		StripeSuccessURL:    env("STRIPE_SUCCESS_URL", "http://localhost:3000/upgrade?checkout=success"),
		StripeCancelURL:     env("STRIPE_CANCEL_URL", "http://localhost:3000/upgrade?checkout=cancel"),
		StripePortalReturn:  env("STRIPE_PORTAL_RETURN_URL", "http://localhost:3000/upgrade"),
	}
	if cfg.DeveloperKey == "3" {
		// The compose default is a convenience for local dev, but it's a known
		// value — anyone who opens the app with ?dev=3 can toggle feature flags.
		// Warn loudly so a public deploy overrides DEVELOPER_KEY (or sets it
		// empty to disable the Cockpit entirely).
		log.Printf("config: DEVELOPER_KEY is the default \"3\" — override it for any non-local deploy (or set it empty to disable the Cockpit)")
	}
	return cfg
}

// envInt reads an integer env var, falling back to def on empty/invalid input.
func envInt(key string, def int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// DSN builds the MySQL data source name.
func (c *Config) DSN() string {
	// multiStatements lets the migration runner execute multi-statement files.
	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&loc=UTC&multiStatements=true",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName,
	)
}

func env(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
