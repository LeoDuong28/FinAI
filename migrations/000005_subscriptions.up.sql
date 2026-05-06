CREATE TABLE subscriptions (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                 VARCHAR(255) NOT NULL,
    merchant_name        VARCHAR(255),
    amount               DECIMAL(15,2) NOT NULL,
    currency_code        VARCHAR(3) NOT NULL DEFAULT 'USD',
    frequency            VARCHAR(20) NOT NULL,  -- weekly|monthly|quarterly|yearly
    category_id          UUID REFERENCES categories(id),
    next_billing         DATE,
    last_charged         DATE,
    status               VARCHAR(20) NOT NULL DEFAULT 'active',  -- active|paused|cancelled
    auto_detected        BOOLEAN NOT NULL DEFAULT FALSE,
    detection_confidence DECIMAL(3,2),
    logo_url             TEXT,
    cancellation_url     TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subs_user ON subscriptions(user_id);
CREATE INDEX idx_subs_active ON subscriptions(user_id, status) WHERE status = 'active';
