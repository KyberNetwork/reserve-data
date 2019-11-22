package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	pgutil "github.com/KyberNetwork/reserve-data/common/postgres"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

const (
	addressesUniqueConstraint     = "addresses_address_key"
	exchangeForeignKeyConstraint  = "asset_exchanges_exchange_id_fkey"
	assetForeignKeyConstraint     = "asset_exchanges_asset_id_fkey"
	exchangeAssetUniqueConstraint = "asset_exchanges_exchange_id_asset_id_key"
)

type createAssetParams struct {
	Symbol       string  `db:"symbol"`
	Name         string  `db:"name"`
	Address      *string `db:"address"`
	Decimals     uint64  `db:"decimals"`
	Transferable bool    `db:"transferable"`
	SetRate      string  `db:"set_rate"`
	Rebalance    bool    `db:"rebalance"`
	IsQuote      bool    `db:"is_quote"`

	AskA                   *float64 `db:"ask_a"`
	AskB                   *float64 `db:"ask_b"`
	AskC                   *float64 `db:"ask_c"`
	AskMinMinSpread        *float64 `db:"ask_min_min_spread"`
	AskPriceMultiplyFactor *float64 `db:"ask_price_multiply_factor"`
	BidA                   *float64 `db:"bid_a"`
	BidB                   *float64 `db:"bid_b"`
	BidC                   *float64 `db:"bid_c"`
	BidMinMinSpread        *float64 `db:"bid_min_min_spread"`
	BidPriceMultiplyFactor *float64 `db:"bid_price_multiply_factor"`

	RebalanceQuadraticA *float64 `db:"rebalance_quadratic_a"`
	RebalanceQuadraticB *float64 `db:"rebalance_quadratic_b"`
	RebalanceQuadraticC *float64 `db:"rebalance_quadratic_c"`

	TargetTotal              *float64 `db:"target_total"`
	TargetReserve            *float64 `db:"target_reserve"`
	TargetRebalanceThreshold *float64 `db:"target_rebalance_threshold"`
	TargetTransferThreshold  *float64 `db:"target_transfer_threshold"`

	PriceUpdateThreshold float64 `db:"stable_param_price_update_threshold"`
	AskSpread            float64 `db:"stable_param_ask_spread"`
	BidSpread            float64 `db:"stable_param_bid_spread"`
	SingleFeedMaxSpread  float64 `db:"stable_param_single_feed_max_spread"`
	MultipleFeedsMaxDiff float64 `db:"stable_param_multiple_feeds_max_diff"`
}

// CreateAsset create a new asset
func (s *Storage) CreateAsset(
	symbol, name string,
	address ethereum.Address,
	decimals uint64,
	transferable bool,
	setRate common.SetRate,
	rebalance, isQuote bool,
	pwi *common.AssetPWI,
	rb *common.RebalanceQuadratic,
	exchanges []common.AssetExchange,
	target *common.AssetTarget,
	stableParam *common.StableParam,
	feedWeight *common.FeedWeight,
) (uint64, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	id, err := s.createAsset(tx, symbol, name, address, decimals, transferable,
		setRate, rebalance, isQuote, pwi, rb, exchanges, target, stableParam, feedWeight)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// CreateAssetExchange create a new asset exchange (asset support by exchange)
func (s *Storage) CreateAssetExchange(exchangeID, assetID uint64, symbol string, depositAddress ethereum.Address,
	minDeposit, withdrawFee, targetRecommended, targetRatio float64, tps []common.TradingPair) (uint64, error) {

	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)
	id, err := s.createAssetExchange(tx, exchangeID, assetID, symbol, depositAddress, minDeposit, withdrawFee,
		targetRecommended, targetRatio, tps)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

func (s *Storage) createAssetExchange(tx *sqlx.Tx, exchangeID, assetID uint64, symbol string,
	depositAddress ethereum.Address, minDeposit, withdrawFee, targetRecommended, targetRatio float64, tps []common.TradingPair) (uint64, error) {
	var assetExchangeID uint64
	var depositAddressParam *string

	// TODO: validate depositAddress, require if transferable

	if !common.IsZeroAddress(depositAddress) {
		depositAddressHex := depositAddress.String()
		depositAddressParam = &depositAddressHex
	}
	err := tx.NamedStmt(s.stmts.newAssetExchange).Get(&assetExchangeID, struct {
		ExchangeID        uint64  `db:"exchange_id"`
		AssetID           uint64  `db:"asset_id"`
		Symbol            string  `db:"symbol"`
		DepositAddress    *string `db:"deposit_address"`
		MinDeposit        float64 `db:"min_deposit"`
		WithdrawFee       float64 `db:"withdraw_fee"`
		TargetRecommended float64 `db:"target_recommended"`
		TargetRatio       float64 `db:"target_ratio"`
	}{
		ExchangeID:        exchangeID,
		AssetID:           assetID,
		Symbol:            symbol,
		DepositAddress:    depositAddressParam,
		MinDeposit:        minDeposit,
		WithdrawFee:       withdrawFee,
		TargetRecommended: targetRecommended,
		TargetRatio:       targetRatio,
	})

	if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown returned err=%s", err.Error())
		}
		s.l.Infow("failed to create new asset exchange", "err", pErr.Message)
		switch pErr.Code {
		case errForeignKeyViolation:
			switch pErr.Constraint {
			case exchangeForeignKeyConstraint:
				return 0, common.ErrExchangeNotExists
			case assetForeignKeyConstraint:
				return 0, common.ErrAssetNotExists
			}
		case errCodeUniqueViolation:
			if pErr.Constraint == exchangeAssetUniqueConstraint {
				return 0, common.ErrAssetExchangeAlreadyExist
			}
		}
	}
	for _, tradingPair := range tps {
		_, err = s.createTradingPair(tx, exchangeID,
			tradingPair.Base,
			tradingPair.Quote,
			tradingPair.PricePrecision,
			tradingPair.AmountPrecision,
			tradingPair.AmountLimitMin,
			tradingPair.AmountLimitMax,
			tradingPair.PriceLimitMin,
			tradingPair.PriceLimitMax,
			tradingPair.MinNotional,
			assetID,
		)
		if err != nil {
			s.l.Infow("failed to create trading pair", "err", err)
			return 0, err
		}
	}
	return assetExchangeID, err
}

func (s *Storage) updateAssetExchange(tx *sqlx.Tx, id uint64, updateOpts storage.UpdateAssetExchangeOpts) error {
	var (
		addressParam *string
	)

	var updateMsgs []string
	if updateOpts.Symbol != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("symbol=%s", *updateOpts.Symbol))
	}
	if updateOpts.DepositAddress != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("deposit_address=%s", updateOpts.DepositAddress))
		addressStr := updateOpts.DepositAddress.String()
		addressParam = &addressStr
	}
	if updateOpts.MinDeposit != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("min_deposit=%f", *updateOpts.MinDeposit))
	}
	if updateOpts.WithdrawFee != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("withdraw_fee=%f", *updateOpts.WithdrawFee))
	}
	if updateOpts.TargetRecommended != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("target_recommended=%f", *updateOpts.TargetRecommended))
	}
	if updateOpts.TargetRatio != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("target_ratio=%f", *updateOpts.TargetRatio))
	}

	if len(updateMsgs) == 0 {
		s.l.Infow("nothing set for update asset exchange, skip now")
		return nil
	}

	s.l.Infow("updating asset_exchange", "id", id, "fields", strings.Join(updateMsgs, " "))
	var updatedID uint64
	err := s.stmts.updateAssetExchange.Get(&updatedID,
		struct {
			ID                uint64   `db:"id"`
			Symbol            *string  `db:"symbol"`
			DepositAddress    *string  `db:"deposit_address"`
			MinDeposit        *float64 `db:"min_deposit"`
			WithdrawFee       *float64 `db:"withdraw_fee"`
			TargetRecommended *float64 `db:"target_recommended"`
			TargetRatio       *float64 `db:"target_ratio"`
		}{
			ID:                id,
			Symbol:            updateOpts.Symbol,
			DepositAddress:    addressParam,
			MinDeposit:        updateOpts.MinDeposit,
			WithdrawFee:       updateOpts.WithdrawFee,
			TargetRecommended: updateOpts.TargetRecommended,
			TargetRatio:       updateOpts.TargetRatio,
		},
	)
	if err == sql.ErrNoRows {
		s.l.Infow("asset_exchange not found in database", "id", id)
		return common.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("failed to update asset_exchange, err=%s", err)
	}
	return nil
}

func (s *Storage) createAsset(
	tx *sqlx.Tx,
	symbol, name string,
	address ethereum.Address,
	decimals uint64,
	transferable bool,
	setRate common.SetRate,
	rebalance, isQuote bool,
	pwi *common.AssetPWI,
	rb *common.RebalanceQuadratic,
	exchanges []common.AssetExchange,
	target *common.AssetTarget,
	stableParam *common.StableParam,
	feedWeight *common.FeedWeight,
) (uint64, error) {
	// create new asset
	var assetID uint64

	if transferable && common.IsZeroAddress(address) {
		return 0, common.ErrAddressMissing
	}

	for _, exchange := range exchanges {
		if transferable && common.IsZeroAddress(exchange.DepositAddress) {
			return 0, common.ErrDepositAddressMissing
		}
	}

	s.l.Infow("creating new asset", "symbol", symbol, "address", address.String())

	var addressParam *string
	if !common.IsZeroAddress(address) {
		addressHex := address.String()
		addressParam = &addressHex
	}
	arg := createAssetParams{
		Symbol:       symbol,
		Name:         name,
		Address:      addressParam,
		Decimals:     decimals,
		Transferable: transferable,
		SetRate:      setRate.String(),
		Rebalance:    rebalance,
		IsQuote:      isQuote,
	}

	if pwi != nil {
		arg.AskA = &pwi.Ask.A
		arg.AskB = &pwi.Ask.B
		arg.AskC = &pwi.Ask.C
		arg.AskMinMinSpread = &pwi.Ask.MinMinSpread
		arg.AskPriceMultiplyFactor = &pwi.Ask.PriceMultiplyFactor
		arg.BidA = &pwi.Bid.A
		arg.BidB = &pwi.Bid.B
		arg.BidC = &pwi.Bid.C
		arg.BidMinMinSpread = &pwi.Bid.MinMinSpread
		arg.BidPriceMultiplyFactor = &pwi.Bid.PriceMultiplyFactor
	}

	if rebalance {
		if rb == nil {
			s.l.Infow("rebalance is enabled but rebalance quadratic is invalid", "symbol", symbol)
			return 0, common.ErrRebalanceQuadraticMissing
		}

		if len(exchanges) == 0 {
			s.l.Infow("rebalance is enabled but no exchange configuration is provided", "symbol", symbol)
			return 0, common.ErrAssetExchangeMissing
		}

		if target == nil {
			s.l.Infow("rebalance is enabled but target configuration is invalid", "symbol", symbol)
			return 0, common.ErrAssetTargetMissing
		}
	}

	if rb != nil {
		arg.RebalanceQuadraticA = &rb.A
		arg.RebalanceQuadraticB = &rb.B
		arg.RebalanceQuadraticC = &rb.C
	}

	if target != nil {
		arg.TargetTotal = &target.Total
		arg.TargetReserve = &target.Reserve
		arg.TargetRebalanceThreshold = &target.RebalanceThreshold
		arg.TargetTransferThreshold = &target.TransferThreshold
	}

	if stableParam != nil {
		arg.PriceUpdateThreshold = stableParam.PriceUpdateThreshold
		arg.AskSpread = stableParam.AskSpread
		arg.BidSpread = stableParam.BidSpread
		arg.SingleFeedMaxSpread = stableParam.SingleFeedMaxSpread
		arg.MultipleFeedsMaxDiff = stableParam.MultipleFeedsMaxDiff
	}

	if err := tx.NamedStmt(s.stmts.newAsset).Get(&assetID, arg); err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown returned err=%s", err.Error())
		}

		s.l.Infow("failed to create new asset", "err", pErr.Message)
		switch pErr.Code {
		case errCodeUniqueViolation:
			switch pErr.Constraint {
			case addressesUniqueConstraint:
				return 0, common.ErrAddressExists
			case "assets_symbol_key":
				return 0, common.ErrSymbolExists
			}
		case errCodeCheckViolation:
			if pErr.Constraint == "pwi_check" {
				return 0, common.ErrPWIMissing
			}
		}
		return 0, pErr
	}
	// create new asset exchange
	for _, exchange := range exchanges {
		var (
			assetExchangeID     uint64
			depositAddressParam *string
		)
		if !common.IsZeroAddress(exchange.DepositAddress) {
			depositAddressHex := exchange.DepositAddress.String()
			depositAddressParam = &depositAddressHex
		}
		err := tx.NamedStmt(s.stmts.newAssetExchange).Get(&assetExchangeID, struct {
			ExchangeID        uint64  `db:"exchange_id"`
			AssetID           uint64  `db:"asset_id"`
			Symbol            string  `db:"symbol"`
			DepositAddress    *string `db:"deposit_address"`
			MinDeposit        float64 `db:"min_deposit"`
			WithdrawFee       float64 `db:"withdraw_fee"`
			TargetRecommended float64 `db:"target_recommended"`
			TargetRatio       float64 `db:"target_ratio"`
		}{
			ExchangeID:        exchange.ExchangeID,
			AssetID:           assetID,
			Symbol:            exchange.Symbol,
			DepositAddress:    depositAddressParam,
			MinDeposit:        exchange.MinDeposit,
			WithdrawFee:       exchange.WithdrawFee,
			TargetRecommended: exchange.TargetRecommended,
			TargetRatio:       exchange.TargetRatio,
		})

		if err != nil {
			return 0, err
		}

		s.l.Infow("asset exchange is created", "id", assetExchangeID)
		// create new trading pair
		for _, pair := range exchange.TradingPairs {
			var (
				tradingPairID uint64
				baseID        = pair.Base
				quoteID       = pair.Quote
			)

			if baseID != 0 && quoteID != 0 {
				s.l.Infow("both base and quote are provided asset_symbol=%s exchange_id=%d", symbol, exchange.ExchangeID)
				return 0, common.ErrBadTradingPairConfiguration
			}

			if baseID == 0 {
				baseID = assetID
			}
			if quoteID == 0 {
				quoteID = assetID
			}

			err = tx.NamedStmt(s.stmts.newTradingPair).Get(
				&tradingPairID,
				struct {
					ExchangeID      uint64  `db:"exchange_id"`
					Base            uint64  `db:"base_id"`
					Quote           uint64  `db:"quote_id"`
					PricePrecision  uint64  `db:"price_precision"`
					AmountPrecision uint64  `db:"amount_precision"`
					AmountLimitMin  float64 `db:"amount_limit_min"`
					AmountLimitMax  float64 `db:"amount_limit_max"`
					PriceLimitMin   float64 `db:"price_limit_min"`
					PriceLimitMax   float64 `db:"price_limit_max"`
					MinNotional     float64 `db:"min_notional"`
				}{
					ExchangeID:      exchange.ExchangeID,
					Base:            baseID,
					Quote:           quoteID,
					PricePrecision:  pair.PricePrecision,
					AmountPrecision: pair.AmountPrecision,
					AmountLimitMin:  pair.AmountLimitMin,
					AmountLimitMax:  pair.AmountLimitMax,
					PriceLimitMin:   pair.PriceLimitMin,
					PriceLimitMax:   pair.PriceLimitMax,
					MinNotional:     pair.MinNotional,
				},
			)
			if err != nil {
				pErr, ok := err.(*pq.Error)
				if !ok {
					return 0, fmt.Errorf("unknown error returned err=%s", err.Error())
				}

				switch pErr.Code {
				case errForeignKeyViolation:
					s.l.Infow("failed to create trading pair as assertion failed", "symbol", symbol,
						"exchange_id", exchange.ExchangeID, "err", pErr.Message)
					return 0, common.ErrBadTradingPairConfiguration
				case errBaseInvalid:
					s.l.Infow("failed to create trading pair as check base failed", "symbol", symbol,
						"exchange_id", exchange.ExchangeID, "err", pErr.Message)
					return 0, common.ErrBaseAssetInvalid
				case errQuoteInvalid:
					s.l.Infow("failed to create trading pair as check quote failed", "symbol", symbol,
						"exchange_id", exchange.ExchangeID, "err", pErr.Message)
					return 0, common.ErrQuoteAssetInvalid
				}

				return 0, fmt.Errorf("failed to create trading pair symbol=%s exchange_id=%d err=%s",
					symbol,
					exchange.ExchangeID,
					pErr.Message,
				)
			}
			s.l.Infow("trading pair created", "id", tradingPairID)

			atpID, err := s.createTradingBy(tx, assetID, tradingPairID)
			if err != nil {
				return 0, fmt.Errorf("failed to create asset trading pair for asset %d, tradingpair %d, err=%v",
					assetID, tradingPairID, err)
			}
			s.l.Infow("asset trading by created", "id", atpID)
		}
	}

	if err := s.createNewFeedWeight(tx, assetID, feedWeight); err != nil {
		return assetID, err
	}
	return assetID, nil
}

func (s Storage) createNewFeedWeight(tx *sqlx.Tx, assetID uint64, feedWeight *common.FeedWeight) error {
	// create new feed weights
	if feedWeight != nil {
		for feed, weight := range *feedWeight {
			var (
				feedWeightID uint64
			)
			if err := tx.NamedStmt(s.stmts.newFeedWeight).Get(&feedWeightID,
				struct {
					AssetID uint64  `db:"asset_id"`
					Feed    string  `db:"feed"`
					Weight  float64 `db:"weight"`
				}{
					AssetID: assetID,
					Feed:    feed,
					Weight:  weight,
				}); err != nil {
				return fmt.Errorf("failed to create feed weight for asset %d, feed %s, weight %f, err=%v",
					assetID, feed, weight, err)
			}
			s.l.Infow("feed weight created", "id", feedWeightID)
		}
	}
	return nil
}

type tradingPairDB struct {
	ID              uint64  `db:"id"`
	ExchangeID      uint64  `db:"exchange_id"`
	BaseID          uint64  `db:"base_id"`
	QuoteID         uint64  `db:"quote_id"`
	PricePrecision  uint64  `db:"price_precision"`
	AmountPrecision uint64  `db:"amount_precision"`
	AmountLimitMin  float64 `db:"amount_limit_min"`
	AmountLimitMax  float64 `db:"amount_limit_max"`
	PriceLimitMin   float64 `db:"price_limit_min"`
	PriceLimitMax   float64 `db:"price_limit_max"`
	MinNotional     float64 `db:"min_notional"`
	BaseSymbol      string  `db:"base_symbol"`
	QuoteSymbol     string  `db:"quote_symbol"`
}

func (tpd *tradingPairDB) ToCommon() common.TradingPair {
	return common.TradingPair{
		ID:              tpd.ID,
		Base:            tpd.BaseID,
		Quote:           tpd.QuoteID,
		PricePrecision:  tpd.PricePrecision,
		AmountPrecision: tpd.AmountPrecision,
		AmountLimitMin:  tpd.AmountLimitMin,
		AmountLimitMax:  tpd.AmountLimitMax,
		PriceLimitMin:   tpd.PriceLimitMin,
		PriceLimitMax:   tpd.PriceLimitMax,
		MinNotional:     tpd.MinNotional,
		ExchangeID:      tpd.ExchangeID,
	}
}

type assetDB struct {
	ID           uint64         `db:"id"`
	Symbol       string         `db:"symbol"`
	Name         string         `db:"name"`
	Address      sql.NullString `db:"address"`
	OldAddresses pq.StringArray `db:"old_addresses"`
	Decimals     uint64         `db:"decimals"`
	Transferable bool           `db:"transferable"`
	SetRate      string         `db:"set_rate"`
	Rebalance    bool           `db:"rebalance"`
	IsQuote      bool           `db:"is_quote"`

	PWIAskA                   *float64 `db:"pwi_ask_a"`
	PWIAskB                   *float64 `db:"pwi_ask_b"`
	PWIAskC                   *float64 `db:"pwi_ask_c"`
	PWIAskMinMinSpread        *float64 `db:"pwi_ask_min_min_spread"`
	PWIAskPriceMultiplyFactor *float64 `db:"pwi_ask_price_multiply_factor"`
	PWIBidA                   *float64 `db:"pwi_bid_a"`
	PWIBidB                   *float64 `db:"pwi_bid_b"`
	PWIBidC                   *float64 `db:"pwi_bid_c"`
	PWIBidMinMinSpread        *float64 `db:"pwi_bid_min_min_spread"`
	PWIBidPriceMultiplyFactor *float64 `db:"pwi_bid_price_multiply_factor"`

	RebalanceQuadraticA *float64 `db:"rebalance_quadratic_a"`
	RebalanceQuadraticB *float64 `db:"rebalance_quadratic_b"`
	RebalanceQuadraticC *float64 `db:"rebalance_quadratic_c"`

	TargetTotal              *float64 `db:"target_total"`
	TargetReserve            *float64 `db:"target_reserve"`
	TargetRebalanceThreshold *float64 `db:"target_rebalance_threshold"`
	TargetTransferThreshold  *float64 `db:"target_transfer_threshold"`

	PriceUpdateThreshold float64 `db:"stable_param_price_update_threshold"`
	AskSpread            float64 `db:"stable_param_ask_spread"`
	BidSpread            float64 `db:"stable_param_bid_spread"`
	SingleFeedMaxSpread  float64 `db:"stable_param_single_feed_max_spread"`
	MultipleFeedsMaxDiff float64 `db:"stable_param_multiple_feeds_max_diff"`

	Created time.Time `db:"created"`
	Updated time.Time `db:"updated"`
}

func (adb *assetDB) ToCommon() (common.Asset, error) {
	result := common.Asset{
		ID:           adb.ID,
		Symbol:       adb.Symbol,
		Name:         adb.Name,
		Address:      ethereum.Address{},
		Decimals:     adb.Decimals,
		Transferable: adb.Transferable,
		Rebalance:    adb.Rebalance,
		IsQuote:      adb.IsQuote,
		Created:      adb.Created,
		Updated:      adb.Updated,
	}

	if adb.Address.Valid {
		result.Address = ethereum.HexToAddress(adb.Address.String)
	}

	for _, oldAddress := range adb.OldAddresses {
		result.OldAddresses = append(result.OldAddresses, ethereum.HexToAddress(oldAddress))
	}

	setRate, ok := common.SetRateFromString(adb.SetRate)
	if !ok {
		return common.Asset{}, fmt.Errorf("invalid set rate value %s", adb.SetRate)
	}

	result.SetRate = setRate

	if adb.PWIAskA != nil &&
		adb.PWIAskB != nil &&
		adb.PWIAskC != nil &&
		adb.PWIAskMinMinSpread != nil &&
		adb.PWIAskPriceMultiplyFactor != nil &&
		adb.PWIBidA != nil &&
		adb.PWIBidB != nil &&
		adb.PWIBidC != nil &&
		adb.PWIBidMinMinSpread != nil &&
		adb.PWIBidPriceMultiplyFactor != nil {
		result.PWI = &common.AssetPWI{
			Ask: common.PWIEquation{
				A:                   *adb.PWIAskA,
				B:                   *adb.PWIAskB,
				C:                   *adb.PWIAskC,
				MinMinSpread:        *adb.PWIAskMinMinSpread,
				PriceMultiplyFactor: *adb.PWIAskPriceMultiplyFactor,
			},
			Bid: common.PWIEquation{
				A:                   *adb.PWIBidA,
				B:                   *adb.PWIBidB,
				C:                   *adb.PWIBidC,
				MinMinSpread:        *adb.PWIBidMinMinSpread,
				PriceMultiplyFactor: *adb.PWIBidPriceMultiplyFactor,
			},
		}
	}
	if adb.RebalanceQuadraticA != nil && adb.RebalanceQuadraticB != nil && adb.RebalanceQuadraticC != nil {
		result.RebalanceQuadratic = &common.RebalanceQuadratic{
			A: *adb.RebalanceQuadraticA,
			B: *adb.RebalanceQuadraticB,
			C: *adb.RebalanceQuadraticC,
		}
	}

	if adb.TargetTotal != nil &&
		adb.TargetReserve != nil &&
		adb.TargetRebalanceThreshold != nil &&
		adb.TargetTransferThreshold != nil {
		result.Target = &common.AssetTarget{
			Total:              *adb.TargetTotal,
			Reserve:            *adb.TargetReserve,
			RebalanceThreshold: *adb.TargetRebalanceThreshold,
			TransferThreshold:  *adb.TargetTransferThreshold,
		}
	}
	result.StableParam = common.StableParam{
		PriceUpdateThreshold: adb.PriceUpdateThreshold,
		AskSpread:            adb.AskSpread,
		BidSpread:            adb.BidSpread,
		SingleFeedMaxSpread:  adb.SingleFeedMaxSpread,
		MultipleFeedsMaxDiff: adb.MultipleFeedsMaxDiff,
	}
	return result, nil
}

// GetAssets return all assets listed
func (s *Storage) GetAssets() ([]common.Asset, error) {
	return s.getAssets(nil)
}

type tradingByDB struct {
	ID            uint64 `db:"id"`
	AssetID       uint64 `db:"asset_id"`
	TradingPairID uint64 `db:"trading_pair_id"`
}

func (db *tradingByDB) ToCommon() common.TradingBy {
	return common.TradingBy{
		TradingPairID: db.TradingPairID,
		AssetID:       db.AssetID,
	}
}

func toTradingPairMap(tps []tradingPairDB) map[uint64]tradingPairDB {
	res := make(map[uint64]tradingPairDB)
	for _, tp := range tps {
		res[tp.ID] = tp
	}
	return res
}

type feedWeightDB struct {
	ID      uint64  `db:"id"`
	AssetID uint64  `db:"asset_id"`
	Feed    string  `db:"feed"`
	Weight  float64 `db:"weight"`
}

func (s *Storage) getAssets(transferable *bool) ([]common.Asset, error) {
	var (
		allAssetDBs       []assetDB
		allAssetExchanges []assetExchangeDB
		allTradingPairs   []tradingPairDB
		allTradingBy      []tradingByDB
		allFeedWeights    []feedWeightDB
		results           []common.Asset
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAsset).Select(&allAssetDBs, nil, transferable); err != nil {
		return nil, err
	}

	if err := tx.NamedStmt(s.stmts.getAssetExchange).Select(&allAssetExchanges, assetExchangeCondition{}); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getTradingPair).Select(&allTradingPairs, nil); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getFeedWeight).Select(&allFeedWeights, nil); err != nil {
		return nil, err
	}

	if err = tx.Stmtx(s.stmts.getTradingBy).Select(&allTradingBy, nil); err != nil {
		return nil, err
	}
	tradingPairMap := toTradingPairMap(allTradingPairs)

	for _, assetDBResult := range allAssetDBs {
		result, err := assetDBResult.ToCommon()
		if err != nil {
			return nil, fmt.Errorf("invalid database record for asset id=%d err=%s", assetDBResult.ID, err.Error())
		}

		for _, assetExchangeResult := range allAssetExchanges {
			if assetExchangeResult.AssetID == assetDBResult.ID {
				exchange := assetExchangeResult.ToCommon()
				for _, tradingBy := range allTradingBy {
					if assetDBResult.ID == tradingBy.AssetID {
						if tradingPair, ok := tradingPairMap[tradingBy.TradingPairID]; ok &&
							tradingPair.ExchangeID == exchange.ExchangeID {
							exchange.TradingPairs = append(exchange.TradingPairs, tradingPair.ToCommon())
						}
					}
				}
				result.Exchanges = append(result.Exchanges, exchange)
			}
		}
		feeds := make(common.FeedWeight)
		for _, feed := range allFeedWeights {
			if feed.AssetID == assetDBResult.ID {
				feeds[feed.Feed] = feed.Weight
			}
		}
		result.FeedWeight = &feeds

		results = append(results, result)
	}

	return results, nil
}

// GetAsset get a single asset by id
func (s *Storage) GetAsset(id uint64) (common.Asset, error) {
	var (
		assetDBResult        assetDB
		assetExchangeResults []assetExchangeDB
		tradingPairResults   []tradingPairDB
		exchanges            []common.AssetExchange
		feedWeightResult     []feedWeightDB
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return common.Asset{}, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	if err := tx.NamedStmt(s.stmts.getAssetExchange).Select(&assetExchangeResults, assetExchangeCondition{
		AssetID: &id,
	}); err != nil {
		return common.Asset{}, fmt.Errorf("failed to query asset exchanges err=%s", err.Error())
	}

	for _, ae := range assetExchangeResults {
		exchanges = append(exchanges, ae.ToCommon())
	}

	if err := tx.Stmtx(s.stmts.getTradingPair).Select(&tradingPairResults, id); err != nil {
		return common.Asset{}, fmt.Errorf("failed to query for trading pairs err=%s", err.Error())
	}

	for _, pair := range tradingPairResults {
		for i := range exchanges {
			if exchanges[i].ExchangeID == pair.ExchangeID {
				exchanges[i].TradingPairs = append(exchanges[i].TradingPairs, pair.ToCommon())
			}
		}
	}

	// get feed weight
	if err := tx.Stmtx(s.stmts.getFeedWeight).Select(&feedWeightResult, id); err != nil {
		return common.Asset{}, fmt.Errorf("failed to get feed weight for asset: %d, error: %s", id, err.Error())
	}
	feeds := make(common.FeedWeight)
	for _, feed := range feedWeightResult {
		feeds[feed.Feed] = feed.Weight
	}

	s.l.Infow("getting asset", "id", id)
	err = tx.Stmtx(s.stmts.getAsset).Get(&assetDBResult, id, nil)
	switch err {
	case sql.ErrNoRows:
		return common.Asset{}, common.ErrNotFound
	case nil:
		result, err := assetDBResult.ToCommon()
		if err != nil {
			return common.Asset{}, fmt.Errorf("invalid database record for asset id=%d err=%s", assetDBResult.ID, err.Error())
		}
		result.Exchanges = exchanges
		result.FeedWeight = &feeds
		return result, nil
	default:
		return common.Asset{}, fmt.Errorf("failed to get asset from database id=%d err=%s", id, err.Error())
	}
}

// GetAssetBySymbol return asset by its symbol
func (s *Storage) GetAssetBySymbol(symbol string) (common.Asset, error) {
	var (
		result common.Asset
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return result, err
	}
	defer pgutil.RollbackUnlessCommitted(tx)

	s.l.Infow("getting assetBySymbol", "symbol", symbol)
	err = tx.Stmtx(s.stmts.getAssetBySymbol).Get(&result, symbol)
	switch err {
	case sql.ErrNoRows:
		return result, common.ErrNotFound
	case nil:
		return result, nil
	default:
		return result, fmt.Errorf("failed to get asset from database symbol=%s err=%s", symbol, err.Error())
	}
}

// UpdateAsset update asset with provide option
func (s *Storage) UpdateAsset(id uint64, opts storage.UpdateAssetOpts) error {
	return s.updateAsset(nil, id, opts)
}

type updateAssetParam struct {
	ID           uint64  `db:"id"`
	Symbol       *string `db:"symbol"`
	Name         *string `db:"name"`
	Address      *string `db:"address"`
	Decimals     *uint64 `db:"decimals"`
	Transferable *bool   `db:"transferable"`
	SetRate      *string `db:"set_rate"`
	Rebalance    *bool   `db:"rebalance"`
	IsQuote      *bool   `db:"is_quote"`

	AskA                   *float64 `db:"ask_a"`
	AskB                   *float64 `db:"ask_b"`
	AskC                   *float64 `db:"ask_c"`
	AskMinMinSpread        *float64 `db:"ask_min_min_spread"`
	AskPriceMultiplyFactor *float64 `db:"ask_price_multiply_factor"`
	BidA                   *float64 `db:"bid_a"`
	BidB                   *float64 `db:"bid_b"`
	BidC                   *float64 `db:"bid_c"`
	BidMinMinSpread        *float64 `db:"bid_min_min_spread"`
	BidPriceMultiplyFactor *float64 `db:"bid_price_multiply_factor"`

	RebalanceQuadraticA *float64 `db:"rebalance_quadratic_a"`
	RebalanceQuadraticB *float64 `db:"rebalance_quadratic_b"`
	RebalanceQuadraticC *float64 `db:"rebalance_quadratic_c"`

	TargetTotal              *float64 `db:"target_total"`
	TargetReserve            *float64 `db:"target_reserve"`
	TargetRebalanceThreshold *float64 `db:"target_rebalance_threshold"`
	TargetTransferThreshold  *float64 `db:"target_transfer_threshold"`

	PriceUpdateThreshold *float64 `db:"stable_param_price_update_threshold"`
	AskSpread            *float64 `db:"stable_param_ask_spread"`
	BidSpread            *float64 `db:"stable_param_bid_spread"`
	SingleFeedMaxSpread  *float64 `db:"stable_param_single_feed_max_spread"`
	MultipleFeedsMaxDiff *float64 `db:"stable_param_multiple_feeds_max_diff"`
}

func (s *Storage) updateAsset(tx *sqlx.Tx, id uint64, uo storage.UpdateAssetOpts) error {
	arg := updateAssetParam{
		ID:           id,
		Symbol:       uo.Symbol,
		Name:         uo.Name,
		Decimals:     uo.Decimals,
		Transferable: uo.Transferable,
		Rebalance:    uo.Rebalance,
		IsQuote:      uo.IsQuote,
	}

	var updateMsgs []string
	if uo.Symbol != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("symbol=%s", *uo.Symbol))
	}
	if uo.Name != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("name=%s", *uo.Name))
	}
	if uo.Address != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("address=%s", uo.Address.String()))
		addressStr := uo.Address.String()
		arg.Address = &addressStr
	}
	if uo.Decimals != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("decimals=%d", *uo.Decimals))
	}
	if uo.Transferable != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("transferable=%p", uo.Transferable))
	}
	if uo.SetRate != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("set_rate=%s", uo.SetRate.String()))
		setRateStr := uo.SetRate.String()
		arg.SetRate = &setRateStr
	}
	if uo.Rebalance != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("rebalance=%t", *uo.Rebalance))
	}
	if uo.IsQuote != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("is_quote=%t", *uo.IsQuote))
	}
	pwi := uo.PWI
	if pwi != nil {
		arg.AskA = &pwi.Ask.A
		arg.AskB = &pwi.Ask.B
		arg.AskC = &pwi.Ask.C
		arg.AskMinMinSpread = &pwi.Ask.MinMinSpread
		arg.AskPriceMultiplyFactor = &pwi.Ask.PriceMultiplyFactor
		arg.BidA = &pwi.Bid.A
		arg.BidB = &pwi.Bid.B
		arg.BidC = &pwi.Bid.C
		arg.BidMinMinSpread = &pwi.Bid.MinMinSpread
		arg.BidPriceMultiplyFactor = &pwi.Bid.PriceMultiplyFactor
		updateMsgs = append(updateMsgs, fmt.Sprintf("pwi=%+v", pwi))
	}
	rb := uo.RebalanceQuadratic

	if rb != nil {
		arg.RebalanceQuadraticA = &rb.A
		arg.RebalanceQuadraticB = &rb.B
		arg.RebalanceQuadraticC = &rb.C
		updateMsgs = append(updateMsgs, fmt.Sprintf("rebalance_quadratic=%+v", rb))
	}

	target := uo.Target

	if target != nil {
		arg.TargetTotal = &target.Total
		arg.TargetReserve = &target.Reserve
		arg.TargetRebalanceThreshold = &target.RebalanceThreshold
		arg.TargetTransferThreshold = &target.TransferThreshold
		updateMsgs = append(updateMsgs, fmt.Sprintf("target=%+v", target))
	}

	if uo.StableParam != nil {
		arg.PriceUpdateThreshold = uo.StableParam.PriceUpdateThreshold
		arg.AskSpread = uo.StableParam.AskSpread
		arg.BidSpread = uo.StableParam.BidSpread
		arg.SingleFeedMaxSpread = uo.StableParam.SingleFeedMaxSpread
		arg.MultipleFeedsMaxDiff = uo.StableParam.MultipleFeedsMaxDiff
		updateMsgs = append(updateMsgs, fmt.Sprintf("StableParam=%+v", target))
	}

	if uo.FeedWeight != nil {
		updateMsgs = append(updateMsgs, fmt.Sprintf("UpdateFeedWeight=%+v", *uo.FeedWeight))
	}

	if len(updateMsgs) == 0 {
		s.l.Infow("nothing set for update asset")
		return nil
	}
	var sts = s.stmts.updateAsset
	if tx != nil {
		sts = tx.NamedStmt(s.stmts.updateAsset)
	}

	s.l.Infow("updating asset", "id", id, "fields", strings.Join(updateMsgs, " "))
	var updatedID uint64
	err := sts.Get(&updatedID, arg)
	if err == sql.ErrNoRows {
		return common.ErrNotFound
	} else if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		if pErr.Code == errCodeUniqueViolation {
			switch pErr.Constraint {
			case "assets_symbol_key":
				s.l.Infow("conflict symbol when updating asset", "id", id, "err", pErr.Message)
				return common.ErrSymbolExists
			case addressesUniqueConstraint:
				s.l.Infow("conflict address when updating asset", "id", id, "err", pErr.Message)
				return common.ErrAddressExists

			}
		}
		if pErr.Code == errCodeCheckViolation {
			s.l.Infow("conflict address when updating asset", "id", id, "err", pErr.Message)
			switch pErr.Constraint {
			case "address_id_check":
				return common.ErrDepositAddressMissing
			case "pwi_check":
				return common.ErrPWIMissing
			case "rebalance_quadratic_check":
				return common.ErrRebalanceQuadraticMissing
			case "target_check":
				return common.ErrAssetTargetMissing
			}
		}

		return fmt.Errorf("failed to update asset err=%s", pErr)
	}
	// remove old feed weight
	if uo.FeedWeight != nil {
		if _, err := tx.Stmt(s.stmts.deleteFeedWeight.Stmt).Exec(uo.AssetID); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("failed to remove old feed weight: %s", err.Error())
			}
		}
	}

	if err := s.createNewFeedWeight(tx, uo.AssetID, uo.FeedWeight); err != nil {
		return err
	}

	return nil
}

// ChangeAssetAddress change address of an asset
func (s *Storage) ChangeAssetAddress(id uint64, address ethereum.Address) error {

	err := s.changeAssetAddress(nil, id, address)
	if err != nil {
		s.l.Infow("change address error", "err", err)
		return err
	}
	s.l.Infow("change asset address successfully", "id", id)
	return nil
}

func (s *Storage) changeAssetAddress(tx *sqlx.Tx, id uint64, address ethereum.Address) error {
	s.l.Infow("changing address", "asset_id", id, "new_address", address.String())
	sts := s.stmts.changeAssetAddress
	if tx != nil {
		sts = tx.Stmtx(sts)
	}
	_, err := sts.Exec(id, address.String())
	if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		switch pErr.Code {
		case errCodeUniqueViolation:
			if pErr.Constraint == addressesUniqueConstraint {
				s.l.Infow("conflict address when changing asset address id=%d err=%s", id, pErr.Message)
				return common.ErrAddressExists
			}
		case errAssertFailure:
			s.l.Infow("asset not found in database", "id", id, "err", pErr.Message)
			return common.ErrNotFound
		}
		return fmt.Errorf("failed to update asset err=%s", pErr)
	}
	return nil
}

// UpdateDepositAddress update deposit addresss for an AssetExchange
func (s *Storage) UpdateDepositAddress(assetID, exchangeID uint64, address ethereum.Address) error {
	var updated uint64
	err := s.stmts.updateDepositAddress.Get(&updated, assetID, exchangeID, address.Hex())
	switch err {
	case sql.ErrNoRows:
		return common.ErrNotFound
	case nil:
		s.l.Infow("asset deposit address is updated", "asset_exchange_id", updated, "deposit_address", address.Hex())
		return nil
	default:
		return fmt.Errorf("failed to update deposit address asset_id=%d exchange_id=%d err=%s", assetID, exchangeID, err.Error())
	}
}
