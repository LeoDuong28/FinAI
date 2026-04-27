-- name: CreateAlert :one
INSERT INTO alerts (user_id, type, title, message, severity, reference_id, reference_type)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAlertsByUserID :many
SELECT * FROM alerts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListUnreadAlertsByUserID :many
SELECT * FROM alerts
WHERE user_id = $1 AND NOT is_read
ORDER BY created_at DESC;

-- name: CountUnreadAlerts :one
SELECT COUNT(*) FROM alerts WHERE user_id = $1 AND NOT is_read;

-- name: MarkAlertAsRead :exec
UPDATE alerts SET is_read = TRUE WHERE id = $1 AND user_id = $2;

-- name: MarkAllAlertsAsRead :execrows
UPDATE alerts SET is_read = TRUE WHERE user_id = $1 AND NOT is_read;

-- name: DeleteOldAlerts :execrows
DELETE FROM alerts WHERE created_at < NOW() - INTERVAL '90 days';
