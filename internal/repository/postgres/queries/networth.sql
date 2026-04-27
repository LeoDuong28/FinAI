-- name: CreateNetworthSnapshot :one
INSERT INTO networth_snapshots (user_id, total_assets, total_liabilities, net_worth, snapshot_date, breakdown)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, snapshot_date) DO UPDATE SET
    total_assets = EXCLUDED.total_assets,
    total_liabilities = EXCLUDED.total_liabilities,
    net_worth = EXCLUDED.net_worth,
    breakdown = EXCLUDED.breakdown
RETURNING *;

-- name: ListNetworthSnapshots :many
SELECT * FROM networth_snapshots
WHERE user_id = $1
ORDER BY snapshot_date DESC
LIMIT $2;

-- name: GetLatestNetworthSnapshot :one
SELECT * FROM networth_snapshots
WHERE user_id = $1
ORDER BY snapshot_date DESC
LIMIT 1;
