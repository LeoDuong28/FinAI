-- name: CreateBill :one
INSERT INTO bills (user_id, name, amount, due_date, frequency, category_id, is_autopay, reminder_days)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetBillByID :one
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM bills b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.id = $1 AND b.user_id = $2;

-- name: ListBillsByUserID :many
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM bills b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1
ORDER BY b.due_date;

-- name: ListUpcomingBills :many
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM bills b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1 AND b.status = 'upcoming'
  AND b.due_date >= CURRENT_DATE
  AND b.due_date <= CURRENT_DATE + ($2::INT * INTERVAL '1 day')
ORDER BY b.due_date;

-- name: ListOverdueBills :many
SELECT b.*, c.name AS category_name, c.icon AS category_icon, c.color AS category_color
FROM bills b
LEFT JOIN categories c ON b.category_id = c.id
WHERE b.user_id = $1 AND b.status = 'overdue'
ORDER BY b.due_date;

-- name: UpdateBill :one
UPDATE bills SET
    name = $3, amount = $4, due_date = $5, frequency = $6,
    category_id = $7, is_autopay = $8, reminder_days = $9
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateBillStatus :exec
UPDATE bills SET status = $3 WHERE id = $1 AND user_id = $2;

-- name: UpdateBillNegotiationTip :exec
UPDATE bills SET negotiation_tip = $3 WHERE id = $1 AND user_id = $2;

-- name: DeleteBill :exec
DELETE FROM bills WHERE id = $1 AND user_id = $2;

-- name: MarkOverdueBills :execrows
UPDATE bills SET status = 'overdue'
WHERE status = 'upcoming' AND due_date < CURRENT_DATE;
