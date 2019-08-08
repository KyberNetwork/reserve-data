package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/KyberNetwork/reserve-data/v3/storage"
)

func (s *Storage) ConfirmUpdateAsset(msg []byte) error {
	var (
		r   common.CreateUpdateAsset
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
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
