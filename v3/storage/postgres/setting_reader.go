package postgres

import (
	"database/sql"
	"log"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// GetTransferableAssets return list of transferable asset
func (s *Storage) GetTransferableAssets() ([]common.Asset, error) {
	transferable := true
	return s.getAssets(&transferable)
}

// GetTradingPair return trading pair by trading pair id
func (s *Storage) GetTradingPair(id uint64) (common.TradingPairSymbols, error) {
	var (
		tradingPairDB tradingPairDB
		result        common.TradingPairSymbols
	)

	if err := s.stmts.getTradingPairByID.Get(&tradingPairDB, id); err != nil {
		if err == sql.ErrNoRows {
			return result, common.ErrNotFound
		}
		return result, err
	}

	result = common.TradingPairSymbols{
		TradingPair: tradingPairDB.ToCommon(),
		BaseSymbol:  tradingPairDB.BaseSymbol,
		QuoteSymbol: tradingPairDB.QuoteSymbol,
	}

	return result, nil
}

type tpCondition struct {
	ExchangeID *uint64 `db:"exchange_id"`
	BaseID     *uint64 `db:"base_id"`
	QuoteID    *uint64 `db:"quote_id"`
}

// GetTradingPairs return list of trading pairs by exchangeID
func (s *Storage) GetTradingPairs(exchangeID uint64) ([]common.TradingPairSymbols, error) {
	var (
		tradingPairs []tradingPairDB
		result       []common.TradingPairSymbols
	)
	if err := s.stmts.getTradingPairSymbols.Select(&tradingPairs, tpCondition{
		ExchangeID: &exchangeID,
		BaseID:     nil,
		QuoteID:    nil,
	}); err != nil {
		return nil, err
	}
	for _, pair := range tradingPairs {
		result = append(result, common.TradingPairSymbols{
			TradingPair: pair.ToCommon(),
			BaseSymbol:  pair.BaseSymbol,
			QuoteSymbol: pair.QuoteSymbol,
		})
	}
	return result, nil
}

func (s *Storage) GetTradingPairsFromBaseQuote(baseID uint64, quoteID uint64) ([]common.TradingPairSymbols, error) {
	var (
		tradingPairs []tradingPairDB
		result       []common.TradingPairSymbols
	)
	if err := s.stmts.getTradingPairSymbols.Select(&tradingPairs, tpCondition{
		ExchangeID: nil,
		BaseID:     &baseID,
		QuoteID:    &quoteID,
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, common.ErrNotFound
		}
		return nil, err
	}
	for _, pair := range tradingPairs {
		result = append(result, common.TradingPairSymbols{
			TradingPair: pair.ToCommon(),
			BaseSymbol:  pair.BaseSymbol,
			QuoteSymbol: pair.QuoteSymbol,
		})
	}
	return result, nil
}

// TODO: rewrite this function to filter the exchange if in SQL query.
func (s *Storage) GetDepositAddresses(exchangeID uint64) (map[string]ethereum.Address, error) {
	var results = make(map[string]ethereum.Address)

	allAssets, err := s.GetAssets()
	if err != nil {
		return nil, err
	}

	for _, asset := range allAssets {
		for _, exchange := range asset.Exchanges {
			if exchange.ExchangeID == exchangeID {
				results[exchange.Symbol] = exchange.DepositAddress
			}
		}
	}

	return results, nil
}

// GetMinNotional return min notional
func (s *Storage) GetMinNotional(exchangeID, baseID, quoteID uint64) (float64, error) {
	var minNotional float64
	log.Printf("getting min notional for exchange=%d base=%d quote=%d",
		exchangeID, baseID, quoteID)
	if err := s.stmts.getMinNotional.Get(&minNotional,
		exchangeID, baseID, quoteID); err == sql.ErrNoRows {
		return 0, common.ErrNotFound
	} else if err != nil {
		return 0, err
	}
	return minNotional, nil
}
