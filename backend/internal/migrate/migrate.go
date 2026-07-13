// Package migrate applies SQL migrations embedded in the binary at startup,
// tracking applied files in a schema_migrations table. This replaces relying on
// MySQL's docker-entrypoint-initdb.d (which only runs on a fresh volume), so
// schema changes apply on a normal restart.
//
// The DB connection must have multiStatements enabled (see config DSN) so each
// migration file can contain multiple statements.
package migrate

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var files embed.FS

// Run applies any not-yet-applied migrations in filename order.
func Run(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		name       VARCHAR(255) NOT NULL PRIMARY KEY,
		applied_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`); err != nil {
		return fmt.Errorf("migrate: create tracking table: %w", err)
	}

	applied, err := appliedSet(db)
	if err != nil {
		return err
	}

	names, err := migrationNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		if applied[name] {
			continue
		}
		content, err := files.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("migrate: apply %s: %w", name, err)
		}
		if _, err := db.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, name); err != nil {
			return fmt.Errorf("migrate: record %s: %w", name, err)
		}
		log.Printf("migrate: applied %s", name)
	}
	return nil
}

func appliedSet(db *sql.DB) (map[string]bool, error) {
	rows, err := db.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func migrationNames() ([]string, error) {
	entries, err := files.ReadDir("migrations")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}
