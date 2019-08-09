package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Storage) ConfirmUpdateTradingPair(id uint64) error {
	var (
		createUpdateTradingPair common.CreateUpdateTradingPair
		err                     error
		pendingObject           common.PendingObject
	)

	pendingObject, err = s.GetPendingObject(id, common.PendingTypeUpdateTradingPair)
	if err != nil {
		return err
	}

	err = json.Unmarshal(pendingObject.Data, &createUpdateTradingPair)
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

	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeUpdateTradingPair.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("update trading pair #%d has been confirm successfully\n", id)
	return nil
}
