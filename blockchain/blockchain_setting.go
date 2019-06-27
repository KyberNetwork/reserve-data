package blockchain

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Setting interface {
	GetInternalTokens() ([]common.Token, error)
	ETHToken() common.Token
}
