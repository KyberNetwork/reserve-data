package postgres

import (
	"database/sql"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// GetTransferableAssets return list of transferable asset
func (s *Storage) GetTransferableAssets() ([]common.Asset, error) {
	transferable := true
	return s.getAssets(&transferable)
}

// GetTradingPair return trading pair by trading pair id
func (s *Storage) GetTradingPair(id rtypes.TradingPairID, withDeleted bool) (common.TradingPairSymbols, error) {
	var (
		tradingPairDB tradingPairDB
		result        common.TradingPairSymbols
	)

	if err := s.stmts.getTradingPairByID.Get(&tradingPairDB, id, withDeleted); err != nil {
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

// GetTradingPairs return list of trading pairs by exchangeID
func (s *Storage) GetTradingPairs(id rtypes.ExchangeID) ([]common.TradingPairSymbols, error) {
	var (
		tradingPairs []tradingPairDB
		result       []common.TradingPairSymbols
	)
	if err := s.stmts.getTradingPairSymbols.Select(&tradingPairs, id); err != nil {
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

func (s *Storage) GetDepositAddresses(exchangeID rtypes.ExchangeID) (map[rtypes.AssetID]ethereum.Address, error) {
	var (
		dbResult []assetExchangeDB
		results  = make(map[rtypes.AssetID]ethereum.Address)
	)
	err := s.stmts.getAssetExchange.Select(&dbResult, assetExchangeCondition{
		ExchangeID: &exchangeID,
	})
	if err != nil {
		return nil, err
	}
	for _, r := range dbResult {
		if r.DepositAddress.Valid {
			results[r.AssetID] = ethereum.HexToAddress(r.DepositAddress.String)
		} else {
			results[r.AssetID] = ethereum.HexToAddress("0x0")
		}
	}
	return results, nil
}

// GetMinNotional return min notional
func (s *Storage) GetMinNotional(exchangeID rtypes.ExchangeID, baseID, quoteID rtypes.AssetID) (float64, error) {
	var minNotional float64
	s.l.Infow("getting min notional", "exchange", exchangeID, "base", baseID, "quote", quoteID)
	if err := s.stmts.getMinNotional.Get(&minNotional,
		exchangeID, baseID, quoteID); err == sql.ErrNoRows {
		return 0, common.ErrNotFound
	} else if err != nil {
		return 0, err
	}
	return minNotional, nil
}
