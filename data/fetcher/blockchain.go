package fetcher

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ethereum "github.com/ethereum/go-ethereum/common"
)

// Blockchain contains all methods for fetcher to interact with blockchain.
type Blockchain interface {
	FetchBalanceData(addr ethereum.Address, atBlock uint64) (map[common.AssetID]common.BalanceEntry, error)
	// fetch current raw rates at specific block
	FetchRates(atBlock uint64, currentBlock uint64) (common.AllRateEntry, error)
	TxStatus(tx ethereum.Hash) (string, uint64, error)
	CurrentBlock() (uint64, error)
	blockchain.MinedNoncePicker
}
