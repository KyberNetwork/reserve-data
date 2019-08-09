package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

func (s *Storage) ConfirmCreateTradingPair(id uint64) error {
	var (
		createCreateTradingPair common.CreateCreateTradingPair
		err                     error
		pendingObject           common.PendingObject
	)
	pendingObject, err = s.GetPendingObject(id, common.PendingTypeCreateTradingPair)
	if err != nil {
		return err
	}
	err = json.Unmarshal(pendingObject.Data, &createCreateTradingPair)
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
	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeCreateTradingPair.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("pending trading pair #%d has been confirm successfully\n", id)
	return nil
}
