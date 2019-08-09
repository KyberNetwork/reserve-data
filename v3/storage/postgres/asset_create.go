package postgres

import (
	"encoding/json"
	"log"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// GetCreateAsset execute a create asset
func (s *Storage) ConfirmCreateAsset(id uint64) error {
	var (
		createCreateAsset common.CreateCreateAsset
		pendingObj        common.PendingObject
		err               error
	)
	pendingObj, err = s.GetPendingObject(id, common.PendingTypeCreateAsset)
	if err != nil {
		return err
	}

	err = json.Unmarshal(pendingObj.Data, &createCreateAsset)
	if err != nil {
		return err
	}

	tx, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer rollbackUnlessCommitted(tx)
	for _, a := range createCreateAsset.AssetInputs {
		_, err := s.createAsset(tx, a.Symbol, a.Name, a.Address, a.Decimals, a.Transferable, a.SetRate, a.Rebalance,
			a.IsQuote, a.PWI, a.RebalanceQuadratic, a.Exchanges, a.Target)
		if err != nil {
			return err
		}
	}
	_, err = tx.Stmtx(s.stmts.deletePendingObject).Exec(id, common.PendingTypeCreateAsset.String())
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	log.Printf("pending asset #%d has been confirm successfully\n", id)
	return nil
}
