-- Revoked JWT tokens (for logout + password change invalidation)
CREATE TABLE revoked_tokens (
    jti        UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    revoked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_revoked_tokens_user ON revoked_tokens(user_id);
CREATE INDEX idx_revoked_tokens_expires ON revoked_tokens(expires_at);

-- Idempotency keys for safe retries on POST endpoints
CREATE TABLE idempotency_keys (
    key             VARCHAR(255) PRIMARY KEY,
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    response_status INT NOT NULL,
    response_body   TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at      TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '24 hours'
);

CREATE INDEX idx_idempotency_expires ON idempotency_keys(expires_at);

-- Plaid webhook deduplication
CREATE TABLE processed_webhooks (
    webhook_id   VARCHAR(100) PRIMARY KEY,
    webhook_type VARCHAR(100),
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
