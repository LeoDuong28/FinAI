-- name: CreateInstitution :one
INSERT INTO institutions (plaid_id, name, logo_url, color)
VALUES ($1, $2, $3, $4)
ON CONFLICT (plaid_id) DO UPDATE SET name = EXCLUDED.name, logo_url = EXCLUDED.logo_url, color = EXCLUDED.color
RETURNING *;

-- name: GetInstitutionByPlaidID :one
SELECT * FROM institutions WHERE plaid_id = $1;

-- name: CreateBankAccount :one
INSERT INTO bank_accounts (
    user_id, institution_id, plaid_account_id, plaid_access_token, plaid_item_id,
    name, official_name, type, subtype, mask, current_balance, available_balance,
    credit_limit, currency_code, is_asset
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetBankAccountByID :one
SELECT * FROM bank_accounts WHERE id = $1 AND user_id = $2;

-- name: ListBankAccountsByUserID :many
SELECT ba.*, i.name AS institution_name, i.logo_url AS institution_logo, i.color AS institution_color
FROM bank_accounts ba
LEFT JOIN institutions i ON ba.institution_id = i.id
WHERE ba.user_id = $1 AND ba.is_active = TRUE
ORDER BY ba.created_at;

-- name: UpdateBankAccountBalance :exec
UPDATE bank_accounts SET
    current_balance = $3,
    available_balance = $4,
    last_synced_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: UpdateBankAccountSyncCursor :exec
UPDATE bank_accounts SET sync_cursor = $3, last_synced_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: DeactivateBankAccount :exec
UPDATE bank_accounts SET is_active = FALSE WHERE id = $1 AND user_id = $2;

-- name: DeleteBankAccount :exec
DELETE FROM bank_accounts WHERE id = $1 AND user_id = $2;

-- name: ListBankAccountsByPlaidItemID :many
SELECT * FROM bank_accounts WHERE plaid_item_id = $1 AND is_active = TRUE;

-- name: GetBankAccountForSync :one
SELECT * FROM bank_accounts WHERE id = $1 AND user_id = $2 AND is_active = TRUE;

-- name: SumAssetBalances :one
SELECT COALESCE(SUM(current_balance), 0)::DECIMAL(15,2) AS total
FROM bank_accounts
WHERE user_id = $1 AND is_active = TRUE AND is_asset = TRUE;

-- name: SumLiabilityBalances :one
SELECT COALESCE(SUM(current_balance), 0)::DECIMAL(15,2) AS total
FROM bank_accounts
WHERE user_id = $1 AND is_active = TRUE AND is_asset = FALSE;
