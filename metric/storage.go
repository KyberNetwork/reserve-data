package metric

import (
	"github.com/KyberNetwork/reserve-data/common"
)

// MetricStorage is the interface that wraps all metrics database operations.
type MetricStorage interface {
	StoreMetric(data *common.MetricEntry, timepoint uint64) error
	StoreTokenTargetQty(id, data string) error
	StorePendingTargetQty(data, dataType string) error
	StoreRebalanceControl(status bool) error
	StoreSetrateControl(status bool) error
	StorePendingPWIEquation(data string) error
	StorePWIEquation(data string) error

	GetMetric(tokens []common.Token, fromTime, toTime uint64) (map[string]common.MetricList, error)
	GetTokenTargetQty() (common.TokenTargetQty, error)
	GetPendingTargetQty() (common.TokenTargetQty, error)
	GetRebalanceControl() (common.RebalanceControl, error)
	GetSetrateControl() (common.SetrateControl, error)
	GetPendingPWIEquation() (common.PWIEquation, error)
	GetPWIEquation() (common.PWIEquation, error)

	RemovePendingTargetQty() error
	RemovePendingPWIEquation() error

	SetStableTokenParams(value []byte) error
	ConfirmStableTokenParams(value []byte) error
	RemovePendingStableTokenParams() error
	GetPendingStableTokenParams() (map[string]interface{}, error)
	GetStableTokenParams() (map[string]interface{}, error)

	StorePendingTargetQtyV2(value []byte) error
	ConfirmTargetQtyV2(value []byte) error
	RemovePendingTargetQtyV2() error
	GetPendingTargetQtyV2() (common.TokenTargetQtyV2, error)
	GetTargetQtyV2() (common.TokenTargetQtyV2, error)

	StorePendingPWIEquationV2([]byte) error
	GetPendingPWIEquationV2() (common.PWIEquationRequestV2, error)
	StorePWIEquationV2(data string) error
	RemovePendingPWIEquationV2() error
	GetPWIEquationV2() (common.PWIEquationRequestV2, error)

	StorePendingRebalanceQuadratic([]byte) error
	GetPendingRebalanceQuadratic() (common.RebalanceQuadraticRequest, error)
	ConfirmRebalanceQuadratic(data []byte) error
	RemovePendingRebalanceQuadratic() error
	GetRebalanceQuadratic() (common.RebalanceQuadraticRequest, error)
	// StorePendingtokenUpdateInfo will attempt to store targetquanty v2. PWIequation V2 and RebalanceQuadracticRequest into database
	// it returns error if occur.
	StorePendingTokenUpdateInfo(common.TokenTargetQtyV2, common.PWIEquationRequestV2, common.RebalanceQuadraticRequest) error
	ConfirmTokenUpdateInfo(common.TokenTargetQtyV2, common.PWIEquationRequestV2, common.RebalanceQuadraticRequest) error
	RemovePendingTokenUpdateInfo() error
}
