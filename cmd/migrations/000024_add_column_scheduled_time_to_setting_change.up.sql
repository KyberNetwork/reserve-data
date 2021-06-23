ALTER TABLE setting_change ADD COLUMN scheduled_time TIMESTAMPTZ;

CREATE OR REPLACE FUNCTION new_setting_change(_cat setting_change.cat%TYPE, _data setting_change.data%TYPE, _proposer setting_change.proposer%TYPE, _scheduled_time setting_change.scheduled_time%TYPE)
    RETURNS int AS
$$

DECLARE
    _id setting_change.id%TYPE;

BEGIN
    INSERT INTO setting_change(created, cat, data, proposer, scheduled_time) VALUES (now(), _cat, _data, _proposer, COALESCE(_scheduled_time)) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;
