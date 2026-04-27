-- name: CreateSavingsGoal :one
INSERT INTO savings_goals (user_id, name, target_amount, target_date, monthly_contribution, icon, color)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetSavingsGoalByID :one
SELECT * FROM savings_goals WHERE id = $1 AND user_id = $2;

-- name: ListSavingsGoalsByUserID :many
SELECT * FROM savings_goals
WHERE user_id = $1
ORDER BY status, created_at;

-- name: UpdateSavingsGoal :one
UPDATE savings_goals SET
    name = $3, target_amount = $4, target_date = $5,
    monthly_contribution = $6, icon = $7, color = $8
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: AddFundsToSavingsGoal :one
UPDATE savings_goals SET
    current_amount = current_amount + $3,
    status = CASE WHEN current_amount + $3 >= target_amount THEN 'completed' ELSE status END
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: WithdrawFundsFromSavingsGoal :one
UPDATE savings_goals SET
    current_amount = current_amount - $3,
    status = CASE WHEN current_amount - $3 < target_amount THEN 'active' ELSE status END
WHERE id = $1 AND user_id = $2 AND current_amount >= $3
RETURNING *;

-- name: DeleteSavingsGoal :exec
DELETE FROM savings_goals WHERE id = $1 AND user_id = $2;

-- name: SumSavingsProgress :one
SELECT COALESCE(SUM(current_amount), 0)::DECIMAL(15,2) AS total_saved,
       COALESCE(SUM(target_amount), 0)::DECIMAL(15,2) AS total_target
FROM savings_goals
WHERE user_id = $1 AND status != 'cancelled';
