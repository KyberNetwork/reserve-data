package blockchain

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// Setting interface for blockchain package
type Setting interface {
	GetInternalTokens() ([]common.Token, error)
	ETHToken() common.Token
	GetAddress(settings.AddressName) (ethereum.Address, error)
}
