package http

import (
	ethereum "github.com/ethereum/go-ethereum/common"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
)

type Setting interface {
	GetInternalTokenByID(tokenID string) (common.Token, error)
	GetActiveTokenByID(tokenID string) (common.Token, error)
	GetTokenByID(tokenID string) (common.Token, error)
	GetInternalTokens() ([]common.Token, error)
	GetAllTokens() ([]common.Token, error)
	NewTokenPairFromID(base, quote string) (common.TokenPair, error)
	GetFee(ex settings.ExchangeName) (common.ExchangeFees, error)
	UpdateFee(ex settings.ExchangeName, data common.ExchangeFees, timestamp uint64) error
	GetMinDeposit(ex settings.ExchangeName) (common.ExchangesMinDeposit, error)
	UpdateMinDeposit(ex settings.ExchangeName, minDeposit common.ExchangesMinDeposit, timestamp uint64) error
	GetDepositAddresses(ex settings.ExchangeName) (common.ExchangeAddresses, error)
	UpdateDepositAddress(ex settings.ExchangeName, addrs common.ExchangeAddresses, timestamp uint64) error
	GetExchangeInfo(ex settings.ExchangeName) (common.ExchangeInfo, error)
	UpdateExchangeInfo(ex settings.ExchangeName, exInfo common.ExchangeInfo, timestamp uint64) error
	GetExchangeStatus() (common.ExchangesStatus, error)
	UpdateExchangeStatus(data common.ExchangesStatus) error
	UpdatePendingTokenUpdates(map[string]common.TokenUpdate) error
	ApplyTokenWithExchangeSetting([]common.Token, map[settings.ExchangeName]*common.ExchangeSetting, uint64) error
	GetPendingTokenUpdates() (map[string]common.TokenUpdate, error)
	RemovePendingTokenUpdates() error
	GetTokenVersion() (uint64, error)
	GetExchangeVersion() (uint64, error)
	GetActiveTokens() ([]common.Token, error)
	GetTokenByAddress(ethereum.Address) (common.Token, error)
}
