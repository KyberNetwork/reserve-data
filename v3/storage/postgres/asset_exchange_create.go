package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmCreateAssetExchange confirm pending asset exchange, return err if any
func (s *Storage) ConfirmCreateAssetExchange(id uint64) error {
	var (
		ccAssetExchange common.CreateCreateAssetExchange
		err             error
		pendingObject   common.PendingObject
	)
	pendingObject, err = s.GetPendingObject(id, common.PendingTypeCreateAssetExchange)
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
		_, err = s.createAssetExchange(tx, r.ExchangeID, r.AssetID, r.Symbol, r.DepositAddress, r.MinDeposit,
			r.WithdrawFee, r.TargetRecommended, r.TargetRatio)
		if err != nil {
			return err
		}
	}
	_, err = s.stmts.deletePendingObject.Exec(id, common.PendingTypeCreateAssetExchange.String())
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
