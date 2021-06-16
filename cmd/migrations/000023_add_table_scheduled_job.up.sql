CREATE TYPE scheduled_job_status AS ENUM ('pending', 'done', 'canceled');

CREATE TABLE scheduled_job (
  id             SERIAL PRIMARY KEY,
  scheduled_time  TIMESTAMPTZ NOT NULL,
  data           JSON NOT NULL,
  http_method    TEXT NOT NULL,
  endpoint       TEXT NOT NULL,
  status scheduled_job_status NOT NULL DEFAULT 'pending'
);
