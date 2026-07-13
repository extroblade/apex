-- Turn the planner into a calendar: an optional scheduled time per plan row,
-- exported as an .ics feed. Additive column, existing data preserved.

ALTER TABLE plan_entries
    ADD COLUMN scheduled_at TIMESTAMP NULL DEFAULT NULL;
