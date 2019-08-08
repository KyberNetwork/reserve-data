package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Storage) ConfirmUpdateTradingPair(msg []byte) error {
	var (
		createUpdateTradingPair common.CreateUpdateTradingPair
		err                     error
	)
	err = json.Unmarshal(msg, &createUpdateTradingPair)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, a := range createUpdateTradingPair.TradingPairs {
		err = s.updateTradingPair(tx, a.ID, a)
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
