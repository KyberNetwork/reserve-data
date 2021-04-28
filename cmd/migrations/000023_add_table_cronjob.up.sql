CREATE TABLE cron_job (
  id             SERIAL PRIMARY KEY,
  schedule_time  TIMESTAMPTZ NOT NULL,
  data           JSON NOT NULL,
  http_method    TEXT NOT NULL,
  endpoint       TEXT NOT NULL
);
