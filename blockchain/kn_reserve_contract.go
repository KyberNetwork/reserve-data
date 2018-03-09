package blockchain

import (
	"log"
	"math/big"
	"os"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type KNReserveContract struct {
	*KNContractBase
}

func (self *KNReserveContract) Withdraw(opts *bind.TransactOpts, token ethereum.Address, amount *big.Int, destination ethereum.Address) (*types.Transaction, error) {
	log.Printf("Intermediator: Withdrawing...")
	return self.KNContractBase.BuildTx(opts, "withdraw", token, amount, destination)
}

func NewKNReserveContract(address ethereum.Address, client *ethclient.Client) (*KNReserveContract, error) {
	file, err := os.Open(
		"/go/src/github.com/KyberNetwork/reserve-data/blockchain/reserve.abi")
	if err != nil {
		return nil, err
	}
	parsed, err := abi.JSON(file)
	if err != nil {
		return nil, err
	}
	return &KNReserveContract{
		NewKNContractBase(address, parsed, client),
	}, nil
}
