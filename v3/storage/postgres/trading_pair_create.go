package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Storage) ConfirmCreateTradingPair(msg []byte) error {
	var (
		createCreateTradingPair common.CreateCreateTradingPair
		err                     error
	)
	err = json.Unmarshal(msg, &createCreateTradingPair)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, a := range createCreateTradingPair.TradingPairs {
		_, err := s.createTradingPair(tx, a.ExchangeID, a.Base, a.Quote, a.PricePrecision, a.AmountPrecision,
			a.AmountLimitMin, a.AmountLimitMax, a.PriceLimitMin, a.PriceLimitMax, a.MinNotional)
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
