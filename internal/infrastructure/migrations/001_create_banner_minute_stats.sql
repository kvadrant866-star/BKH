-- +goose Up
CREATE TABLE IF NOT EXISTS banner_minute_stats (
  banner_id  BIGINT      NOT NULL,
  minute_ts  TIMESTAMPTZ NOT NULL,
  count      BIGINT      NOT NULL DEFAULT 0,
  PRIMARY KEY (banner_id, minute_ts)
);

-- +goose Down
DROP TABLE IF EXISTS banner_minute_stats;


