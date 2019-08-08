package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmCreateTradingBy to execute the pending trading by request
func (s *Storage) ConfirmCreateTradingBy(msg []byte) error {
	var (
		createCreateTradingBy common.CreateCreateTradingBy
		err                   error
	)
	err = json.Unmarshal(msg, &createCreateTradingBy)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, tradingByEntry := range createCreateTradingBy.TradingBys {
		_, err := s.createTradingBy(tx, tradingByEntry.AssetID, tradingByEntry.TradingPairID)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
