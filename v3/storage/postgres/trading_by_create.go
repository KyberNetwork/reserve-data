package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmCreateTradingBy to execute the pending trading by request
func (s *Storage) ConfirmCreateTradingBy(id uint64) error {
	var (
		createCreateTradingBy common.CreateCreateTradingBy
		err                   error
		pendingObject         common.PendingObject
	)
	pendingObject, err = s.GetPendingObject(id, common.PendingTypeCreateTradingBy)
	if err != nil {
		return err
	}
	err = json.Unmarshal(pendingObject.Data, &createCreateTradingBy)
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
	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeCreateTradingBy.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("pending trading by #%d has been confirm successfully\n", id)
	return nil
}
