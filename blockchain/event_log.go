package blockchain

import (
	ethereum "github.com/ethereum/go-ethereum/common"
)

func LogDataToTradeParams(data []byte) (ethereum.Address, ethereum.Address, ethereum.Hash, ethereum.Hash) {
	srcAddr := ethereum.BytesToAddress(data[0:32])
	desAddr := ethereum.BytesToAddress(data[32:64])
	srcAmount := ethereum.BytesToHash(data[64:96])
	desAmount := ethereum.BytesToHash(data[96:128])
	return srcAddr, desAddr, srcAmount, desAmount
}

func LogDataToFeeWalletParams(data []byte) (ethereum.Address, ethereum.Address, ethereum.Hash) {
	reserveAddr := ethereum.BytesToAddress(data[0:32])
	walletAddr := ethereum.BytesToAddress(data[32:64])
	walletFee := ethereum.BytesToHash(data[64:96])
	return reserveAddr, walletAddr, walletFee
}

func LogDataToBurnFeeParams(data []byte) (ethereum.Address, ethereum.Hash) {
	reserveAddr := ethereum.BytesToAddress(data[0:32])
	burnFees := ethereum.BytesToHash(data[32:64])
	return reserveAddr, burnFees
}
