-- Drop the single-column UNIQUE constraint on plaid_txn_id from migration 000004.
-- The composite unique index (plaid_txn_id, user_id) from migration 000015 is the
-- correct constraint for the ON CONFLICT upsert clause.
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_plaid_txn_id_key;
