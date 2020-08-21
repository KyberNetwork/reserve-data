package postgres

import (
	"database/sql"

	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// assetExchangeCondition is placeholder for namedStmt to select asset exchange
type assetExchangeCondition struct {
	AssetID    *rtypes.AssetID         `db:"asset_id"`
	ExchangeID *rtypes.ExchangeID      `db:"exchange_id"`
	ID         *rtypes.AssetExchangeID `db:"id"`
}

type assetExchangeDB struct {
	ID                rtypes.AssetExchangeID `db:"id"`
	ExchangeID        rtypes.ExchangeID      `db:"exchange_id"`
	AssetID           rtypes.AssetID         `db:"asset_id"`
	Symbol            string                 `db:"symbol"`
	DepositAddress    sql.NullString         `db:"deposit_address"`
	MinDeposit        float64                `db:"min_deposit"`
	WithdrawFee       float64                `db:"withdraw_fee"`
	PricePrecision    int64                  `db:"price_precision"`
	AmountPrecision   int64                  `db:"amount_precision"`
	AmountLimitMin    float64                `db:"amount_limit_min"`
	AmountLimitMax    float64                `db:"amount_limit_max"`
	PriceLimitMin     float64                `db:"price_limit_min"`
	PriceLimitMax     float64                `db:"price_limit_max"`
	TargetRecommended sql.NullFloat64        `db:"target_recommended"`
	TargetRatio       sql.NullFloat64        `db:"target_ratio"`
}

func (aeDB *assetExchangeDB) ToCommon() common.AssetExchange {
	result := common.AssetExchange{
		ID:           aeDB.ID,
		AssetID:      aeDB.AssetID,
		ExchangeID:   aeDB.ExchangeID,
		Symbol:       aeDB.Symbol,
		MinDeposit:   aeDB.MinDeposit,
		WithdrawFee:  aeDB.WithdrawFee,
		TradingPairs: nil,
	}
	if aeDB.DepositAddress.Valid {
		result.DepositAddress = ethereum.HexToAddress(aeDB.DepositAddress.String)
	}
	if aeDB.TargetRecommended.Valid {
		result.TargetRecommended = aeDB.TargetRecommended.Float64
	}
	if aeDB.TargetRatio.Valid {
		result.TargetRatio = aeDB.TargetRatio.Float64
	}
	return result
}
