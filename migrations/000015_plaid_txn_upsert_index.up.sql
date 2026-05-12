CREATE UNIQUE INDEX IF NOT EXISTS idx_txn_plaid_user_unique
ON transactions (plaid_txn_id, user_id) WHERE plaid_txn_id IS NOT NULL;
