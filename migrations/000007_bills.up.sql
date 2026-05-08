CREATE TABLE bills (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            VARCHAR(255) NOT NULL,
    amount          DECIMAL(15,2) NOT NULL,
    due_date        DATE NOT NULL,
    frequency       VARCHAR(20) NOT NULL,  -- once|monthly|quarterly|yearly
    category_id     UUID REFERENCES categories(id),
    is_autopay      BOOLEAN NOT NULL DEFAULT FALSE,
    status          VARCHAR(20) NOT NULL DEFAULT 'upcoming',  -- upcoming|paid|overdue
    reminder_days   INT NOT NULL DEFAULT 3,
    negotiation_tip TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_bills_user ON bills(user_id);
CREATE INDEX idx_bills_due ON bills(user_id, due_date);
CREATE INDEX idx_bills_overdue ON bills(status) WHERE status = 'overdue';
