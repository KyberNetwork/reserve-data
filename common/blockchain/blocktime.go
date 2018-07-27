package blockchain

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

var CachedBlockno uint64
var CachedBlockHeader *types.Header

func (bb *BaseBlockchain) InterpretTimestamp(blockno uint64, txindex uint) (uint64, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var block *types.Header
	var err error
	if CachedBlockno == blockno {
		block = CachedBlockHeader
	} else {
		block, err = bb.client.HeaderByNumber(timeout, big.NewInt(int64(blockno)))
	}
	if err != nil {
		if block == nil {
			return uint64(0), err
		} else {
			// error because parity and geth are not compatible in mix hash
			// so we ignore it
			CachedBlockno = blockno
			CachedBlockHeader = block
			err = nil
		}
	}
	unixSecond := block.Time.Uint64()
	unixNano := uint64(time.Unix(int64(unixSecond), 0).UnixNano())
	result := unixNano + uint64(txindex)
	return result, nil
}
