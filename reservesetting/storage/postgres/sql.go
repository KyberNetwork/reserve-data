package postgres

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type preparedStmts struct {
	getExchanges        *sqlx.Stmt
	getExchange         *sqlx.Stmt
	getExchangeByName   *sqlx.Stmt
	updateExchange      *sqlx.NamedStmt
	newAsset            *sqlx.NamedStmt
	newAssetExchange    *sqlx.NamedStmt
	updateAssetExchange *sqlx.NamedStmt
	deleteAssetExchange *sqlx.Stmt
	newTradingPair      *sqlx.NamedStmt
	newFeedWeight       *sqlx.NamedStmt
	getFeedWeight       *sqlx.Stmt
	deleteFeedWeight    *sqlx.Stmt

	getAsset                 *sqlx.Stmt
	getAssetBySymbol         *sqlx.Stmt
	getAssetExchange         *sqlx.NamedStmt
	getAssetExchangeBySymbol *sqlx.Stmt
	getTradingPair           *sqlx.Stmt
	updateAsset              *sqlx.NamedStmt
	changeAssetAddress       *sqlx.Stmt
	updateDepositAddress     *sqlx.Stmt
	updateTradingPair        *sqlx.NamedStmt

	deleteTradingPair     *sqlx.Stmt
	getTradingPairByID    *sqlx.Stmt
	getTradingPairSymbols *sqlx.Stmt
	getMinNotional        *sqlx.Stmt
	// getTransferableAssets *sqlx.Stmt
	newTradingBy    *sqlx.Stmt
	getTradingBy    *sqlx.Stmt
	deleteTradingBy *sqlx.Stmt

	newSettingChange          *sqlx.Stmt
	updateSettingChangeStatus *sqlx.Stmt
	getSettingChange          *sqlx.Stmt

	newPriceFactor      *sqlx.Stmt
	getPriceFactor      *sqlx.Stmt
	newSetRate          *sqlx.Stmt
	getSetRate          *sqlx.Stmt
	newRebalance        *sqlx.Stmt
	getRebalance        *sqlx.Stmt
	newStableTokenParam *sqlx.Stmt
	getStableTokenParam *sqlx.Stmt

	setFeedConfiguration  *sqlx.NamedStmt
	getFeedConfiguration  *sqlx.Stmt
	getFeedConfigurations *sqlx.Stmt

	getGeneralData *sqlx.Stmt
	setGeneralData *sqlx.NamedStmt
}

func newPreparedStmts(db *sqlx.DB) (*preparedStmts, error) {
	getExchanges, getExchange, getExchangeByName, updateExchange, err := exchangeStatements(db)
	if err != nil {
		return nil, err
	}

	newAsset, getAsset, updateAsset, getAssetBySymbol, err := assetStatements(db)
	if err != nil {
		return nil, err
	}

	newAssetExchange, updateAssetExchange, getAssetExchange, getAssetExchangeBySymbol, deleteAssetExchangeStmt, err := assetExchangeStatements(db)
	if err != nil {
		return nil, err
	}

	tradingPairStmts, err := tradingPairStatements(db)
	if err != nil {
		return nil, err
	}

	const changeAssetAddressQuery = `SELECT change_asset_address($1, $2);`
	changeAssetAddress, err := db.Preparex(changeAssetAddressQuery)
	if err != nil {
		return nil, err
	}

	const newFeedWeightQuery = `INSERT INTO feed_weight(asset_id, feed, weight)
									VALUES (:asset_id, :feed, :weight) RETURNING id;`
	newFeedWeight, err := db.PrepareNamed(newFeedWeightQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare newFeedWeight")
	}

	const getFeedWeightQuery = `SELECT id, asset_id, feed, weight FROM feed_weight
								WHERE asset_id = coalesce($1, asset_id)`
	getFeedWeight, err := db.Preparex(getFeedWeightQuery)
	if err != nil {
		return nil, err
	}

	const deleteFeedWeightQuery = `DELETE FROM feed_weight WHERE asset_id = $1;`
	deleteFeedWeight, err := db.Preparex(deleteFeedWeightQuery)
	if err != nil {
		return nil, err
	}

	const getMinNotionalQuery = `SELECT min_notional
									FROM trading_pairs
									WHERE exchange_id = $1
									  AND base_id = $2
									  AND quote_id = $3;
									`
	getMinNotional, err := db.Preparex(getMinNotionalQuery)
	if err != nil {
		return nil, err
	}

	const updateDepositAddressQuery = `UPDATE asset_exchanges
									SET deposit_address = $3
									WHERE asset_id = $1
									  AND exchange_id = $2 RETURNING id;`
	updateDepositAddress, err := db.Preparex(updateDepositAddressQuery)
	if err != nil {
		return nil, err
	}

	newTradingBy, getTradingBy, deleteTradingBy, err := tradingByStatements(db)
	if err != nil {
		return nil, err
	}

	newSettingChange, updateSettingChangeStatus, getSettingChange, err := settingChangeStatements(db)
	if err != nil {
		return nil, err
	}

	newPriceFactor, getPriceFactor, err := priceFactorStatements(db)
	if err != nil {
		return nil, err
	}

	newSetRate, getSetRate, err := setRateControlStatements(db)
	if err != nil {
		return nil, err
	}

	newRebalance, getRebalance, err := rebalanceControlStatements(db)
	if err != nil {
		return nil, err
	}

	newStabeTokenParams, getStableTokenParams, err := stableTokenParamsControlStatements(db)
	if err != nil {
		return nil, err
	}

	setFeedConfigurationStmt, getFeedConfigurationStmt, getFeedConfigurationsStmt, err := feedConfigurationStatements(db)
	if err != nil {
		return nil, err
	}

	setGeneralDataStmt, getGeneralDataStmt, err := generalDataStatements(db)
	if err != nil {
		return nil, err
	}

	return &preparedStmts{
		getExchanges:        getExchanges,
		getExchange:         getExchange,
		getExchangeByName:   getExchangeByName,
		updateExchange:      updateExchange,
		newAsset:            newAsset,
		newAssetExchange:    newAssetExchange,
		updateAssetExchange: updateAssetExchange,
		deleteAssetExchange: deleteAssetExchangeStmt,
		newFeedWeight:       newFeedWeight,
		getFeedWeight:       getFeedWeight,
		deleteFeedWeight:    deleteFeedWeight,

		newTradingPair:  tradingPairStmts.newStmt,
		newTradingBy:    newTradingBy,
		getTradingBy:    getTradingBy,
		deleteTradingBy: deleteTradingBy,

		getAsset:                 getAsset,
		getAssetBySymbol:         getAssetBySymbol,
		getAssetExchange:         getAssetExchange,
		getAssetExchangeBySymbol: getAssetExchangeBySymbol,
		getTradingPair:           tradingPairStmts.getStmt,
		updateAsset:              updateAsset,
		changeAssetAddress:       changeAssetAddress,
		updateDepositAddress:     updateDepositAddress,
		updateTradingPair:        tradingPairStmts.updateStmt,

		deleteTradingPair:     tradingPairStmts.deleteStmt,
		getTradingPairByID:    tradingPairStmts.getByIDStmt,
		getTradingPairSymbols: tradingPairStmts.getBySymbolStmt,
		getMinNotional:        getMinNotional,

		newSettingChange:          newSettingChange,
		updateSettingChangeStatus: updateSettingChangeStatus,
		getSettingChange:          getSettingChange,

		newPriceFactor:      newPriceFactor,
		getPriceFactor:      getPriceFactor,
		newSetRate:          newSetRate,
		getSetRate:          getSetRate,
		newRebalance:        newRebalance,
		getRebalance:        getRebalance,
		newStableTokenParam: newStabeTokenParams,
		getStableTokenParam: getStableTokenParams,

		setFeedConfiguration:  setFeedConfigurationStmt,
		getFeedConfiguration:  getFeedConfigurationStmt,
		getFeedConfigurations: getFeedConfigurationsStmt,

		getGeneralData: getGeneralDataStmt,
		setGeneralData: setGeneralDataStmt,
	}, nil
}

type tradingPairStmts struct {
	newStmt         *sqlx.NamedStmt
	getStmt         *sqlx.Stmt
	updateStmt      *sqlx.NamedStmt
	getByIDStmt     *sqlx.Stmt
	getBySymbolStmt *sqlx.Stmt
	deleteStmt      *sqlx.Stmt
}

func tradingPairStatements(db *sqlx.DB) (*tradingPairStmts, error) {
	const newTradingPairQuery = `SELECT new_trading_pair
									FROM new_trading_pair(:exchange_id,
									                      :base_id,
									                      :quote_id,
									                      :price_precision,
									                      :amount_precision,
									                      :amount_limit_min,
									                      :amount_limit_max,
									                      :price_limit_min,
									                      :price_limit_max,
									                      :min_notional);`
	newTradingPair, err := db.PrepareNamed(newTradingPairQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare newTradingPair")
	}
	const getTradingPairQuery = `SELECT DISTINCT tp.id,
									                tp.exchange_id,
									                tp.base_id,
									                tp.quote_id,
									                tp.price_precision,
									                tp.amount_precision,
									                tp.amount_limit_min,
									                tp.amount_limit_max,
									                tp.price_limit_min,
									                tp.price_limit_max,
									                tp.min_notional
									FROM trading_pairs tp
									         INNER JOIN asset_exchanges ae ON tp.exchange_id = ae.exchange_id
									WHERE ae.asset_id = coalesce($1, ae.asset_id);
									`
	getTradingPair, err := db.Preparex(getTradingPairQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare getTradingPair")
	}
	const updateTradingPairQuery = `UPDATE "trading_pairs"
									SET price_precision  = coalesce(:price_precision, price_precision),
									    amount_precision = coalesce(:amount_precision, amount_precision),
									    amount_limit_min = coalesce(:amount_limit_min, amount_limit_min),
									    amount_limit_max = coalesce(:amount_limit_max, amount_limit_max),
									    price_limit_min  = coalesce(:price_limit_min, price_limit_min),
									    price_limit_max  = coalesce(:price_limit_max, price_limit_max),
									    min_notional= coalesce(:min_notional, min_notional)
									WHERE id = :id RETURNING id; `
	updateTradingPair, err := db.PrepareNamed(updateTradingPairQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare updateTradingPair")
	}

	const getTradingPairByIDQuery = `WITH selected AS (
	SELECT tp.id,tp.exchange_id, tp.base_id,tp.quote_id, tp.price_precision, tp.amount_precision, tp.amount_limit_min,tp.amount_limit_max,
	tp.price_limit_min, tp.price_limit_max, tp.min_notional	FROM trading_pairs tp WHERE tp.id=$1
	UNION ALL SELECT tpd.id,tpd.exchange_id, tpd.base_id,tpd.quote_id, tpd.price_precision, tpd.amount_precision, tpd.amount_limit_min,tpd.amount_limit_max,
	tpd.price_limit_min, tpd.price_limit_max, tpd.min_notional FROM trading_pairs_deleted tpd WHERE tpd.id=$1 and $2 IS TRUE
) SELECT DISTINCT tp.id,
									                tp.exchange_id,
									                tp.base_id,
									                tp.quote_id,
									                tp.price_precision,
									                tp.amount_precision,
									                tp.amount_limit_min,
									                tp.amount_limit_max,
									                tp.price_limit_min,
									                tp.price_limit_max,
									                tp.min_notional,
									                bae.symbol AS base_symbol,
									                qae.symbol AS quote_symbol
									FROM selected AS tp
									         INNER JOIN assets AS ba ON tp.base_id = ba.id
									         INNER JOIN asset_exchanges AS bae ON ba.id = bae.asset_id
									         INNER JOIN assets AS qa ON tp.quote_id = qa.id
									         INNER JOIN asset_exchanges AS qae ON qa.id = qae.asset_id
									WHERE tp.exchange_id = bae.exchange_id AND tp.exchange_id = qae.exchange_id`
	getTradingPairByID, err := db.Preparex(getTradingPairByIDQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare getTradingPairByID")
	}

	const getTradingPairSymbolsQuery = `SELECT DISTINCT tp.id,
									                tp.exchange_id,
									                tp.base_id,
									                tp.quote_id,
									                tp.price_precision,
									                tp.amount_precision,
									                tp.amount_limit_min,
									                tp.amount_limit_max,
									                tp.price_limit_min,
									                tp.price_limit_max,
									                tp.min_notional,
									                bae.symbol AS base_symbol,
									                qae.symbol AS quote_symbol
									FROM trading_pairs AS tp
									         INNER JOIN assets AS ba ON tp.base_id = ba.id
									         INNER JOIN asset_exchanges AS bae ON ba.id = bae.asset_id
									         INNER JOIN assets AS qa ON tp.quote_id = qa.id
									         INNER JOIN asset_exchanges AS qae ON qa.id = qae.asset_id
									WHERE tp.exchange_id = $1 AND bae.exchange_id=tp.exchange_id and qae.exchange_id=tp.exchange_id;`
	getTradingPairSymbols, err := db.Preparex(getTradingPairSymbolsQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare getTradingPairSymbols")
	}

	const deleteTradingPairQuery = `WITH aa AS ( 
INSERT INTO trading_pairs_deleted
SELECT NOW() AS deleted_at,* FROM trading_pairs WHERE id=$1
) DELETE FROM trading_pairs WHERE id=$1 RETURNING id`
	deleteStmt, err := db.Preparex(deleteTradingPairQuery)
	if err != nil {
		return nil, errors.Wrap(err, "failed to prepare deleteTradingPairQuery")
	}

	return &tradingPairStmts{
		newStmt:         newTradingPair,
		getStmt:         getTradingPair,
		updateStmt:      updateTradingPair,
		getByIDStmt:     getTradingPairByID,
		getBySymbolStmt: getTradingPairSymbols,
		deleteStmt:      deleteStmt,
	}, nil
}

func assetStatements(db *sqlx.DB) (*sqlx.NamedStmt, *sqlx.Stmt, *sqlx.NamedStmt, *sqlx.Stmt, error) {
	const newAssetQuery = `SELECT new_asset
		FROM new_asset(
		             :symbol,
		             :name,
		             :address,
		             :decimals,
		             :transferable,
		             :set_rate,
		             :rebalance,
								 :is_quote,
								 :is_enabled,
		             :ask_a,
		             :ask_b,
		             :ask_c,
		             :ask_min_min_spread,
		             :ask_price_multiply_factor,
		             :bid_a,
		             :bid_b,
		             :bid_c,
		             :bid_min_min_spread,
		             :bid_price_multiply_factor,
		             :rebalance_size_quadratic_a,
		             :rebalance_size_quadratic_b,
		             :rebalance_size_quadratic_c,
		             :rebalance_price_quadratic_a,
		             :rebalance_price_quadratic_b,
		             :rebalance_price_quadratic_c,
		             :target_total,
		             :target_reserve,
		             :target_rebalance_threshold,
		             :target_transfer_threshold,
		    		 :stable_param_price_update_threshold,
					 :stable_param_ask_spread,
		    		 :stable_param_bid_spread,
		    		 :stable_param_single_feed_max_spread,
						 :stable_param_multiple_feeds_max_diff,
						 :normal_update_per_period,
						 :max_imbalance_ratio
		         );`
	newAsset, err := db.PrepareNamed(newAssetQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare newAsset")
	}
	const getAssetQuery = `SELECT assets.id,
								       assets.symbol,
								       assets.name,
								       a.address,
								       array_agg(oa.address) FILTER ( WHERE oa.address IS NOT NULL ) AS old_addresses,
								       assets.decimals,
								       assets.transferable,
								       assets.set_rate,
								       assets.rebalance,
								       assets.is_quote,
								       assets.pwi_ask_a,
								       assets.pwi_ask_b,
								       assets.pwi_ask_c,
								       assets.pwi_ask_min_min_spread,
								       assets.pwi_ask_price_multiply_factor,
								       assets.pwi_bid_a,
								       assets.pwi_bid_b,
								       assets.pwi_bid_c,
								       assets.pwi_bid_min_min_spread,
								       assets.pwi_bid_price_multiply_factor,
								       assets.rebalance_size_quadratic_a,
								       assets.rebalance_size_quadratic_b,
								       assets.rebalance_size_quadratic_c,
								       assets.rebalance_price_quadratic_a,
								       assets.rebalance_price_quadratic_b,
								       assets.rebalance_price_quadratic_c,
								       assets.target_total,
								       assets.target_reserve,
								       assets.target_rebalance_threshold,
								       assets.target_transfer_threshold,
       								   assets.stable_param_price_update_threshold,
       								   assets.stable_param_ask_spread,
									   assets.stable_param_bid_spread,
									   assets.stable_param_single_feed_max_spread,
										 assets.stable_param_multiple_feeds_max_diff,
										 assets.normal_update_per_period,
										 assets.max_imbalance_ratio,
								       assets.created,
								       assets.updated
								FROM assets
								         LEFT JOIN addresses a on assets.address_id = a.id
								         LEFT JOIN asset_old_addresses aoa on assets.id = aoa.asset_id
								         LEFT JOIN addresses oa ON aoa.address_id = oa.id
								WHERE assets.id = coalesce($1, assets.id)
								  AND assets.transferable = coalesce($2, assets.transferable)
								GROUP BY assets.id,
								         assets.symbol,
								         assets.name,
								         a.address,
								         assets.decimals,
								         assets.transferable,
								         assets.set_rate,
								         assets.rebalance,
								         assets.is_quote,
								         assets.pwi_ask_a,
								         assets.pwi_ask_b,
								         assets.pwi_ask_c,
								         assets.pwi_ask_min_min_spread,
								         assets.pwi_ask_price_multiply_factor,
								         assets.pwi_bid_a,
								         assets.pwi_bid_b,
								         assets.pwi_bid_c,
								         assets.pwi_bid_min_min_spread,
								         assets.pwi_bid_price_multiply_factor,
								         assets.rebalance_size_quadratic_a,
								         assets.rebalance_size_quadratic_b,
								         assets.rebalance_size_quadratic_c,
								         assets.rebalance_price_quadratic_a,
								         assets.rebalance_price_quadratic_b,
								         assets.rebalance_price_quadratic_c,
								         assets.target_total,
								         assets.target_reserve,
								         assets.target_rebalance_threshold,
								         assets.target_transfer_threshold,
								         assets.stable_param_price_update_threshold,
       								   	 assets.stable_param_ask_spread,
									     assets.stable_param_bid_spread,
									     assets.stable_param_single_feed_max_spread,
									     assets.stable_param_multiple_feeds_max_diff,
								         assets.created,
								         assets.updated
								ORDER BY assets.id`
	getAsset, err := db.Preparex(getAssetQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare getAsset")
	}

	getAssetBySymbolQuery := `SELECT id, decimals FROM assets WHERE symbol = $1`
	getAssetBySymbol, err := db.Preparex(getAssetBySymbolQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare getAssetBySymbolQuery")
	}

	const updateAssetQuery = `WITH updated AS (
			UPDATE "addresses"
				SET address = COALESCE(:address, addresses.address)
				FROM "assets"
				WHERE assets.id = :id AND assets.address_id = addresses.id
			)
		UPDATE "assets"
		SET symbol       = COALESCE(:symbol, symbol),
			name         = COALESCE(:name, name),
			decimals     = COALESCE(:decimals, decimals),
			transferable = COALESCE(:transferable, transferable),
			set_rate     = COALESCE(:set_rate, set_rate),
			rebalance    = COALESCE(:rebalance, rebalance),
			is_quote     = COALESCE(:is_quote, is_quote),
			is_enabled   = COALESCE(:is_enabled, is_enabled),
			pwi_ask_a = COALESCE(:ask_a,pwi_ask_a),
			pwi_ask_b = COALESCE(:ask_b, pwi_ask_b),
			pwi_ask_c = COALESCE(:ask_c, pwi_ask_c),
			pwi_ask_min_min_spread = COALESCE(:ask_min_min_spread,pwi_ask_min_min_spread),
			pwi_ask_price_multiply_factor = COALESCE(:ask_price_multiply_factor, pwi_ask_price_multiply_factor),
			pwi_bid_a = COALESCE(:bid_a,pwi_bid_a),
			pwi_bid_b = COALESCE(:bid_b,pwi_bid_b),
			pwi_bid_c = COALESCE(:bid_c,pwi_bid_c),
			pwi_bid_min_min_spread = COALESCE(:bid_min_min_spread,pwi_bid_min_min_spread),
			pwi_bid_price_multiply_factor = COALESCE(:bid_price_multiply_factor,pwi_bid_price_multiply_factor),
			rebalance_size_quadratic_a = COALESCE(:rebalance_size_quadratic_a,rebalance_size_quadratic_a),
			rebalance_size_quadratic_b = COALESCE(:rebalance_size_quadratic_b,rebalance_size_quadratic_b),
			rebalance_size_quadratic_c = COALESCE(:rebalance_size_quadratic_c,rebalance_size_quadratic_c),
			rebalance_price_quadratic_a = COALESCE(:rebalance_price_quadratic_a,rebalance_price_quadratic_a),
			rebalance_price_quadratic_b = COALESCE(:rebalance_price_quadratic_b,rebalance_price_quadratic_b),
			rebalance_price_quadratic_c = COALESCE(:rebalance_price_quadratic_c,rebalance_price_quadratic_c),
			target_total = COALESCE(:target_total,target_total),
			target_reserve = COALESCE(:target_reserve,target_reserve),
			target_rebalance_threshold = COALESCE(:target_rebalance_threshold,target_rebalance_threshold),
			target_transfer_threshold = COALESCE(:target_transfer_threshold,target_transfer_threshold),
			stable_param_price_update_threshold = COALESCE(:stable_param_price_update_threshold,stable_param_price_update_threshold),
		    stable_param_ask_spread = COALESCE(:stable_param_ask_spread,stable_param_ask_spread),
		    stable_param_bid_spread = COALESCE(:stable_param_bid_spread,stable_param_bid_spread),
		    stable_param_single_feed_max_spread = COALESCE(:stable_param_single_feed_max_spread,stable_param_single_feed_max_spread),
				stable_param_multiple_feeds_max_diff = COALESCE(:stable_param_multiple_feeds_max_diff,stable_param_multiple_feeds_max_diff),
				normal_update_per_period = COALESCE(:normal_update_per_period,normal_update_per_period),
				max_imbalance_ratio = COALESCE(:max_imbalance_ratio,max_imbalance_ratio),
		    updated      = now()
		WHERE id = :id RETURNING id;
		`
	updateAsset, err := db.PrepareNamed(updateAssetQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare updateAsset")
	}
	return newAsset, getAsset, updateAsset, getAssetBySymbol, nil
}

func assetExchangeStatements(db *sqlx.DB) (*sqlx.NamedStmt, *sqlx.NamedStmt, *sqlx.NamedStmt, *sqlx.Stmt, *sqlx.Stmt, error) {
	const newAssetExchangeQuery string = `INSERT INTO asset_exchanges(exchange_id,
		                            asset_id,
		                            symbol,
		                            deposit_address,
		                            min_deposit,
		                            withdraw_fee,
		                            target_recommended,
		                            target_ratio)
		VALUES (:exchange_id,
		        :asset_id,
		        :symbol,
		        :deposit_address,
		        :min_deposit,
		        :withdraw_fee,
		        :target_recommended,
		        :target_ratio) RETURNING id`
	newAssetExchange, err := db.PrepareNamed(newAssetExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "failed to prepare newAssetExchange")
	}
	const updateAssetExchangeQuery string = `UPDATE "asset_exchanges"
		SET symbol = COALESCE(:symbol, symbol),
		    deposit_address = COALESCE(:deposit_address, deposit_address),
		    min_deposit           = COALESCE(:min_deposit, min_deposit),
			withdraw_fee = coalesce(:withdraw_fee, withdraw_fee),
		    target_recommended = coalesce(:target_recommended,target_recommended),
		    target_ratio = coalesce(:target_ratio, target_ratio)
		WHERE id = :id RETURNING id;`
	updateAssetExchange, err := db.PrepareNamed(updateAssetExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "failed to prepare updateAssetExchange")
	}

	const getAssetExchangeQuery = `SELECT id,
			       exchange_id,
			       asset_id,
			       symbol,
			       deposit_address,
			       min_deposit,
			       withdraw_fee,
			       target_recommended,
			       target_ratio
			FROM asset_exchanges
			WHERE asset_id = coalesce(:asset_id, asset_id)
			AND id = coalesce(:id, id)
			AND exchange_id= coalesce(:exchange_id, exchange_id)`
	getAssetExchange, err := db.PrepareNamed(getAssetExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "failed to prepare getAssetExchange")
	}

	const getAssetExchangeBySymbolQuery = `SELECT
		asset_exchanges.asset_id as id,
		asset_exchanges.symbol as symbol,
		a.decimals as decimals	
	FROM asset_exchanges
		LEFT JOIN assets a on asset_exchanges.asset_id = a.id
	WHERE asset_exchanges.exchange_id = $1
	AND asset_exchanges.symbol= $2`
	getAssetExchangeBySymbol, err := db.Preparex(getAssetExchangeBySymbolQuery)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "failed to prepare getAssetExchangeBySymbol")
	}

	const deleteAssetExchangeQuery = `SELECT * FROM delete_asset_exchange($1)`
	deleteAssetExchangeStmt, err := db.Preparex(deleteAssetExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, nil, errors.Wrap(err, "failed to prepare deleteAssetExchangeStmt")
	}

	return newAssetExchange, updateAssetExchange, getAssetExchange,
		getAssetExchangeBySymbol, deleteAssetExchangeStmt, nil
}

func exchangeStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, *sqlx.Stmt, *sqlx.NamedStmt, error) {
	const getExchangesQuery = `SELECT * FROM "exchanges";`
	getExchanges, err := db.Preparex(getExchangesQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare get exchanges")
	}
	const getExchangeQuery = `SELECT * FROM "exchanges" WHERE id = $1`
	getExchange, err := db.Preparex(getExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare get exchange")
	}
	const getExchangeByNameQuery = `SELECT * FROM "exchanges" WHERE name = $1`
	getExchangeByName, err := db.Preparex(getExchangeByNameQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare get exchange by name")
	}
	const updateExchangeQuery = `UPDATE "exchanges"
	SET trading_fee_maker = COALESCE(:trading_fee_maker, trading_fee_maker),
	    trading_fee_taker = COALESCE(:trading_fee_taker, trading_fee_taker),
	    disable           = COALESCE(:disable, disable)
	WHERE id = :id RETURNING id
	`
	updateExchange, err := db.PrepareNamed(updateExchangeQuery)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrap(err, "failed to prepare update exchange")
	}
	return getExchanges, getExchange, getExchangeByName, updateExchange, nil
}

func tradingByStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, *sqlx.Stmt, error) {
	const createTradingByQuery = `SELECT new_trading_by FROM new_trading_by($1,$2);`
	tradingBy, err := db.Preparex(createTradingByQuery)
	if err != nil {
		return nil, nil, nil, err
	}

	const getTradingByQuery = `SELECT id,asset_id,trading_pair_id FROM trading_by WHERE id=COALESCE($1,trading_by.id)`
	getTradingByPairs, err := db.Preparex(getTradingByQuery)
	if err != nil {
		return nil, nil, nil, err
	}

	const deleteTradingByQuery = `DELETE FROM trading_by WHERE id = $1 RETURNING id`
	deleteTradingByStmt, err := db.Preparex(deleteTradingByQuery)
	if err != nil {
		return nil, nil, nil, err
	}
	return tradingBy, getTradingByPairs, deleteTradingByStmt, nil
}

func settingChangeStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, *sqlx.Stmt, error) {
	const newSettingChangeQuery = `SELECT new_setting_change FROM new_setting_change($1, $2)`
	newSettingChangeStmt, err := db.Preparex(newSettingChangeQuery)
	if err != nil {
		return nil, nil, nil, err
	}
	const updateSettingChangeStatus = `UPDATE setting_change SET status = $2 WHERE id=$1 returning id`
	updateSettingChangeStatusStmt, err := db.Preparex(updateSettingChangeStatus)
	if err != nil {
		return nil, nil, nil, err
	}
	const listSettingChangeQuery = `SELECT id,created,data FROM setting_change WHERE id=COALESCE($1, setting_change.id) AND cat=COALESCE($2, setting_change.cat)
	AND status=COALESCE($3, 'pending'::setting_change_status)`
	listSettingChangeStmt, err := db.Preparex(listSettingChangeQuery)
	if err != nil {
		return nil, nil, nil, err
	}
	return newSettingChangeStmt, updateSettingChangeStatusStmt, listSettingChangeStmt, nil
}

func priceFactorStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, error) {
	const newPriceFactorQuery = `INSERT INTO price_factor(timepoint,data) VALUES ($1,$2) RETURNING id;`
	newPriceFactorStmt, err := db.Preparex(newPriceFactorQuery)
	if err != nil {
		return nil, nil, err
	}
	const listSettingChangeQuery = `SELECT id,timepoint,data FROM price_factor WHERE $1 <= timepoint AND timepoint <= $2`
	listSettingChangeStmt, err := db.Preparex(listSettingChangeQuery)
	if err != nil {
		return nil, nil, err
	}
	return newPriceFactorStmt, listSettingChangeStmt, nil
}

func setRateControlStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, error) {
	const newSetRateQuery = `SELECT FROM new_set_rate_control($1);`
	newSetRateStmt, err := db.Preparex(newSetRateQuery)
	if err != nil {
		return nil, nil, err
	}
	const getSetRateQuery = `SELECT id,timepoint,status FROM set_rate_control`
	getSetRateStmt, err := db.Preparex(getSetRateQuery)
	if err != nil {
		return nil, nil, err
	}
	return newSetRateStmt, getSetRateStmt, nil
}

func rebalanceControlStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, error) {
	const newRebalanceQuery = `SELECT FROM new_rebalance_control($1);`
	newRebalanceStmt, err := db.Preparex(newRebalanceQuery)
	if err != nil {
		return nil, nil, err
	}
	const getRebalanceQuery = `SELECT id,timepoint,status FROM rebalance_control ORDER BY timepoint DESC`
	getRebalanceStmt, err := db.Preparex(getRebalanceQuery)
	if err != nil {
		return nil, nil, err
	}
	return newRebalanceStmt, getRebalanceStmt, nil
}

func stableTokenParamsControlStatements(db *sqlx.DB) (*sqlx.Stmt, *sqlx.Stmt, error) {
	const newStableTokenQuery = `SELECT FROM new_stable_token_params_control($1);`
	newStableTokenStmt, err := db.Preparex(newStableTokenQuery)
	if err != nil {
		return nil, nil, err
	}
	const getStableTokenQuery = `SELECT id,timepoint,data FROM stable_token_params_control ORDER BY timepoint DESC`
	getStableTokenStmt, err := db.Preparex(getStableTokenQuery)
	if err != nil {
		return nil, nil, err
	}
	return newStableTokenStmt, getStableTokenStmt, nil
}

func feedConfigurationStatements(db *sqlx.DB) (*sqlx.NamedStmt, *sqlx.Stmt, *sqlx.Stmt, error) {
	const setFeedConfiguration = `UPDATE "feed_configurations"
	SET enabled                = COALESCE(:enabled, enabled),
	    base_volatility_spread = COALESCE(:base_volatility_spread, base_volatility_spread),
	    normal_spread          = COALESCE(:normal_spread, normal_spread)
	WHERE name = :name AND set_rate = :set_rate RETURNING name;
	`
	setFeedConfigurationStmt, err := db.PrepareNamed(setFeedConfiguration)
	if err != nil {
		return nil, nil, nil, err
	}
	const getFeedConfigurations = `SELECT name, set_rate, enabled, base_volatility_spread, normal_spread FROM feed_configurations;`
	getFeedConfigurationsStmt, err := db.Preparex(getFeedConfigurations)
	if err != nil {
		return nil, nil, nil, err
	}
	const getFeedConfiguration = `SELECT name, set_rate, enabled, base_volatility_spread, normal_spread FROM feed_configurations WHERE name = $1 AND set_rate = $2;`
	getFeedConfigurationStmt, err := db.Preparex(getFeedConfiguration)
	if err != nil {
		return nil, nil, nil, err
	}
	return setFeedConfigurationStmt, getFeedConfigurationStmt, getFeedConfigurationsStmt, nil
}

func generalDataStatements(db *sqlx.DB) (*sqlx.NamedStmt, *sqlx.Stmt, error) {
	const setQuery = `INSERT INTO general_data(key, value, timestamp) VALUES (:key, :value, now()) RETURNING id;`
	setGeneralDataStmt, err := db.PrepareNamed(setQuery)
	if err != nil {
		return nil, nil, err
	}
	const getQuery = `SELECT key, value FROM general_data WHERE key=$1 ORDER BY timestamp DESC LIMIT 1;`
	getGeneralDataStmt, err := db.Preparex(getQuery)
	if err != nil {
		return nil, nil, err
	}
	return setGeneralDataStmt, getGeneralDataStmt, err
}
