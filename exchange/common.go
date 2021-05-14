package exchange

import (
	"strconv"

	blockchain "github.com/KyberNetwork/reserve-data/exchange/blockchain"
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

func isOPAvailable(blockchain *blockchain.Blockchain, opName string) bool {
	return blockchain.GetOperator(opName) != nil
}
