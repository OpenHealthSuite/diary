-- Storage functions for user config

-- name: GetUserConfig :one
SELECT * FROM user_config WHERE user_id = ? AND id = ?;

-- name: StoreUserConfig :exec
INSERT INTO user_config (user_id, id, config_value)
VALUES (?, ?, ?)
ON CONFLICT (user_id, id) DO UPDATE SET config_value = EXCLUDED.config_value;
