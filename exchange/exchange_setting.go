package exchange

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Setting interface {
	GetInternalTokenByID(tokenID string) (common.Token, error)
	GetInternalTokens() ([]common.Token, error)
	GetAllTokens() ([]common.Token, error)
	GetTokenByID(tokenID string) (common.Token, error)
	GetFee(ex settings.ExchangeName) (common.ExchangeFees, error)
	GetMinDeposit(ex settings.ExchangeName) (common.ExchangesMinDeposit, error)
	GetDepositAddresses(ex settings.ExchangeName) (common.ExchangeAddresses, error)
	GetAddress(name settings.AddressName) (ethereum.Address, error)
	UpdateDepositAddress(name settings.ExchangeName, addrs common.ExchangeAddresses) error
	GetExchangeInfo(ex settings.ExchangeName) (common.ExchangeInfo, error)
	UpdateExchangeInfo(ex settings.ExchangeName, exInfo common.ExchangeInfo) error
}
