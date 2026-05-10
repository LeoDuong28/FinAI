DROP MATERIALIZED VIEW IF EXISTS mv_top_merchants;
DROP MATERIALIZED VIEW IF EXISTS mv_monthly_spending;

DROP INDEX IF EXISTS idx_txn_merchant_trgm;
DROP INDEX IF EXISTS idx_txn_name_trgm;

DROP TRIGGER IF EXISTS set_updated_at ON savings_goals;
DROP TRIGGER IF EXISTS set_updated_at ON bills;
DROP TRIGGER IF EXISTS set_updated_at ON budgets;
DROP TRIGGER IF EXISTS set_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS set_updated_at ON transactions;
DROP TRIGGER IF EXISTS set_updated_at ON bank_accounts;
DROP TRIGGER IF EXISTS set_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at();

DROP EXTENSION IF EXISTS pg_trgm;
