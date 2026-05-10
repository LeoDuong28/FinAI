CREATE TABLE networth_snapshots (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    total_assets      DECIMAL(15,2) NOT NULL,
    total_liabilities DECIMAL(15,2) NOT NULL,
    net_worth         DECIMAL(15,2) NOT NULL,
    snapshot_date     DATE NOT NULL,
    breakdown         JSONB,
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT unique_user_snapshot_date UNIQUE(user_id, snapshot_date)
);

CREATE INDEX idx_networth_user ON networth_snapshots(user_id, snapshot_date DESC);
