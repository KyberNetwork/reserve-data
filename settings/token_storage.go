package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

//TokenStorage defines a set of function abstracting the storage interface for Token
//All key for update and lookup inside the storage must be in lower key
type TokenStorage interface {
	//Add Tokens by ID
	AddTokenByID(common.Token, uint64) error
	//Add Tokens by Address
	AddTokenByAddress(common.Token, uint64) error
	//Update common
	UpdateToken(common.Token, uint64) error

	//Active Tokens (Network Tokens)
	GetActiveTokens() ([]common.Token, error)
	GetActiveTokenByID(id string) (common.Token, error)
	GetActiveTokenByAddress(ethereum.Address) (common.Token, error)

	//All Tokens (Supported tokens)
	GetAllTokens() ([]common.Token, error)
	GetTokenByID(id string) (common.Token, error)
	GetTokenByAddress(ethereum.Address) (common.Token, error)
	//Internal Active Tokens
	GetInternalTokens() ([]common.Token, error)
	GetInternalTokenByID(id string) (common.Token, error)
	GetInternalTokenByAddress(ethereum.Address) (common.Token, error)
	//External Active Tokens
	GetExternalTokens() ([]common.Token, error)
	GetExternalTokenByID(id string) (common.Token, error)
	GetExternalTokenByAddress(ethereum.Address) (common.Token, error)
	StorePendingTokenUpdates(map[string]common.TokenUpdate) error
	GetPendingTokenUpdates() (map[string]common.TokenUpdate, error)
	UpdateTokenWithExchangeSetting(t []common.Token, exSetting map[ExchangeName]*common.ExchangeSetting, availExs []ExchangeName, timestamp uint64) error
	RemovePendingTokenUpdates() error
	GetTokenVersion() (uint64, error)
}
