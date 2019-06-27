package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Setting interface {
	GetExchangeStatus() (common.ExchangesStatus, error)
	UpdateExchangeStatus(common.ExchangesStatus) error
}
