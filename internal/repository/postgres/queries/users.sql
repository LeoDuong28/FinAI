-- name: CreateUser :one
INSERT INTO users (email, password_hash, first_name, last_name, currency, timezone)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users SET
    first_name = $2,
    last_name = $3,
    avatar_url = $4,
    theme = $5,
    currency = $6,
    timezone = $7,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserEmailVerified :exec
UPDATE users SET email_verified = TRUE WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: UpdateUserOnboardingCompleted :exec
UPDATE users SET onboarding_completed = TRUE WHERE id = $1;

-- name: SetTOTPSecret :exec
UPDATE users SET totp_secret = $2, totp_enabled = $3 WHERE id = $1;

-- name: IncrementFailedLogin :one
UPDATE users SET failed_login_attempts = failed_login_attempts + 1 WHERE id = $1
RETURNING failed_login_attempts;

-- name: ResetFailedLogin :exec
UPDATE users SET failed_login_attempts = 0, locked_until = NULL WHERE id = $1;

-- name: LockAccount :exec
UPDATE users SET locked_until = $2 WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;
