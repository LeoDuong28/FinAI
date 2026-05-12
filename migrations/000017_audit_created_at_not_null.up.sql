-- Backfill any NULL created_at values, then add NOT NULL constraint.
UPDATE audit_logs SET created_at = NOW() WHERE created_at IS NULL;
ALTER TABLE audit_logs ALTER COLUMN created_at SET NOT NULL;
