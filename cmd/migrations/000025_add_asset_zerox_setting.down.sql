ALTER TABLE assets DROP COLUMN ref_eth_amount;
ALTER TABLE assets DROP COLUMN eth_step;
ALTER TABLE assets DROP COLUMN max_eth_size_buy;
ALTER TABLE assets DROP COLUMN max_eth_size_sell;
ALTER TABLE assets DROP COLUMN zerox_enabled;

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
                                     _rebalance_price_offset assets.rebalance_price_offset%TYPE,
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
                                     _max_imbalance_ratio assets.max_imbalance_ratio%TYPE,
                                     _order_duration_millis assets.order_duration_millis%TYPE,
                                     _price_eth_amount assets.price_eth_amount%TYPE,
                                     _exchange_eth_amount assets.exchange_eth_amount%TYPE,
                                     _target_min_withdraw_threshold assets.target_min_withdraw_threshold%TYPE,
                                     _sanity_threshold assets.sanity_threshold%TYPE,
                                     _sanity_rate_provider assets.sanity_rate_provider%TYPE,
                                     _sanity_rate_path assets.sanity_rate_path%TYPE
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
                rebalance_price_offset,
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
                order_duration_millis,
                price_eth_amount,
                exchange_eth_amount,
                target_min_withdraw_threshold,
                sanity_threshold,
                sanity_rate_provider,
                sanity_rate_path,
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
            _rebalance_price_offset,
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
            _order_duration_millis,
            _price_eth_amount,
            _exchange_eth_amount,
            _target_min_withdraw_threshold,
            _sanity_threshold,
            _sanity_rate_provider,
            _sanity_rate_path,
            now(),
            now()) RETURNING id INTO _id;

    RETURN _id;
END
$$ LANGUAGE PLPGSQL;
