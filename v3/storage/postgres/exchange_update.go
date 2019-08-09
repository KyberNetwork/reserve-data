package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

// ConfirmUpdateExchange apply pending changes in UpdateExchange object.
func (s *Storage) ConfirmUpdateExchange(id uint64) error {
	var (
		r             common.CreateUpdateExchange
		err           error
		pendingObject common.PendingObject
	)
	pendingObject, err = s.GetPendingObject(id, common.PendingTypeUpdateExchange)
	if err != nil {
		return err
	}

	err = json.Unmarshal(pendingObject.Data, &r)
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

	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeUpdateExchange.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("update exchange #%d has been confirm successfully\n", id)
	return nil
}
