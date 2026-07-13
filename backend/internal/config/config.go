package config

import (
	"fmt"
	"os"
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
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() *Config {
	return &Config{
		Port:                env("PORT", "8080"),
		DBHost:              env("DB_HOST", "localhost"),
		DBPort:              env("DB_PORT", "3306"),
		DBUser:              env("DB_USER", "app"),
		DBPassword:          env("DB_PASSWORD", "app"),
		DBName:              env("DB_NAME", "app"),
		CORSOrigin:          env("CORS_ORIGIN", "*"),
		CookieSecure:        env("COOKIE_SECURE", "false") == "true",
		EncryptionKey:       env("APP_ENCRYPTION_KEY", ""),
		IRacingClientID:     env("IRACING_CLIENT_ID", ""),
		IRacingClientSecret: env("IRACING_CLIENT_SECRET", ""),
		IRacingRedirectURI:  env("IRACING_OAUTH_REDIRECT_URI", ""),
		DeveloperKey:        env("DEVELOPER_KEY", ""),
		RedisAddr:           env("REDIS_ADDR", ""),
	}
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
