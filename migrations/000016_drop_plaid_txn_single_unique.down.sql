-- Restore the single-column UNIQUE constraint on plaid_txn_id.
ALTER TABLE transactions ADD CONSTRAINT transactions_plaid_txn_id_key UNIQUE (plaid_txn_id);
