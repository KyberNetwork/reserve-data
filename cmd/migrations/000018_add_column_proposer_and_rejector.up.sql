ALTER TABLE setting_change ADD COLUMN proposer TEXT;
ALTER TABLE setting_change ADD COLUMN rejector TEXT;

CREATE OR REPLACE FUNCTION new_setting_change(_cat setting_change.cat%TYPE, _data setting_change.data%TYPE, _proposer setting_change.proposer%TYPE)
    RETURNS int AS
$$

DECLARE
    _id setting_change.id%TYPE;

BEGIN
    INSERT INTO setting_change(created, cat, data, proposer) VALUES (now(), _cat, _data, _proposer) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;
