package settings

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type TokenStorage interface {
	//Add Tokens by ID
	AddTokenByID(common.Token) error
	AddInternalTokenByID(common.Token) error
	AddExternalTokenByID(common.Token) error
	AddActiveTokenByID(common.Token) error
	//Add Tokens by Address
	AddTokenByAddress(common.Token) error
	AddInternalTokenByAddress(common.Token) error
	AddExternalTokenByAddress(common.Token) error
	AddActiveTokenByAddress(common.Token) error

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
}
