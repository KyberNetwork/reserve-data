package postgres

import (
	"encoding/json"

	"github.com/KyberNetwork/reserve-data/v3/common"
)

// ConfirmCreateAssetExchange confirm pending asset exchange, return err if any
func (s *Storage) ConfirmCreateAssetExchange(msg []byte) error {
	var (
		ccAssetExchange common.CreateCreateAssetExchange
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
		_, err = s.createAssetExchange(tx, r.ExchangeID, r.AssetID, r.Symbol, r.DepositAddress, r.MinDeposit,
			r.WithdrawFee, r.TargetRecommended, r.TargetRatio)
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
