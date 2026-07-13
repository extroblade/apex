package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql" // registers the "mysql" driver

	"apex/internal/config"
)

// Connect opens a MySQL connection pool and waits until the database is reachable.
func Connect(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.DSN()
	pool, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	pool.SetMaxOpenConns(25)
	pool.SetMaxIdleConns(25)
	pool.SetConnMaxLifetime(5 * time.Minute)

	// Retry ping so the API can start alongside a still-booting MySQL container.
	var lastErr error
	for attempt := 1; attempt <= 30; attempt++ {
		if lastErr = pool.Ping(); lastErr == nil {
			log.Println("db: connected")
			return pool, nil
		}
		log.Printf("db: waiting for mysql (attempt %d/30): %v", attempt, lastErr)
		time.Sleep(2 * time.Second)
	}
	return nil, lastErr
}
