-- Storage functions for food log entries

-- name: CreateFoodLogEntry :one
INSERT INTO user_foodlogentry (user_id, name, labels, time_start, time_end)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: CreateFoodLogEntryMetric :exec
INSERT INTO user_foodlogentry_metrics (user_id, foodlogentry_id, metric_key, metric_value)
VALUES ($1, $2, $3, $4);

-- name: GetFoodLogEntry :one
SELECT * FROM user_foodlogentry WHERE user_id = $1 AND id = $2;

-- name: GetFoodLogEntryMetrics :many
SELECT metric_key, metric_value FROM user_foodlogentry_metrics WHERE user_id = $1 AND foodlogentry_id = $2;

-- name: QueryFoodLogEntries :many
SELECT ufle.*, jsonb_agg(
    jsonb_build_object('key', uflem.metric_key, 'value', uflem.metric_value)
  ) AS metrics
FROM user_foodlogentry ufle
INNER JOIN user_foodlogentry_metrics uflem ON ufle.id = uflem.foodlogentry_id AND uflem.user_id = ufle.user_id
WHERE ufle.user_id = $1 AND ufle.time_start >= $2 AND ufle.time_end <= $3
GROUP BY ufle.id
ORDER BY ufle.time_start ASC;

-- name: UpdateFoodLogEntry :one
UPDATE user_foodlogentry
SET name = $3, labels = $4, time_start = $5, time_end = $6
WHERE user_id = $1 AND id = $2
RETURNING id;

-- name: DeleteFoodLogEntry :exec
DELETE FROM user_foodlogentry WHERE user_id = $1 AND id = $2;

-- name: DeleteFoodLogEntryMetrics :exec
DELETE FROM user_foodlogentry_metrics WHERE user_id = $1 AND foodlogentry_id = $2;

-- name: PurgeFoodLogEntries :exec
DELETE FROM user_foodlogentry WHERE user_id = $1;

-- name: ExportFoodLogEntries :many
SELECT  ufle.*, jsonb_agg(
    jsonb_build_object('key', uflem.metric_key, 'value', uflem.metric_value)
  ) AS metrics
FROM user_foodlogentry ufle
INNER JOIN user_foodlogentry_metrics uflem ON ufle.id = uflem.foodlogentry_id AND uflem.user_id = ufle.user_id
WHERE ufle.user_id = $1 
GROUP BY ufle.id
ORDER BY ufle.time_start ASC;
