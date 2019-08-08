package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

// ConfirmUpdateExchange apply pending changes in UpdateExchange object.
func (s *Storage) ConfirmUpdateExchange(msg []byte) error {
	var (
		r   common.CreateUpdateExchange
		err error
	)
	err = json.Unmarshal(msg, &r)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, e := range r.Exchanges {
		err = s.updateExchange(tx, e.ExchangeID, storage.UpdateExchangeOpts{
			TradingFeeMaker: e.TradingFeeMaker,
			TradingFeeTaker: e.TradingFeeTaker,
			Disable:         e.Disable,
		})
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
