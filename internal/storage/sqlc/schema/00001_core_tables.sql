CREATE TABLE IF NOT EXISTS user_foodlogentry (
  user_id TEXT NOT NULL,
  id UUID PRIMARY KEY,
  name TEXT NOT NULL,
  labels TEXT[] NOT NULL,
  time_start TIMESTAMP NOT NULL,
  time_end TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS user_foodlogentry_metrics (
  user_id TEXT NOT NULL,
  foodlogentry_id UUID NOT NULL REFERENCES user_foodlogentry(id) ON DELETE CASCADE,
  metric_key TEXT NOT NULL,
  metric_value INTEGER NOT NULL,
  PRIMARY KEY (user_id, foodlogentry_id, metric_key)
);

CREATE TABLE IF NOT EXISTS user_config (
  user_id UUID NOT NULL,
  id TEXT NOT NULL,
  config_value JSONB NOT NULL,
  PRIMARY KEY (user_id, id)
);