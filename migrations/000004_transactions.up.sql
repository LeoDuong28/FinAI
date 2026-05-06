CREATE TABLE transactions (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id             UUID REFERENCES bank_accounts(id) ON DELETE SET NULL,
    category_id            UUID REFERENCES categories(id),
    plaid_txn_id           VARCHAR(100) UNIQUE,
    amount                 DECIMAL(15,2) NOT NULL,
    currency_code          VARCHAR(3) NOT NULL DEFAULT 'USD',
    date                   DATE NOT NULL,
    name                   VARCHAR(500) NOT NULL,
    merchant_name          VARCHAR(255),
    pending                BOOLEAN NOT NULL DEFAULT FALSE,
    type                   VARCHAR(20) NOT NULL,  -- debit | credit
    notes                  TEXT,
    is_excluded            BOOLEAN NOT NULL DEFAULT FALSE,
    is_recurring           BOOLEAN NOT NULL DEFAULT FALSE,
    ai_category_confidence DECIMAL(3,2),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_txn_user_date ON transactions(user_id, date DESC);
CREATE INDEX idx_txn_user_category ON transactions(user_id, category_id);
CREATE INDEX idx_txn_plaid ON transactions(plaid_txn_id) WHERE plaid_txn_id IS NOT NULL;
CREATE INDEX idx_txn_merchant ON transactions(user_id, merchant_name) WHERE merchant_name IS NOT NULL;
CREATE INDEX idx_txn_recurring ON transactions(user_id, is_recurring) WHERE is_recurring;
