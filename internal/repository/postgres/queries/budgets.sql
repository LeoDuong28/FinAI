-- name: CreateBudget :one
INSERT INTO budgets (user_id, category_id, name, amount_limit, period, start_date, alert_threshold)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetBudgetByID :one
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM budgets b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.id = $1 AND b.user_id = $2;

-- name: ListBudgetsByUserID :many
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM budgets b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1 AND b.is_active = TRUE
ORDER BY b.name;

-- name: UpdateBudget :one
UPDATE budgets SET
    name = $3, amount_limit = $4, period = $5, start_date = $6, alert_threshold = $7
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeactivateBudget :exec
UPDATE budgets SET is_active = FALSE WHERE id = $1 AND user_id = $2;

-- name: DeleteBudget :exec
DELETE FROM budgets WHERE id = $1 AND user_id = $2;
