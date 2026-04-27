-- name: CreateAuditLog :one
INSERT INTO audit_logs (user_id, action, entity_type, entity_id, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListAuditLogsByUserID :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: DeleteOldAuditLogs :execrows
DELETE FROM audit_logs WHERE created_at < $1;
