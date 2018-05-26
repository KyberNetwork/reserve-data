package metric

type TokenMetric struct {
	AfpMid float64
	Spread float64
}

type MetricEntry struct {
	Timestamp uint64
	// data contain all token metric for all tokens
	Data map[string]TokenMetric
}

type TokenMetricResponse struct {
	Timestamp uint64
	AfpMid    float64
	Spread    float64
}

// Metric list for one token
type MetricList []TokenMetricResponse

type MetricResponse struct {
	Timestamp  uint64
	ReturnTime uint64
	Data       map[string]MetricList
}

type TokenTargetQty struct {
	ID        uint64
	Timestamp uint64
	Data      string
	Status    string
	Type      int64
}

type PWIEquation struct {
	ID   uint64 `json:"id"`
	Data string `json:"data"`
}

type RebalanceControl struct {
	Status bool `json:"status"`
}

type SetrateControl struct {
	Status bool `json:"status"`
}

type TargetQtySet struct {
	TotalTarget        float64
	ReserveTarget      float64
	RebalanceThreshold float64
	TransferThreshold  float64
}

type TargetQtyStruct struct {
	SetTarget TargetQtySet
}
