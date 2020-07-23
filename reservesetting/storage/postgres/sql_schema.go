package postgres

const schema = `CREATE TABLE IF NOT EXISTS "exchanges"
(
    id                INT PRIMARY KEY,
    name              TEXT UNIQUE NOT NULL,
    trading_fee_maker FLOAT,
    trading_fee_taker FLOAT,
    disable           BOOLEAN     NOT NULL DEFAULT TRUE
        -- only allow to enable exchange if trading_fee_maker and trading_fee_taker are both set
        CONSTRAINT disable_check CHECK (disable OR
                                        ((trading_fee_maker IS NOT NULL) AND (trading_fee_taker IS NOT NULL)))
);

CREATE TABLE IF NOT EXISTS "addresses"
(
    id      SERIAL PRIMARY KEY,
    address TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS "assets"
(
    id                            SERIAL PRIMARY KEY,
    symbol                        TEXT      NOT NULL UNIQUE,
    name                          TEXT      NOT NULL,
    address_id                    INT       NULL REFERENCES addresses (id)
        CONSTRAINT address_id_check CHECK ( address_id IS NOT NULL OR NOT transferable),
    decimals                      INT       NOT NULL,
    -- transferable must be set to true if it is possible to withdraw/deposit 
    -- to reserve (ETH or ERC20 tokens). If transferable is true, the address and 
    -- deposit address of related asset_exchange records are required.
    transferable                  BOOLEAN   NOT NULL,
    set_rate                      TEXT      NOT NULL,
    rebalance                     BOOLEAN   NOT NULL,
    is_quote                      BOOLEAN   NOT NULL,

    pwi_ask_a                     FLOAT     NULL,
    pwi_ask_b                     FLOAT     NULL,
    pwi_ask_c                     FLOAT     NULL,
    pwi_ask_min_min_spread        FLOAT     NULL,
    pwi_ask_price_multiply_factor FLOAT     NULL,
    pwi_bid_a                     FLOAT     NULL,
    pwi_bid_b                     FLOAT     NULL,
    pwi_bid_c                     FLOAT     NULL,
    pwi_bid_min_min_spread        FLOAT     NULL,
    pwi_bid_price_multiply_factor FLOAT     NULL,

    rebalance_size_quadratic_a    FLOAT     NULL,
    rebalance_size_quadratic_b    FLOAT     NULL,
    rebalance_size_quadratic_c    FLOAT     NULL,
    rebalance_price_quadratic_a   FLOAT     NULL,
    rebalance_price_quadratic_b   FLOAT     NULL,
    rebalance_price_quadratic_c   FLOAT     NULL,

    target_total                  FLOAT     NULL,
    target_reserve                FLOAT     NULL,
    target_rebalance_threshold    FLOAT     NULL,
    target_transfer_threshold     FLOAT     NULL,
	
	stable_param_price_update_threshold 	FLOAT	DEFAULT 0,
	stable_param_ask_spread					FLOAT	DEFAULT	0,
	stable_param_bid_spread					FLOAT	DEFAULT	0,
	stable_param_single_feed_max_spread		FLOAT	DEFAULT	0,
    stable_param_multiple_feeds_max_diff 	FLOAT	DEFAULT 0,
    
    normal_update_per_period FLOAT DEFAULT 1 
    CONSTRAINT normal_update_per_period_check CHECK(normal_update_per_period > 0),
    max_imbalance_ratio FLOAT DEFAULT 2
    CONSTRAINT max_imbalance_ratio_check CHECK(max_imbalance_ratio > 0),

    created                       TIMESTAMPTZ NOT NULL,
    updated                       TIMESTAMPTZ NOT NULL,

    is_enabled BOOLEAN NOT NULL DEFAULT TRUE
    -- if set_rate strategy is defined, pwi columns are required
    CONSTRAINT pwi_check CHECK (
            set_rate = 'not_set'
            OR (pwi_ask_a IS NOT NULL AND
                pwi_ask_b IS NOT NULL AND
                pwi_ask_c IS NOT NULL AND
                pwi_ask_min_min_spread IS NOT NULL AND
                pwi_ask_price_multiply_factor IS NOT NULL AND
                pwi_bid_a IS NOT NULL AND
                pwi_bid_b IS NOT NULL AND
                pwi_bid_c IS NOT NULL AND
                pwi_bid_min_min_spread IS NOT NULL AND
                pwi_bid_price_multiply_factor IS NOT NULL
                )),
    -- if rebalance is true, rebalance quadratic is required
    CONSTRAINT rebalance_quadratic_check CHECK (
            NOT rebalance OR
            (rebalance_size_quadratic_a IS NOT NULL AND
             rebalance_size_quadratic_b IS NOT NULL AND
             rebalance_size_quadratic_c IS NOT NULL AND
             rebalance_price_quadratic_a IS NOT NULL AND
             rebalance_price_quadratic_b IS NOT NULL AND
             rebalance_price_quadratic_c IS NOT NULL)),
    -- if rebalance is true, target configuration is required
    CONSTRAINT target_check CHECK (
            NOT rebalance OR
            (target_total IS NOT NULL AND
             target_reserve IS NOT NULL AND
             target_rebalance_threshold IS NOT NULL AND
             target_transfer_threshold IS NOT NULL))
);

CREATE TABLE IF NOT EXISTS "feed_weight"
(
    id SERIAL PRIMARY KEY,
    asset_id  INT NOT NULL REFERENCES assets (id),
    feed      TEXT NOT NULL,
    weight    FLOAT NOT NULL
);

CREATE TABLE IF NOT EXISTS "asset_old_addresses"
(
    id         SERIAL PRIMARY KEY,
    address_id INT NOT NULL REFERENCES addresses (id),
    asset_id   INT NOT NULL REFERENCES assets (id)
    -- TODO add a constraint to ensure that asset_id is not linked to any asset in address field already 
);

CREATE TABLE IF NOT EXISTS "asset_exchanges"
(
    id                 SERIAL PRIMARY KEY,
    exchange_id        INT REFERENCES exchanges (id) NOT NULL,
    asset_id           INT REFERENCES assets (id)    NOT NULL,
    symbol             TEXT                          NOT NULL,
    deposit_address    TEXT                          NULL,
    min_deposit        FLOAT                         NOT NULL,
    withdraw_fee       FLOAT                         NOT NULL,
    target_recommended FLOAT                         NOT NULL,
    target_ratio       FLOAT                         NOT NULL,
    UNIQUE (exchange_id, asset_id)
);

CREATE TABLE IF NOT EXISTS trading_pairs
(
    id               SERIAL PRIMARY KEY,
    exchange_id      INT REFERENCES exchanges (id) NOT NULL,
    base_id          INT REFERENCES assets (id)    NOT NULL,
    quote_id         INT REFERENCES assets (id)    NOT NULL,
    price_precision  INT                           NOT NULL,
    amount_precision INT                           NOT NULL,
    amount_limit_min FLOAT                         NOT NULL,
    amount_limit_max FLOAT                         NOT NULL,
    price_limit_min  FLOAT                         NOT NULL,
    price_limit_max  FLOAT                         NOT NULL,
    min_notional     FLOAT                         NOT NULL,
    UNIQUE (exchange_id, base_id, quote_id),
    CONSTRAINT trading_pair_check CHECK ( base_id != quote_id)
);
-- this table manage which asset will be use to buy/sell when trading.
CREATE TABLE IF NOT EXISTS trading_by
(
    id              SERIAL PRIMARY KEY,
    asset_id        INT REFERENCES assets (id)                          NOT NULL,
    trading_pair_id INT REFERENCES trading_pairs (id) ON DELETE CASCADE NOT NULL,
    UNIQUE (asset_id, trading_pair_id)
);

--create enum types if exist then alter 
DO
$$
    BEGIN
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'setting_change_cat') THEN
            CREATE TYPE setting_change_cat AS ENUM ('set_target', 'set_pwis',
                'set_stable_token','set_rebalance_quadratic', 'main', 'update_exchange', 'set_feed_configuration');
        END IF;
        IF NOT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'setting_change_status') THEN
            CREATE TYPE setting_change_status AS ENUM ('pending', 'accepted', 'rejected');
        END IF;
    END
$$;


CREATE TABLE IF NOT EXISTS setting_change
(
    id      SERIAL PRIMARY KEY,
    created TIMESTAMPTZ                 NOT NULL,
    cat     setting_change_cat NOT NULL,
    data    JSON                      NOT NULL,
    status  setting_change_status DEFAULT 'pending'
);

CREATE TABLE IF NOT EXISTS price_factor
(
    id        serial primary key,
    timepoint bigint NOT NULL,
    data      json   NOT NULL
);

CREATE TABLE IF NOT EXISTS set_rate_control
(
    id        SERIAL PRIMARY KEY,
    timepoint TIMESTAMPTZ NOT NULL,
    status    BOOLEAN   NOT NULL
);

CREATE TABLE IF NOT EXISTS rebalance_control
(
    id        SERIAL PRIMARY KEY,
    timepoint TIMESTAMPTZ NOT NULL,
    status    BOOLEAN   NOT NULL
);

CREATE TABLE IF NOT EXISTS stable_token_params_control
(
    id        SERIAL PRIMARY KEY,
    timepoint TIMESTAMPTZ NOT NULL,
    data      JSON      NOT NULL
);



CREATE OR REPLACE FUNCTION new_stable_token_params_control(_data stable_token_params_control.data%TYPE)
    RETURNS int AS
$$
DECLARE
    _id stable_token_params_control.id%TYPE;
BEGIN
    DELETE FROM stable_token_params_control;
    INSERT INTO stable_token_params_control(timepoint, data) VALUES (now(), _data) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION new_rebalance_control(_status rebalance_control.status%TYPE)
    RETURNS int AS
$$
DECLARE
    _id rebalance_control.id%TYPE;
BEGIN
    DELETE FROM rebalance_control;
    INSERT INTO rebalance_control(timepoint, status) VALUES (now(), _status) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION new_set_rate_control(_status set_rate_control.status%TYPE)
    RETURNS int AS
$$
DECLARE
    _id set_rate_control.id%TYPE;
BEGIN
    DELETE FROM set_rate_control;
    INSERT INTO set_rate_control(timepoint, status) VALUES (now(), _status) RETURNING id INTO _id;
    RETURN _id;
END

$$ LANGUAGE PLPGSQL;

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

CREATE OR REPLACE FUNCTION new_asset(_symbol assets.symbol%TYPE,
                                     _name assets.symbol%TYPE,
                                     _address addresses.address%TYPE,
                                     _decimals assets.decimals%TYPE,
                                     _transferable assets.transferable%TYPE,
                                     _set_rate assets.set_rate%TYPE,
                                     _rebalance assets.rebalance%TYPE,
                                     _is_quote assets.is_quote%TYPE,
                                     _is_enabled assets.is_enabled%TYPE,
                                     _pwi_ask_a assets.pwi_ask_a%TYPE,
                                     _pwi_ask_b assets.pwi_ask_b%TYPE,
                                     _pwi_ask_c assets.pwi_ask_c%TYPE,
                                     _pwi_ask_min_min_spread assets.pwi_ask_min_min_spread%TYPE,
                                     _pwi_ask_price_multiply_factor assets.pwi_ask_price_multiply_factor%TYPE,
                                     _pwi_bid_a assets.pwi_bid_a%TYPE,
                                     _pwi_bid_b assets.pwi_bid_b%TYPE,
                                     _pwi_bid_c assets.pwi_bid_c%TYPE,
                                     _pwi_bid_min_min_spread assets.pwi_bid_min_min_spread%TYPE,
                                     _pwi_bid_price_multiply_factor assets.pwi_bid_price_multiply_factor%TYPE,
                                     _rebalance_size_quadratic_a assets.rebalance_size_quadratic_a%TYPE,
                                     _rebalance_size_quadratic_b assets.rebalance_size_quadratic_b%TYPE,
                                     _rebalance_size_quadratic_c assets.rebalance_size_quadratic_c%TYPE,
                                     _rebalance_price_quadratic_a assets.rebalance_price_quadratic_a%TYPE,
                                     _rebalance_price_quadratic_b assets.rebalance_price_quadratic_b%TYPE,
                                     _rebalance_price_quadratic_c assets.rebalance_price_quadratic_c%TYPE,
                                     _target_total assets.target_total%TYPE,
                                     _target_reserve assets.target_reserve%TYPE,
                                     _target_rebalance_threshold assets.target_rebalance_threshold%TYPE,
                                     _target_transfer_threshold assets.target_total%TYPE,
									 _stable_param_price_update_threshold assets.stable_param_price_update_threshold%TYPE,
									 _stable_param_ask_spread assets.stable_param_ask_spread%TYPE,
									 _stable_param_bid_spread assets.stable_param_bid_spread%TYPE,
									 _stable_param_single_feed_max_spread assets.stable_param_single_feed_max_spread%TYPE,
                                     _stable_param_multiple_feeds_max_diff assets.stable_param_multiple_feeds_max_diff%TYPE,
                                     _normal_update_per_period assets.normal_update_per_period%TYPE,
                                     _max_imbalance_ratio assets.max_imbalance_ratio%TYPE
									)
    RETURNS int AS
$$
DECLARE
    _address_id addresses.id%TYPE;
    _id         assets.id%TYPE;
BEGIN
    IF _address IS NOT NULL THEN
        INSERT INTO "addresses" (address) VALUES (_address) RETURNING id INTO _address_id;
    END IF;

    INSERT
    INTO assets(symbol,
                name,
                address_id,
                decimals,
                transferable,
                set_rate,
                rebalance,
                is_quote,
                is_enabled,
                pwi_ask_a,
                pwi_ask_b,
                pwi_ask_c,
                pwi_ask_min_min_spread,
                pwi_ask_price_multiply_factor,
                pwi_bid_a,
                pwi_bid_b,
                pwi_bid_c,
                pwi_bid_min_min_spread,
                pwi_bid_price_multiply_factor,
                rebalance_size_quadratic_a,
                rebalance_size_quadratic_b,
                rebalance_size_quadratic_c,
                rebalance_price_quadratic_a,
                rebalance_price_quadratic_b,
                rebalance_price_quadratic_c,
                target_total,
                target_reserve,
                target_rebalance_threshold,
                target_transfer_threshold,
				stable_param_price_update_threshold,
				stable_param_ask_spread,
				stable_param_bid_spread,
				stable_param_single_feed_max_spread,
                stable_param_multiple_feeds_max_diff,
                normal_update_per_period,
                max_imbalance_ratio,
                created,
                updated)
    VALUES (_symbol,
            _name,
            _address_id,
            _decimals,
            _transferable,
            _set_rate,
            _rebalance,
            _is_quote,
            _is_enabled,
            _pwi_ask_a,
            _pwi_ask_b,
            _pwi_ask_c,
            _pwi_ask_min_min_spread,
            _pwi_ask_price_multiply_factor,
            _pwi_bid_a,
            _pwi_bid_b,
            _pwi_bid_c,
            _pwi_bid_min_min_spread,
            _pwi_bid_price_multiply_factor,
            _rebalance_size_quadratic_a,
            _rebalance_size_quadratic_b,
            _rebalance_size_quadratic_c,
            _rebalance_price_quadratic_a,
            _rebalance_price_quadratic_b,
            _rebalance_price_quadratic_c,
            _target_total,
            _target_reserve,
            _target_rebalance_threshold,
            _target_transfer_threshold,
			_stable_param_price_update_threshold,
			_stable_param_ask_spread,
			_stable_param_bid_spread,
			_stable_param_single_feed_max_spread,
            _stable_param_multiple_feeds_max_diff,
            _normal_update_per_period,
            _max_imbalance_ratio,
            now(),
            now()) RETURNING id INTO _id;

    RETURN _id;
END
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION change_asset_address(_id assets.id%TYPE, _address addresses.address%TYPE) RETURNS VOID AS
$$
DECLARE
    _new_address_id addresses.id%TYPE;
BEGIN
    PERFORM id FROM assets WHERE id = _id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'asset with id % does not exists', _id USING ERRCODE = 'assert_failure';
    END IF;

    INSERT INTO "asset_old_addresses" (address_id, asset_id)
    SELECT addresses.id, assets.id
    FROM assets
             LEFT JOIN addresses ON assets.address_id = addresses.id
    WHERE assets.id = _id;

    INSERT INTO "addresses" (address) VALUES (_address) RETURNING id INTO _new_address_id;

    UPDATE "assets"
    SET address_id = _new_address_id,
        updated    = now()
    WHERE assets.id = _id;
    RETURN;
END
$$ LANGUAGE PLPGSQL;

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

CREATE OR REPLACE FUNCTION new_trading_by(_asset_id assets.id%TYPE,
                                          _trading_pair_id trading_pairs.id%TYPE)
    RETURNS INT AS
$$
DECLARE
    _id trading_by.id%TYPE;
BEGIN
    PERFORM id FROM trading_pairs WHERE base_id = _asset_id OR quote_id = _asset_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'asset must be base or quote in trading pair, asset=%',
            _asset_id USING ERRCODE = 'assert_failure';
    END IF;

    INSERT INTO trading_by (asset_id, trading_pair_id)
    VALUES (_asset_id, _trading_pair_id) RETURNING id INTO _id;
    RETURN _id;
END
$$ LANGUAGE PLPGSQL;

CREATE OR REPLACE FUNCTION delete_asset_exchange(_asset_exchange_id asset_exchanges.id%TYPE)
    RETURNS INT AS
$$
DECLARE
    _id asset_exchanges.id%TYPE;
BEGIN
    PERFORM trading_pairs.id
    FROM "trading_pairs"
             INNER JOIN "asset_exchanges"
                        ON (trading_pairs.base_id = asset_exchanges.asset_id OR
                            trading_pairs.base_id = asset_exchanges.asset_id)
                            AND trading_pairs.exchange_id = asset_exchanges.exchange_id
    WHERE asset_exchanges.id = _asset_exchange_id;
    IF FOUND THEN
        RAISE EXCEPTION 'trading pair must be deleted before remove asset exchange, id=%',
            _asset_exchange_id USING ERRCODE = 'restrict_violation';
    END IF;

    DELETE FROM "asset_exchanges" WHERE id = _asset_exchange_id RETURNING id INTO _id;
    RETURN _id;
END
$$ LANGUAGE PLPGSQL;

CREATE TABLE IF NOT EXISTS "feed_configurations"
(
	name                   TEXT    NOT NULL,
	set_rate               TEXT    NOT NULL,
    enabled                BOOLEAN NOT NULL,
    base_volatility_spread FLOAT   DEFAULT 0,
    normal_spread          FLOAT   DEFAULT 0,
	PRIMARY KEY (name,set_rate)
);
`
