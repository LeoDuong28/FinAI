-- Index for Plaid webhook hot path: ListBankAccountsByPlaidItemID
CREATE INDEX IF NOT EXISTS idx_accounts_plaid_item
ON bank_accounts(plaid_item_id) WHERE is_active;

-- Index for MarkOverdueBills batch query performance
CREATE INDEX IF NOT EXISTS idx_bills_upcoming_due
ON bills(due_date) WHERE status = 'upcoming';

-- Fix NULL category_id bypassing unique budget-per-period constraint.
-- PostgreSQL treats NULL != NULL in unique indexes, allowing unlimited
-- "uncategorized" budgets per period. Add a partial unique index for NULLs.
CREATE UNIQUE INDEX IF NOT EXISTS idx_budget_user_period_no_category
ON budgets(user_id, period) WHERE category_id IS NULL;
