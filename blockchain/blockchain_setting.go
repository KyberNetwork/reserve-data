package blockchain

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Setting interface {
	GetInternalTokens() ([]common.Token, error)
	ETHToken() common.Token
	GetAddress(settings.AddressName) (ethereum.Address, error)
}
