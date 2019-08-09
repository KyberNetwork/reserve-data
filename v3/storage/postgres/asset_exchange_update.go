package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmUpdateAssetExchange confirm pending asset exchange, return err if any
func (s *Storage) ConfirmUpdateAssetExchange(id uint64) error {
	var (
		ccAssetExchange common.CreateUpdateAssetExchange
		err             error
		pendingObject   common.PendingObject
	)
	pendingObject, err = s.GetPendingObject(id, common.PendingTypeUpdateAssetExchange)
	if err != nil {
		return err
	}
	err = json.Unmarshal(pendingObject.Data, &ccAssetExchange)
	if err != nil {
		return err
	}
	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, r := range ccAssetExchange.AssetExchanges {
		err = s.updateAssetExchange(tx, r.ID, r)
		if err != nil {
			return err
		}
	}
	_, err = s.stmts.deletePendingObject.Exec(id, common.PendingTypeUpdateAssetExchange.String())
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("pending asset exchange #%d has been confirm successfully\n", id)
	return nil
}
