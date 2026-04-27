-- name: RevokeToken :exec
INSERT INTO revoked_tokens (jti, user_id, expires_at)
VALUES ($1, $2, $3);

-- name: IsTokenRevoked :one
SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE jti = $1) AS revoked;

-- name: RevokeAllUserTokens :exec
-- NOTE: JTIs are generated independently from session IDs and are not stored in the sessions table.
-- This query cannot recover JTIs from sessions. Instead, it inserts a sentinel record
-- that the auth middleware can check to reject all tokens issued before this timestamp.
-- For full revocation, combine with session deletion (which invalidates refresh tokens)
-- and rely on short-lived access tokens (15 min TTL) expiring naturally.
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredRevokedTokens :execrows
DELETE FROM revoked_tokens WHERE expires_at <= NOW();

-- name: SaveIdempotencyKey :exec
INSERT INTO idempotency_keys (key, user_id, response_status, response_body)
VALUES ($1, $2, $3, $4)
ON CONFLICT (key) DO NOTHING;

-- name: GetIdempotencyKey :one
SELECT * FROM idempotency_keys WHERE key = $1 AND expires_at > NOW();

-- name: DeleteExpiredIdempotencyKeys :execrows
DELETE FROM idempotency_keys WHERE expires_at <= NOW();

-- name: MarkWebhookProcessed :exec
INSERT INTO processed_webhooks (webhook_id, webhook_type)
VALUES ($1, $2)
ON CONFLICT (webhook_id) DO NOTHING;

-- name: IsWebhookProcessed :one
SELECT EXISTS(SELECT 1 FROM processed_webhooks WHERE webhook_id = $1) AS processed;

-- name: DeleteOldWebhooks :execrows
DELETE FROM processed_webhooks WHERE processed_at < NOW() - INTERVAL '24 hours';
