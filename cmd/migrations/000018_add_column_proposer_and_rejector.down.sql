ALTER TABLE setting_change DROP COLUMN proposer;
ALTER TABLE setting_change DROP COLUMN rejector;

CREATE OR REPLACE FUNCTION new_setting_change(_cat setting_change.cat%TYPE, _data setting_change.data%TYPE)
    RETURNS int AS
$$

DECLARE
    _id setting_change.id%TYPE;

BEGIN
    INSERT INTO setting_change(created, cat, data) VALUES (now(), _cat, _data) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;
