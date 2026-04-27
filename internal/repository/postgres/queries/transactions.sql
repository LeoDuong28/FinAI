-- name: CreateTransaction :one
INSERT INTO transactions (
    user_id, account_id, category_id, plaid_txn_id, amount, currency_code,
    date, name, merchant_name, pending, type, notes, is_recurring, ai_category_confidence
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetTransactionByID :one
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.id = $1 AND t.user_id = $2;

-- name: ListTransactions :many
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1
ORDER BY t.date DESC, t.id DESC
LIMIT $2;

-- name: ListTransactionsWithCursor :many
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1
  AND (t.date, t.id) < (@cursor_date::DATE, @cursor_id::UUID)
ORDER BY t.date DESC, t.id DESC
LIMIT $2;

-- name: ListTransactionsByCategory :many
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1 AND t.category_id = $2
ORDER BY t.date DESC, t.id DESC
LIMIT $3;

-- name: ListTransactionsByDateRange :many
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1 AND t.date >= $2 AND t.date <= $3
ORDER BY t.date DESC, t.id DESC
LIMIT $4;

-- name: SearchTransactions :many
SELECT t.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color,
       similarity(t.name, $2) AS relevance
FROM transactions t
LEFT JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1
  AND (t.name % $2 OR t.merchant_name % $2)
ORDER BY relevance DESC
LIMIT $3;

-- name: UpdateTransactionCategory :exec
UPDATE transactions SET category_id = $3, ai_category_confidence = $4
WHERE id = $1 AND user_id = $2;

-- name: UpdateTransactionNotes :exec
UPDATE transactions SET notes = $3
WHERE id = $1 AND user_id = $2;

-- name: DeleteTransaction :exec
DELETE FROM transactions WHERE id = $1 AND user_id = $2;

-- name: GetTransactionByPlaidID :one
SELECT * FROM transactions WHERE plaid_txn_id = $1 AND user_id = $2;

-- name: CountTransactionsByUser :one
SELECT COUNT(*) FROM transactions WHERE user_id = $1;

-- name: SumTransactionsByUserAndDateRange :one
SELECT COALESCE(SUM(amount), 0)::DECIMAL(15,2) AS total
FROM transactions
WHERE user_id = $1 AND category_id IS NOT DISTINCT FROM $2 AND date >= $3 AND date <= $4 AND type = 'debit' AND NOT is_excluded;

-- name: UpsertTransactionByPlaidID :one
INSERT INTO transactions (user_id, account_id, plaid_txn_id, amount, currency_code, date, name, merchant_name, pending, type)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (plaid_txn_id, user_id) WHERE plaid_txn_id IS NOT NULL
DO UPDATE SET amount = EXCLUDED.amount, date = EXCLUDED.date, name = EXCLUDED.name,
    merchant_name = EXCLUDED.merchant_name, pending = EXCLUDED.pending
RETURNING *;

-- name: DeleteTransactionByPlaidID :exec
DELETE FROM transactions WHERE plaid_txn_id = $1 AND user_id = $2;

-- name: SumSpendingByUser :one
SELECT COALESCE(SUM(amount), 0)::DECIMAL(15,2) AS total
FROM transactions
WHERE user_id = $1 AND type = 'debit' AND NOT is_excluded
  AND date >= $2 AND date <= $3;

-- name: SumIncomeByUser :one
SELECT COALESCE(SUM(amount), 0)::DECIMAL(15,2) AS total
FROM transactions
WHERE user_id = $1 AND type = 'credit' AND NOT is_excluded
  AND date >= $2 AND date <= $3;

-- name: SpendingByCategory :many
SELECT c.id AS category_id, c.name AS category_name, c.icon, c.color,
       COALESCE(SUM(t.amount), 0)::DECIMAL(15,2) AS total,
       COUNT(t.id)::BIGINT AS txn_count
FROM transactions t
JOIN categories c ON t.category_id = c.id
WHERE t.user_id = $1 AND t.type = 'debit' AND NOT t.is_excluded
  AND t.date >= $2 AND t.date <= $3
GROUP BY c.id, c.name, c.icon, c.color
ORDER BY total DESC;

-- name: DailySpendingHistory :many
SELECT date, SUM(amount)::DECIMAL(15,2) AS total
FROM transactions
WHERE user_id = $1 AND type = 'debit' AND NOT is_excluded
  AND date >= $2 AND date <= $3
GROUP BY date
ORDER BY date;
