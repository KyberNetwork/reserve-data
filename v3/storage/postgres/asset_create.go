package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// GetCreateAsset execute a create asset
func (s *Storage) ConfirmCreateAsset(msg []byte) error {
	var (
		createCreateAsset common.CreateCreateAsset
		err               error
	)

	err = json.Unmarshal(msg, &createCreateAsset)
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
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
