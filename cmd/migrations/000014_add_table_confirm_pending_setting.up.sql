CREATE TABLE confirm_pending_setting (
  id                 SERIAL PRIMARY KEY,
  key_id             TEXT        NOT NULL,
  timestamp          TIMESTAMPTZ NOT NULL,
  setting_change_id  INT         NOT NULL,
  UNIQUE (key_id, setting_change_id),
  CONSTRAINT setting_change_fk
    FOREIGN KEY(setting_change_id)
      REFERENCES setting_change(id)
      ON DELETE CASCADE
);
