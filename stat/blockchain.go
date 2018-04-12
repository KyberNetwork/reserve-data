package stat

import (
	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Blockchain interface {
	CurrentBlock() (uint64, error)
	GetLogs(fromBlock uint64, toBlock uint64) ([]common.KNLog, error)
	GetReserveRates(atBlock, currentBlock uint64, reserveAddress ethereum.Address, tokens []common.Token) (common.ReserveRates, error)
}
