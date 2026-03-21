CREATE TABLE IF NOT EXISTS user_foodlogentry (
  user_id TEXT NOT NULL,
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  labels JSONB NOT NULL,
  time_start TIMESTAMP NOT NULL,
  time_end TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS user_foodlogentry_metrics (
  user_id TEXT NOT NULL,
  foodlogentry_id TEXT NOT NULL REFERENCES user_foodlogentry(id) ON DELETE CASCADE,
  metric_key TEXT NOT NULL,
  metric_value double precision NOT NULL,
  PRIMARY KEY (user_id, foodlogentry_id, metric_key)
);

CREATE TABLE IF NOT EXISTS user_config (
  user_id TEXT NOT NULL,
  id TEXT NOT NULL,
  config_value JSONB NOT NULL,
  PRIMARY KEY (user_id, id)
);