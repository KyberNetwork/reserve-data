package common

type GoldRate struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   uint64  `json:"time"`
}

type GoldData struct {
	Valid     bool
	Timestamp uint64
	Status    string     `json:"success"`
	Data      []GoldRate `json:"data"`
}

// {"value":2.00591,"text":"1 XAU is worth 2.00591 ETH","timestamp":1524811993}
