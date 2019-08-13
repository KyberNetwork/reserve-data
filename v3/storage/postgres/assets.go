package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
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
) (uint64, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessCommitted(tx)

	id, err := s.createAsset(
		tx,
		symbol,
		name,
		address,
		decimals,
		transferable,
		setRate,
		rebalance,
		isQuote,
		pwi,
		rb,
		exchanges,
		target,
	)
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
	minDeposit, withdrawFee, targetRecommended, targetRatio float64) (uint64, error) {

	tx, err := s.db.Beginx()
	if err != nil {
		return 0, err
	}
	defer rollbackUnlessCommitted(tx)
	id, err := s.createAssetExchange(tx, exchangeID, assetID, symbol, depositAddress, minDeposit, withdrawFee,
		targetRecommended, targetRatio)
	if err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return id, nil
}

// UpdateAssetExchange update information about asset exchange
func (s *Storage) UpdateAssetExchange(id uint64, opts storage.UpdateAssetExchangeOpts) error {

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}

	defer rollbackUnlessCommitted(tx)
	err = s.updateAssetExchange(tx, id, opts)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *Storage) createAssetExchange(tx *sqlx.Tx, exchangeID, assetID uint64, symbol string,
	depositAddress ethereum.Address, minDeposit, withdrawFee, targetRecommended, targetRatio float64) (uint64, error) {
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
		log.Printf("failed to create new asset exchange err=%s", pErr.Message)
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
		log.Printf("nothing set for update asset exchange, skip now")
		return nil
	}

	log.Printf("updating asset_exchange %d %s", id, strings.Join(updateMsgs, " "))
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
		log.Printf("asset_exchange not found in database id=%d", id)
		return common.ErrNotFound
	} else if err != nil {
		return fmt.Errorf("failed to update asset err=%s", err)
	}
	// TODO: check more error
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
) (uint64, error) {
	var assetID uint64

	if transferable && common.IsZeroAddress(address) {
		return 0, common.ErrAddressMissing
	}

	for _, exchange := range exchanges {
		if transferable && common.IsZeroAddress(exchange.DepositAddress) {
			return 0, common.ErrDepositAddressMissing
		}
	}

	log.Printf("creating new asset symbol=%s adress=%s", symbol, address.String())

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
			log.Printf("rebalance is enabled but rebalance quadratic is invalid symbol=%s", symbol)
			return 0, common.ErrRebalanceQuadraticMissing
		}

		if len(exchanges) == 0 {
			log.Printf("rebalance is enabled but no exchange configuration is provided symbol=%s", symbol)
			return 0, common.ErrAssetExchangeMissing
		}

		if target == nil {
			log.Printf("rebalance is enabled but target configuration is invalid symbol=%s", symbol)
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

	if err := tx.NamedStmt(s.stmts.newAsset).Get(&assetID, arg); err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown returned err=%s", err.Error())
		}

		log.Printf("failed to create new asset err=%s", pErr.Message)
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

		log.Printf("asset exchange is created id=%d", assetExchangeID)

		for _, pair := range exchange.TradingPairs {
			var (
				tradingPairID uint64
				baseID        = pair.Base
				quoteID       = pair.Quote
			)

			if baseID != 0 && quoteID != 0 {
				log.Printf(
					"both base and quote are provided asset_symbol=%s exchange_id=%d",
					symbol,
					exchange.ExchangeID)
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
					log.Printf("failed to create trading pair as assertion failed symbol=%s exchange_id=%d err=%s",
						symbol,
						exchange.ExchangeID,
						pErr.Message)
					return 0, common.ErrBadTradingPairConfiguration
				case errBaseInvalid:
					log.Printf("failed to create trading pair as check base failed symbol=%s exchange_id=%d err=%s",
						symbol,
						exchange.ExchangeID,
						pErr.Message)
					return 0, common.ErrBaseAssetInvalid
				case errQuoteInvalid:
					log.Printf("failed to create trading pair as check quote failed symbol=%s exchange_id=%d err=%s",
						symbol,
						exchange.ExchangeID,
						pErr.Message)
					return 0, common.ErrQuoteAssetInvalid
				}

				return 0, fmt.Errorf("failed to create trading pair symbol=%s exchange_id=%d err=%s",
					symbol,
					exchange.ExchangeID,
					pErr.Message,
				)
			}
			log.Printf("trading pair created id=%d", tradingPairID)

			atpID, err := s.createTradingBy(tx, assetID, tradingPairID)
			if err != nil {
				return 0, fmt.Errorf("failed to create asset trading pair for asset %d, tradingpair %d, err=%v",
					assetID, tradingPairID, err)
			}
			log.Printf("asset trading pair created %d\n", atpID)
		}
	}

	return assetID, nil
}

type assetExchangeDB struct {
	ID                uint64          `db:"id"`
	ExchangeID        uint64          `db:"exchange_id"`
	AssetID           uint64          `db:"asset_id"`
	Symbol            string          `db:"symbol"`
	DepositAddress    sql.NullString  `db:"deposit_address"`
	MinDeposit        float64         `db:"min_deposit"`
	WithdrawFee       float64         `db:"withdraw_fee"`
	PricePrecision    int64           `db:"price_precision"`
	AmountPrecision   int64           `db:"amount_precision"`
	AmountLimitMin    float64         `db:"amount_limit_min"`
	AmountLimitMax    float64         `db:"amount_limit_max"`
	PriceLimitMin     float64         `db:"price_limit_min"`
	PriceLimitMax     float64         `db:"price_limit_max"`
	TargetRecommended sql.NullFloat64 `db:"target_recommended"`
	TargetRatio       sql.NullFloat64 `db:"target_ratio"`
}

func (aedb *assetExchangeDB) ToCommon() common.AssetExchange {
	result := common.AssetExchange{
		ID:           aedb.ID,
		AssetID:      aedb.AssetID,
		ExchangeID:   aedb.ExchangeID,
		Symbol:       aedb.Symbol,
		MinDeposit:   aedb.MinDeposit,
		WithdrawFee:  aedb.WithdrawFee,
		TradingPairs: nil,
	}
	if aedb.DepositAddress.Valid {
		result.DepositAddress = ethereum.HexToAddress(aedb.DepositAddress.String)
	}
	if aedb.TargetRecommended.Valid {
		result.TargetRecommended = aedb.TargetRecommended.Float64
	}
	if aedb.TargetRatio.Valid {
		result.TargetRatio = aedb.TargetRatio.Float64
	}
	return result
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

func (s *Storage) getAssets(transferable *bool) ([]common.Asset, error) {
	var (
		allAssetDBs       []assetDB
		allAssetExchanges []assetExchangeDB
		allTradingPairs   []tradingPairDB
		allTradingBy      []tradingByDB
		results           []common.Asset
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer rollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAsset).Select(&allAssetDBs, nil, transferable); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getAssetExchange).Select(&allAssetExchanges, nil, nil); err != nil {
		return nil, err
	}

	if err := tx.Stmtx(s.stmts.getTradingPair).Select(&allTradingPairs, nil); err != nil {
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
	)

	tx, err := s.db.Beginx()
	if err != nil {
		return common.Asset{}, err
	}
	defer rollbackUnlessCommitted(tx)

	if err := tx.Stmtx(s.stmts.getAssetExchange).Select(&assetExchangeResults, id, nil); err != nil {
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

	log.Printf("getting asset id=%d", id)
	err = tx.Stmtx(s.stmts.getAsset).Get(&assetDBResult, id, nil)
	switch err {
	case sql.ErrNoRows:
		log.Printf("asset not found id=%d", id)
		return common.Asset{}, common.ErrNotFound
	case nil:
		result, err := assetDBResult.ToCommon()
		if err != nil {
			return common.Asset{}, fmt.Errorf("invalid database record for asset id=%d err=%s", assetDBResult.ID, err.Error())
		}
		result.Exchanges = exchanges
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
	defer rollbackUnlessCommitted(tx)

	log.Printf("getting asset symbol=%s", symbol)
	err = tx.Stmtx(s.stmts.getAssetBySymbol).Get(&result, symbol)
	switch err {
	case sql.ErrNoRows:
		log.Printf("asset not found symbol=%s", symbol)
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

	if len(updateMsgs) == 0 {
		log.Printf("nothing set for update asset, skip now")
		return nil
	}
	var sts = s.stmts.updateAsset
	if tx != nil {
		sts = tx.NamedStmt(s.stmts.updateAsset)
	}

	log.Printf("updating asset %d %s", id, strings.Join(updateMsgs, " "))
	var updatedID uint64
	err := sts.Get(&updatedID, arg)
	if err == sql.ErrNoRows {
		log.Printf("asset not found in database id=%d", id)
		return common.ErrNotFound
	} else if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		if pErr.Code == errCodeUniqueViolation {
			switch pErr.Constraint {
			case "assets_symbol_key":
				log.Printf("conflict symbol when updating asset id=%d err=%s", id, pErr.Message)
				return common.ErrSymbolExists
			case addressesUniqueConstraint:
				log.Printf("conflict address when updating asset id=%d err=%s", id, pErr.Message)
				return common.ErrAddressExists
			}
		}

		return fmt.Errorf("failed to update asset err=%s", pErr)
	}
	return nil
}

// ChangeAssetAddress change address of an asset
func (s *Storage) ChangeAssetAddress(id uint64, address ethereum.Address) error {

	err := s.changeAssetAddress(nil, id, address)
	if err != nil {
		log.Printf("change address error, err=%v\n", err)
		return err
	}
	log.Printf("change asset address successfully id=%d\n", id)
	return nil
}

func (s *Storage) changeAssetAddress(tx *sqlx.Tx, id uint64, address ethereum.Address) error {
	log.Printf("changing address of asset id=%d new_address=%s", id, address.String())
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
				log.Printf("conflict address when changing asset address id=%d err=%s", id, pErr.Message)
				return common.ErrAddressExists
			}
		case errAssertFailure:
			log.Printf("asset not found in database id=%d err=%s", id, pErr.Message)
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
		log.Printf("asset deposit address is updated asset_exchange_id=%d deposit_address=%s",
			updated, address.Hex())
		return nil
	default:
		return fmt.Errorf("failed to update deposit address asset_id=%d exchange_id=%d err=%s", assetID, exchangeID, err.Error())
	}
}
