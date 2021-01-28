ALTER TABLE trading_pairs DROP COLUMN stall_threshold;
ALTER TABLE trading_pairs_deleted DROP COLUMN stall_threshold;

CREATE OR REPLACE FUNCTION new_trading_pair(_exchange_id trading_pairs.exchange_id%TYPE,
                                            _base_id trading_pairs.base_id%TYPE,
                                            _quote_id trading_pairs.quote_id%TYPE,
                                            _price_precision trading_pairs.price_precision%TYPE,
                                            _amount_precision trading_pairs.amount_precision%TYPE,
                                            _amount_limit_min trading_pairs.amount_limit_min%TYPE,
                                            _amount_limit_max trading_pairs.amount_limit_max%TYPE,
                                            _price_limit_min trading_pairs.price_limit_min%TYPE,
                                            _price_limit_max trading_pairs.price_limit_max%TYPE,
                                            _min_notional trading_pairs.min_notional%TYPE)
    RETURNS INT AS
$$
DECLARE
    _id                   trading_pairs.id%TYPE;
    _quote_asset_is_quote assets.is_quote%TYPE;
BEGIN
    PERFORM id FROM asset_exchanges WHERE exchange_id = _exchange_id AND asset_id = _base_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'base asset is not configured for exchange base_id=% exchange_id=%',
            _base_id,_exchange_id USING ERRCODE = 'KEBAS';
    END IF;

    PERFORM id FROM asset_exchanges WHERE exchange_id = _exchange_id AND asset_id = _quote_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'quote asset is not configured for exchange quote_id=% exchange_id=%',
            _quote_id,_exchange_id USING ERRCODE = 'KEQUO';
    END IF;

    PERFORM id FROM assets WHERE id = _base_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'base asset is not found base_id=%', _base_id USING ERRCODE = 'KEBAS';
    END IF;

    SELECT is_quote FROM assets WHERE id = _quote_id INTO _quote_asset_is_quote;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'quote asset is not found quote_id=%', _quote_id USING ERRCODE = 'KEQUO';
    END IF;

    IF NOT _quote_asset_is_quote THEN
        RAISE EXCEPTION 'quote asset is not configured as quote id=%', _quote_id USING ERRCODE = 'KEQUO';
    END IF;

    INSERT INTO trading_pairs (exchange_id,
                               base_id,
                               quote_id,
                               price_precision,
                               amount_precision,
                               amount_limit_min,
                               amount_limit_max,
                               price_limit_min,
                               price_limit_max,
                               min_notional)
    VALUES (_exchange_id,
            _base_id,
            _quote_id,
            _price_precision,
            _amount_precision,
            _amount_limit_min,
            _amount_limit_max,
            _price_limit_min,
            _price_limit_max,
            _min_notional) RETURNING id INTO _id;
    RETURN _id;
END
$$ LANGUAGE PLPGSQL;