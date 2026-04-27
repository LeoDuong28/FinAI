-- name: CreateChatMessage :one
INSERT INTO chat_messages (user_id, session_id, role, content, context_data, tokens_used)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListChatMessages :many
SELECT * FROM chat_messages
WHERE user_id = $1 AND session_id = $2
ORDER BY created_at ASC;

-- name: ListRecentChatSessions :many
WITH recent_sessions AS (
    SELECT cm1.session_id, MAX(cm1.created_at)::TIMESTAMPTZ AS last_activity
    FROM chat_messages cm1
    WHERE cm1.user_id = $1
    GROUP BY cm1.session_id
    ORDER BY last_activity DESC
    LIMIT 10
),
first_messages AS (
    SELECT DISTINCT ON (cm2.session_id)
        cm2.session_id,
        rs.last_activity AS created_at,
        cm2.content
    FROM chat_messages cm2
    JOIN recent_sessions rs ON cm2.session_id = rs.session_id
    WHERE cm2.role = 'user'
    ORDER BY cm2.session_id, cm2.created_at ASC
)
SELECT session_id, created_at, content
FROM first_messages
ORDER BY created_at DESC;

-- name: DeleteChatHistory :execrows
DELETE FROM chat_messages WHERE user_id = $1;

-- name: DeleteOldChatMessages :execrows
DELETE FROM chat_messages WHERE created_at < NOW() - INTERVAL '90 days';
