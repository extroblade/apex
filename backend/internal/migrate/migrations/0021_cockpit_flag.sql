-- Cockpit dev-overlay: gated by a developer cookie matching DEVELOPER_KEY env.
-- Off by default; the flag itself is the first-layer gate (404 when off) so
-- the cookie check only runs when the flag is on.

INSERT INTO feature_flags (flag_key, enabled, description) VALUES
    ('cockpit', 0, 'Dev overlay: runtime feature-flag toggles and debug logging')
ON DUPLICATE KEY UPDATE description = VALUES(description);
