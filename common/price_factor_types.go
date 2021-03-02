package common

import (
	"fmt"
)

// TokenTargetQty ...
type TokenTargetQty struct {
	ID        uint64
	Timestamp uint64
	Data      string
	Status    string
	Type      int64
}

// PWIEquation ...
type PWIEquation struct {
	ID   uint64 `json:"id"`
	Data string `json:"data"`
}

// RebalanceControl represent status of rebalance, true is enable and false is disable
type RebalanceControl struct {
	Status bool `json:"status"`
}

// SetrateControl represent status of set rate ability, true is enable and false is disable
type SetrateControl struct {
	Status bool `json:"status"`
}

//TargetQtySet represent a set of target quantity
type TargetQtySet struct {
	TotalTarget        float64 `json:"total_target"`
	ReserveTarget      float64 `json:"reserve_target"`
	RebalanceThreshold float64 `json:"rebalance_threshold"`
	TransferThreshold  float64 `json:"transfer_threshold"`
}

//TargetQtyV2 object for save target qty
type TargetQtyV2 struct {
	SetTarget        TargetQtySet       `json:"set_target"`
	RecommendBalance map[string]float64 `json:"recomend_balance,omitempty"`
	ExchangeRatio    map[string]float64 `json:"exchange_ratio,omitempty"`
}

//TokenTargetQtyV2 represent a map of token and its target quantity struct
type TokenTargetQtyV2 map[string]TargetQtyV2

// PWIEquationV2 contains the information of a PWI equation.
type PWIEquationV2 struct {
	A                   float64 `json:"a"`
	B                   float64 `json:"b"`
	C                   float64 `json:"c"`
	MinMinSpread        float64 `json:"min_min_spread"`
	PriceMultiplyFactor float64 `json:"price_multiply_factor"`
}

// PWIEquationTokenV2 is a mapping between a token id and a PWI equation.
type PWIEquationTokenV2 map[string]PWIEquationV2

// validate checks if the equation has all the required fields.
// Example:
// {
//  "bid": {
//    "a": "750",
//    "b": "500",
//    "c": "0",
//    "min_min_spread": "0",
//    "price_multiply_factor": "0"
//  },
//  "ask": {
//    "a": "800",
//    "b": "600",
//    "c": "0",
//    "min_min_spread": "0",
//    "price_multiply_factor": "0"
//  }
//}
func (et PWIEquationTokenV2) validate() error {
	var requiredFields = []string{"bid", "ask"}
	// validates that both bid and ask are present
	for _, field := range requiredFields {
		if _, ok := et[field]; !ok {
			return fmt.Errorf("missing required field: %s", field)
		}
	}
	return nil
}

// PWIEquationRequestV2 is the input SetPWIEquationV2 api.
type PWIEquationRequestV2 map[string]PWIEquationTokenV2

// Validate returns nil if the input is valid, otherwise it will
// return an error with detail which field is invalid.
// Example input:
// [{"token_id": {equation_token}}, ...]
func (input PWIEquationRequestV2) Validate() error {
	for tokenID, et := range input {
		if err := et.validate(); err != nil {
			return fmt.Errorf("invalid equation for token %s: %s", tokenID, err)
		}

	}
	return nil
}

//RebalanceQuadraticEquation represent an equation
type RebalanceQuadraticEquation struct {
	RebalanceQuadratic struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
		C float64 `json:"c"`
	} `json:"rebalance_quadratic"`
}

//RebalanceQuadraticRequest represent data request to set rebalance quadratic
//map[token]equation
type RebalanceQuadraticRequest map[string]RebalanceQuadraticEquation
