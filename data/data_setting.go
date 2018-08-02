package data

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
)

type Setting interface {
	GetInternalTokenByID(tokenID string) (common.Token, error)
	GetExchangeStatus() (common.ExchangesStatus, error)
	UpdateExchangeStatus(data common.ExchangesStatus) error
	UpdateExchangeNotification(exchange, action, tokenPair string, fromTime, toTime uint64, isWarning bool, msg string) error
	GetExchangeNotifications() (common.ExchangeNotifications, error)
	GetInternalTokens() ([]common.Token, error)
	GetDepositAddresses(settings.ExchangeName) (common.ExchangeAddresses, error)
}
