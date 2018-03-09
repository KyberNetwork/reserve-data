package blockchain

import (
	"context"
	"log"
	"math/big"
	"os"
	"time"

	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func (self *Blockchain) getIntermediateTransactOpts(nonce, gasPrice *big.Int) (*bind.TransactOpts, context.CancelFunc, error) {
	shared := self.intermediateSigner.GetTransactOpts()
	var err error
	if nonce == nil {
		nonce, err = getNextNonce(self.nonceIntermediate)
	}
	if err != nil {
		return nil, donothing, err
	}
	if gasPrice == nil {
		gasPrice = big.NewInt(50100000000)
	}
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	result := bind.TransactOpts{
		shared.From,
		nonce,
		shared.Signer,
		shared.Value,
		gasPrice,
		shared.GasLimit,
		timeout,
	}
	return &result, cancel, nil
}

func packData(method string, params ...interface{}) ([]byte, error) {
	file, err := os.Open(
		"/go/src/github.com/KyberNetwork/reserve-data/blockchain/ERC20.abi")
	if err != nil {
		return nil, err
	}
	packabi, err := abi.JSON(file)
	if err != nil {
		return nil, err
	}
	data, err := packabi.Pack(method, params...)
	if err != nil {
		log.Println("Intermediator: Can not pack the data")
		return nil, err
	}
	var v interface{}
	err = packabi.Unpack(v, "transfer", data)
	if err != nil {
		log.Printf("Intermediator: can not unpack the data, %v", err)
	} else {
		log.Printf("Intermediator: v is %v \n\n", v)
	}
	return data, nil
}

func (self *Blockchain) SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address, tokenAddress ethereum.Address) (*types.Transaction, error) {
	opts, cancel, err := self.getIntermediateTransactOpts(nil, nil)
	ctx := opts.Context
	defer cancel()
	//build msg and get gas limit
	log.Printf("Intermediator: Amount is %d", amount)
	data, err := packData("transfer", exchangeAddress, amount)
	if err != nil {
		return nil, err
	}

	log.Printf("Intermediator: Data is %v", data)
	log.Printf("Intermediator: Hex form of the data is %x", data)
	msg := ether.CallMsg{From: opts.From, To: &tokenAddress, Value: big.NewInt(0), Data: data}
	log.Printf("msg in hex is  %x", msg)
	gasLimit, err := self.client.EstimateGas(ensureContext(opts.Context), msg)
	if err != nil {
		log.Printf("Intermediator: Can not estimate gas %v", err)
		gasLimit = big.NewInt(25000)
	} else {
		log.Println("Intermediator: gas limit estimated is : %d", gasLimit)
		if (gasLimit.Cmp(big.NewInt(25000))) < 0 {
			gasLimit = big.NewInt(25000)
		}
	}
	//build tx, sign and send

	tx := types.NewTransaction(opts.Nonce.Uint64(), tokenAddress, big.NewInt(0), gasLimit, opts.GasPrice, data)
	signTX, err := self.intermediateSigner.Sign(tx)
	if err != nil {
		log.Println("Intermediator: Can not sign the transaction")
		return nil, err
	}
	err = self.client.SendTransaction(ctx, signTX)
	if err != nil {
		log.Println("Intermediator: ERROR: Can't send the transaction")
		return nil, err
	}
	return signTX, nil
}

func (self *Blockchain) SendETHFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address) (*types.Transaction, error) {

	opts, cancel, err := self.getIntermediateTransactOpts(nil, nil)
	ctx := opts.Context
	defer cancel()
	//build msg and get gas limit
	msg := ether.CallMsg{From: opts.From, To: &exchangeAddress, Value: amount, Data: nil}
	gasLimit, err := self.client.EstimateGas(ensureContext(opts.Context), msg)
	//build tx, sign and send
	tx := types.NewTransaction(opts.Nonce.Uint64(), exchangeAddress, amount, gasLimit, opts.GasPrice, nil)
	signTX, err := self.intermediateSigner.Sign(tx)
	if err != nil {
		log.Println("Intermediator: Can not sign the transaction")
		return nil, err
	}
	log.Printf("Intermediator: the signed TX is: \n%v ", signTX)
	err = self.client.SendTransaction(ctx, signTX)
	if err != nil {
		log.Println("ERROR: Can't send the transaction")
		return nil, err
	}
	return signTX, nil
}
