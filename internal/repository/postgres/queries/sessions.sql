-- name: CreateSession :one
INSERT INTO sessions (user_id, refresh_token, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByToken :one
SELECT * FROM sessions WHERE refresh_token = $1 AND expires_at > NOW();

-- name: DeleteSessionByID :exec
DELETE FROM sessions WHERE id = $1 AND user_id = $2;

-- name: DeleteSessionsByUserID :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteExpiredSessions :execrows
DELETE FROM sessions WHERE expires_at <= NOW();
