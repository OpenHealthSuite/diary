-- Storage functions for food log entries

-- name: CreateFoodLogEntry :one
INSERT INTO user_foodlogentry (user_id, id, name, labels, time_start, time_end)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: CreateFoodLogEntryMetric :exec
INSERT INTO user_foodlogentry_metrics (user_id, foodlogentry_id, metric_key, metric_value)
VALUES ($1, $2, $3, $4);

-- name: GetFoodLogEntry :one
SELECT * FROM user_foodlogentry WHERE user_id = $1 AND id = $2;

-- name: GetFoodLogEntryMetrics :many
SELECT metric_key, metric_value FROM user_foodlogentry_metrics WHERE user_id = $1 AND foodlogentry_id = $2;

-- name: QueryFoodLogEntries :many
SELECT * FROM user_foodlogentry
WHERE user_id = $1 AND time_start >= $2 AND time_end <= $3
ORDER BY time_start ASC;

-- name: QueryFoodLogEntryMetrics :many
SELECT foodlogentry_id, metric_key, metric_value FROM user_foodlogentry_metrics
WHERE user_id = $1 AND foodlogentry_id = ANY($2::uuid[]);

-- name: UpdateFoodLogEntry :one
UPDATE user_foodlogentry
SET name = $3, labels = $4, time_start = $5, time_end = $6
WHERE user_id = $1 AND id = $2
RETURNING *;

-- name: DeleteFoodLogEntry :exec
DELETE FROM user_foodlogentry WHERE user_id = $1 AND id = $2;

-- name: DeleteFoodLogEntryMetrics :exec
DELETE FROM user_foodlogentry_metrics WHERE user_id = $1 AND foodlogentry_id = $2;

-- name: PurgeFoodLogEntries :exec
DELETE FROM user_foodlogentry WHERE user_id = $1;

-- name: ExportFoodLogEntries :many
SELECT * FROM user_foodlogentry WHERE user_id = $1 ORDER BY time_start ASC;

-- name: ExportFoodLogEntryMetrics :many
SELECT foodlogentry_id, metric_key, metric_value FROM user_foodlogentry_metrics WHERE user_id = $1;