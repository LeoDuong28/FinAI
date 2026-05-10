DROP TABLE IF EXISTS processed_webhooks;
DROP INDEX IF EXISTS idx_idempotency_expires;
DROP TABLE IF EXISTS idempotency_keys;
DROP INDEX IF EXISTS idx_revoked_tokens_expires;
DROP INDEX IF EXISTS idx_revoked_tokens_user;
DROP TABLE IF EXISTS revoked_tokens;
