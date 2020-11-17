package exchange

import (
	"strconv"
)

const (
	tradeTypeBuy  = "buy"
	tradeTypeSell = "sell"
)

func remainingQty(orgQty, executedQty string) (float64, error) {
	oAmount, err := strconv.ParseFloat(orgQty, 64)
	if err != nil {
		return 0, err
	}
	oExecutedQty, err := strconv.ParseFloat(executedQty, 64)
	if err != nil {
		return 0, err
	}
	return oAmount - oExecutedQty, nil
}
