-- name: CreateSubscription :one
INSERT INTO subscriptions (
    user_id, name, merchant_name, amount, currency_code, frequency,
    category_id, next_billing, last_charged, status, auto_detected,
    detection_confidence, logo_url, cancellation_url
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING *;

-- name: GetSubscriptionByID :one
SELECT s.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM subscriptions s
LEFT JOIN categories c ON s.category_id = c.id
WHERE s.id = $1 AND s.user_id = $2;

-- name: ListSubscriptionsByUserID :many
SELECT s.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM subscriptions s
LEFT JOIN categories c ON s.category_id = c.id
WHERE s.user_id = $1
ORDER BY s.status, s.next_billing;

-- name: ListActiveSubscriptionsByUserID :many
SELECT s.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM subscriptions s
LEFT JOIN categories c ON s.category_id = c.id
WHERE s.user_id = $1 AND s.status = 'active'
ORDER BY s.next_billing;

-- name: UpdateSubscription :one
UPDATE subscriptions SET
    name = $3, amount = $4, frequency = $5, category_id = $6,
    next_billing = $7, status = $8
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateSubscriptionLastCharged :exec
UPDATE subscriptions SET last_charged = $3 WHERE id = $1 AND user_id = $2;

-- name: DeleteSubscription :exec
DELETE FROM subscriptions WHERE id = $1 AND user_id = $2;

-- name: SumActiveSubscriptions :one
SELECT COALESCE(SUM(
    CASE frequency
        WHEN 'weekly' THEN amount * 4.33
        WHEN 'monthly' THEN amount
        WHEN 'quarterly' THEN amount / 3
        WHEN 'yearly' THEN amount / 12
        ELSE amount
    END
), 0)::DECIMAL(15,2) AS monthly_total
FROM subscriptions
WHERE user_id = $1 AND status = 'active';
