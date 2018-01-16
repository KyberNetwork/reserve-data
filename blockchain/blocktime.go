package blockchain

import (
	"time"
)

func InterpretTimestamp(blockno uint64, txindex uint) uint64 {
	// TODO: Using real timestamp
	return uint64(time.Second)*15*blockno + uint64(txindex)
}
