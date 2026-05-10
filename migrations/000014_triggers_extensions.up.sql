-- Enable pg_trgm for fuzzy text search
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Auto-update updated_at on any UPDATE
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to all tables with updated_at
CREATE TRIGGER set_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON bank_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON transactions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON budgets
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON bills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
CREATE TRIGGER set_updated_at BEFORE UPDATE ON savings_goals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Full-text search indexes (fuzzy matching on transaction names)
CREATE INDEX idx_txn_name_trgm ON transactions USING GIN (name gin_trgm_ops);
CREATE INDEX idx_txn_merchant_trgm ON transactions USING GIN (merchant_name gin_trgm_ops)
    WHERE merchant_name IS NOT NULL;

-- Materialized views for insights
CREATE MATERIALIZED VIEW mv_monthly_spending AS
SELECT user_id, category_id, DATE_TRUNC('month', date)::DATE AS month,
       SUM(amount) AS total, COUNT(*) AS txn_count
FROM transactions
WHERE type = 'debit' AND NOT is_excluded
GROUP BY user_id, category_id, DATE_TRUNC('month', date)::DATE;

CREATE UNIQUE INDEX ON mv_monthly_spending(user_id, COALESCE(category_id, '00000000-0000-0000-0000-000000000000'), month);

CREATE MATERIALIZED VIEW mv_top_merchants AS
SELECT user_id, merchant_name, COUNT(*) AS txn_count,
       SUM(amount) AS total_spent
FROM transactions
WHERE merchant_name IS NOT NULL AND type = 'debit'
GROUP BY user_id, merchant_name;

CREATE UNIQUE INDEX ON mv_top_merchants(user_id, merchant_name);
