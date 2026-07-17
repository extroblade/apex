package main

import (
	"context"
	"database/sql"
	"log"
)

// schema creates the table this service owns. It is intentionally self-managed
// (rather than living in the API's migration runner) so the service stays
// independently deployable.
const schema = `
CREATE TABLE IF NOT EXISTS nav_items (
    id            INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    item_key      VARCHAR(64)  NOT NULL UNIQUE,
    label_key     VARCHAR(128) NOT NULL,
    href          VARCHAR(255) NOT NULL,
    icon          VARCHAR(64)  NOT NULL DEFAULT '',
    -- Comma-separated: "side" (desktop sidebar) and/or "bottom" (mobile bar).
    placements    VARCHAR(64)  NOT NULL DEFAULT 'side',
    sort_order    INT          NOT NULL DEFAULT 0,
    requires_auth TINYINT(1)   NOT NULL DEFAULT 0,
    -- Empty means the item is not gated behind a feature flag.
    feature_flag  VARCHAR(64)  NOT NULL DEFAULT '',
    enabled       TINYINT(1)   NOT NULL DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`

// seed is the default menu — it reproduces the nav the SPA used to hard-code.
// INSERT IGNORE keys off item_key's UNIQUE index, so this is idempotent: a fresh
// DB gets the defaults, an upgrade picks up newly added items, and rows edited
// through the Cockpit are left alone.
const seed = `
INSERT IGNORE INTO nav_items
    (item_key, label_key, href, icon, placements, sort_order, requires_auth, feature_flag)
VALUES
    ('home',      'nav.home',      '/',          'home',           'side,bottom', 10,  0, ''),
    ('fuel',      'nav.fuel',      '/fuel',      'fuel',           'side,bottom', 20,  0, ''),
    ('planner',   'nav.planner',   '/planner',   'calendar-range', 'side,bottom', 30,  1, ''),
    ('garage',    'nav.garage',    '/garage',    'warehouse',      'side,bottom', 40,  1, ''),
    ('setups',    'nav.setups',    '/setups',    'wrench',         'side,bottom', 50,  1, ''),
    ('goals',     'nav.goals',     '/goals',     'target',         'side,bottom', 60,  1, ''),
    ('drivers',   'nav.drivers',   '/drivers',   'users',          'side,bottom', 70,  1, 'iracing_oauth'),
    ('dashboard', 'nav.dashboard', '/dashboard', 'gauge',          'side,bottom', 80,  1, 'iracing_oauth'),
    ('compare',   'nav.compare',   '/compare',   'bar-chart-3',    'side,bottom', 90,  1, 'iracing_oauth');`

// migrate creates and seeds the nav table.
func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return err
	}
	res, err := db.ExecContext(ctx, seed)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n > 0 {
		log.Printf("nav: seeded %d default menu item(s)", n)
	}
	return nil
}
