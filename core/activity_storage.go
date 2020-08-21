package core

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// ActivityStorage is the interface contains all database operations of core.
type ActivityStorage interface {
	Record(
		action string,
		id common.ActivityID,
		destination string,
		params common.ActivityParams,
		result common.ActivityResult,
		estatus string,
		mstatus string,
		timepoint uint64) error
	HasPendingDeposit(token commonv3.Asset, exchange common.Exchange) (bool, error)

	GetActivity(exchangeID rtypes.ExchangeID, orderID string) (common.ActivityRecord, error)

	// PendingActivityForAction return the last pending set rate and number of pending
	// transactions.
	PendingActivityForAction(minedNonce uint64, activityType string) (*common.ActivityRecord, uint64, error)
}
