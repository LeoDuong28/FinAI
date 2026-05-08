CREATE TABLE alerts (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type           VARCHAR(50) NOT NULL,
    title          VARCHAR(255) NOT NULL,
    message        TEXT NOT NULL,
    severity       VARCHAR(20) NOT NULL DEFAULT 'info',  -- info|warning|critical
    is_read        BOOLEAN NOT NULL DEFAULT FALSE,
    reference_id   UUID,
    reference_type VARCHAR(50),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_alerts_user_unread ON alerts(user_id, is_read, created_at DESC) WHERE NOT is_read;
CREATE INDEX idx_alerts_user_recent ON alerts(user_id, created_at DESC);
