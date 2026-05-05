-- Bank Accounts & Institutions
CREATE TABLE institutions (
    id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plaid_id VARCHAR(100) UNIQUE NOT NULL,
    name     VARCHAR(255) NOT NULL,
    logo_url TEXT,
    color    VARCHAR(7)
);

CREATE TABLE bank_accounts (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id            UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    institution_id     UUID REFERENCES institutions(id),
    plaid_account_id   TEXT,
    plaid_access_token TEXT,
    plaid_item_id      TEXT,
    name               VARCHAR(255) NOT NULL,
    official_name      TEXT,
    type               VARCHAR(50) NOT NULL,
    subtype            VARCHAR(50),
    mask               VARCHAR(4),
    current_balance    DECIMAL(15,2) NOT NULL DEFAULT 0,
    available_balance  DECIMAL(15,2),
    credit_limit       DECIMAL(15,2),
    currency_code      VARCHAR(3) NOT NULL DEFAULT 'USD',
    is_active          BOOLEAN NOT NULL DEFAULT TRUE,
    is_asset           BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at     TIMESTAMPTZ,
    sync_cursor        TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_accounts_user ON bank_accounts(user_id);
CREATE INDEX idx_accounts_active ON bank_accounts(user_id, is_active) WHERE is_active;
