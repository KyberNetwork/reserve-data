package blockchain

import (
	"math/big"
)

type TxOpts struct {
	Operator *Operator // Ethereum account to send the transaction from
	Nonce    *big.Int  // Nonce to use for the transaction execution (nil = use pending state)

	Value    *big.Int // Funds to transfer along along the transaction (nil = 0 = no funds)
	GasPrice *big.Int // Gas price to use for the transaction execution (nil = gas price oracle)
	GasLimit *big.Int // Gas limit to set for the transaction execution (0 = estimate)
}

type CallOpts struct {
	Block *big.Int // Block number that the call is invoked at. Nil means calling in pending state
}
