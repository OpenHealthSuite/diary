-- Storage functions for user config

-- name: GetUserConfig :one
SELECT * FROM user_config WHERE user_id = $1 AND id = $2;

-- name: StoreUserConfig :exec
INSERT INTO user_config (user_id, id, config_value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, id) DO UPDATE SET config_value = EXCLUDED.config_value;
