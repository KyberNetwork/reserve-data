package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

func (s *Storage) ConfirmUpdateAsset(id uint64) error {
	var (
		r          common.CreateUpdateAsset
		err        error
		pendingObj common.PendingObject
	)
	pendingObj, err = s.GetPendingObject(id, common.PendingTypeUpdateAsset)
	if err != nil {
		return err
	}
	err = json.Unmarshal(pendingObj.Data, &r)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, e := range r.Assets {
		err = s.updateAsset(tx, e.AssetID, storage.UpdateAssetOpts{
			Symbol:             e.Symbol,
			Transferable:       e.Transferable,
			Address:            e.Address,
			IsQuote:            e.IsQuote,
			Rebalance:          e.Rebalance,
			SetRate:            e.SetRate,
			Decimals:           e.Decimals,
			Name:               e.Name,
			Target:             e.Target,
			PWI:                e.PWI,
			RebalanceQuadratic: e.RebalanceQuadratic,
		})
		if err != nil {
			return err
		}
	}

	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeUpdateAsset.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("update asset #%d has been confirm successfully\n", id)
	return nil
}
