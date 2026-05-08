CREATE TABLE budgets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category_id     UUID REFERENCES categories(id),
    name            VARCHAR(255) NOT NULL,
    amount_limit    DECIMAL(15,2) NOT NULL,
    period          VARCHAR(20) NOT NULL DEFAULT 'monthly',  -- weekly|monthly|yearly
    start_date      DATE NOT NULL,
    alert_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.80,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_category_budget UNIQUE(user_id, category_id, period)
);

CREATE INDEX idx_budgets_user ON budgets(user_id);
