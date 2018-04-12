package http

import "github.com/KyberNetwork/reserve-data/common"

type Huobi interface {
	PendingIntermediateTxs() (map[common.ActivityID]common.TXEntry, error)
}
