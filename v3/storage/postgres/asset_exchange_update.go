package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmUpdateAssetExchange confirm pending asset exchange, return err if any
func (s *Storage) ConfirmUpdateAssetExchange(msg []byte) error {
	var (
		ccAssetExchange common.CreateUpdateAssetExchange
		err             error
	)
	err = json.Unmarshal(msg, &ccAssetExchange)
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
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
