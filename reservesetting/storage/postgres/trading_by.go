package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Storage) createTradingBy(tx *sqlx.Tx, assetID rtypes.AssetID, tradingPairID rtypes.TradingPairID) (rtypes.TradingByID, error) {
	var tradingByID rtypes.TradingByID
	err := tx.Stmtx(s.stmts.newTradingBy).Get(&tradingByID, assetID, tradingPairID)
	if err != nil {
		pErr, ok := err.(*pq.Error)
		if !ok {
			return 0, fmt.Errorf("unknown error returned err=%s", err.Error())
		}

		if pErr.Code == errCodeUniqueViolation {
			return 0, common.ErrTradingByAlreadyExists
		}

		return 0, fmt.Errorf("failed to create TradingBy, err=%s", pErr)
	}
	return tradingByID, nil
}

// GetTradingBy get a trading by with a given ID
func (s *Storage) GetTradingBy(tradingByID rtypes.TradingByID) (common.TradingBy, error) {
	var (
		result tradingByDB
	)
	err := s.stmts.getTradingBy.Get(&result, tradingByID)
	switch err {
	case sql.ErrNoRows:
		return common.TradingBy{}, common.ErrNotFound
	case nil:
		return result.ToCommon(), nil
	default:
		return common.TradingBy{}, err
	}
}
