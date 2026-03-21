-- Storage functions for food log entries

-- name: CreateFoodLogEntry :one
INSERT INTO user_foodlogentry (id, user_id, name, labels, time_start, time_end)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: CreateFoodLogEntryMetric :exec
INSERT INTO user_foodlogentry_metrics (user_id, foodlogentry_id, metric_key, metric_value)
VALUES (?, ?, ?, ?);

-- name: GetFoodLogEntry :one
SELECT * FROM user_foodlogentry WHERE user_id = ? AND id = ?;

-- name: GetFoodLogEntryMetrics :many
SELECT metric_key, metric_value FROM user_foodlogentry_metrics WHERE user_id = ? AND foodlogentry_id = ?;

-- name: QueryFoodLogEntries :many
SELECT ufle.*, json_group_array(
    json_object('key', uflem.metric_key, 'value', uflem.metric_value)
  ) AS metrics
FROM user_foodlogentry ufle
INNER JOIN user_foodlogentry_metrics uflem ON ufle.id = uflem.foodlogentry_id AND uflem.user_id = ufle.user_id
WHERE ufle.user_id = ? AND ufle.time_start >= ? AND ufle.time_end <= ?
GROUP BY ufle.id
ORDER BY ufle.time_start ASC;

-- name: UpdateFoodLogEntry :exec
UPDATE user_foodlogentry
SET name = ?, labels = ?, time_start = ?, time_end = ?
WHERE user_id = ? AND id = ?;

-- name: DeleteFoodLogEntry :exec
DELETE FROM user_foodlogentry WHERE user_id = ? AND id = ?;

-- name: DeleteFoodLogEntryMetrics :exec
DELETE FROM user_foodlogentry_metrics WHERE user_id = ? AND foodlogentry_id = ?;

-- name: PurgeFoodLogEntries :exec
DELETE FROM user_foodlogentry WHERE user_id = ?;

-- name: ExportFoodLogEntries :many
SELECT  ufle.*, json_group_array(
    json_object('key', uflem.metric_key, 'value', uflem.metric_value)
  ) AS metrics
FROM user_foodlogentry ufle
INNER JOIN user_foodlogentry_metrics uflem ON ufle.id = uflem.foodlogentry_id AND uflem.user_id = ufle.user_id
WHERE ufle.user_id = ? 
GROUP BY ufle.id
ORDER BY ufle.time_start ASC;
