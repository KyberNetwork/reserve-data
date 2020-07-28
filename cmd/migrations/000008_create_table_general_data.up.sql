CREATE TABLE general_data (
  id SERIAL PRIMARY KEY,
  key   TEXT NOT NULL,
  value TEXT NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL
);
