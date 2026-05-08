CREATE TABLE savings_goals (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                 VARCHAR(255) NOT NULL,
    target_amount        DECIMAL(15,2) NOT NULL,
    current_amount       DECIMAL(15,2) NOT NULL DEFAULT 0,
    target_date          DATE,
    monthly_contribution DECIMAL(15,2),
    icon                 VARCHAR(50),
    color                VARCHAR(7),
    status               VARCHAR(20) NOT NULL DEFAULT 'active',  -- active|completed|cancelled
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goals_user ON savings_goals(user_id);
