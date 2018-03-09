package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"

	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type KNContractBase struct {
	address ethereum.Address
	abi     abi.ABI
	client  *ethclient.Client
}

// ensureContext is a helper method to ensure a context is not nil, even if the
// user specified it as such.
func ensureContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.TODO()
	}
	return ctx
}

func (self *KNContractBase) Call(opts *bind.CallOpts, atBlock *big.Int, result interface{}, method string, params ...interface{}) error {
	// Don't crash on a lazy user
	if opts == nil {
		opts = new(bind.CallOpts)
	}
	// Pack the input, call and unpack the results
	input, err := self.abi.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = ether.CallMsg{From: opts.From, To: &self.address, Data: input}
		ctx    = ensureContext(opts.Context)
		code   []byte
		output []byte
	)
	// not support pending calling yet
	output, err = self.client.CallContract(ctx, msg, atBlock)
	if err == nil && len(output) == 0 {
		// Make sure we have a contract to operate on, and bail out otherwise.
		if code, err = self.client.CodeAt(ctx, self.address, nil); err != nil {
			return err
		} else if len(code) == 0 {
			return bind.ErrNoCode
		}
	}
	if err != nil {
		return err
	}
	return self.abi.Unpack(result, method, output)
}

func (self *KNContractBase) BuildTx(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	input, err := self.abi.Pack(method, params...)
	log.Printf("Intermediator: messagie is %x", input)
	if err != nil {
		return nil, err
	}
	return self.transactTx(opts, &self.address, input)
}

func (self *KNContractBase) transactTx(opts *bind.TransactOpts, contract *ethereum.Address, input []byte) (*types.Transaction, error) {
	var err error
	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		return nil, errors.New("nonce must be specified")
	} else {
		nonce = opts.Nonce.Uint64()
	}
	// Figure out the gas allowance and gas price values
	gasPrice := opts.GasPrice
	if gasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := opts.GasLimit
	if gasLimit == nil {
		// Gas estimation cannot succeed without code for method invocations
		if contract != nil {
			if code, err := self.client.PendingCodeAt(ensureContext(opts.Context), self.address); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, bind.ErrNoCode
			}
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ether.CallMsg{From: opts.From, To: contract, Value: value, Data: input}
		gasLimit, err = self.client.EstimateGas(ensureContext(opts.Context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
		// add gas limit by 50K gas
		gasLimit.Add(gasLimit, big.NewInt(50000))
	}
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if contract == nil {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, gasPrice, input)
	} else {
		rawTx = types.NewTransaction(nonce, self.address, value, gasLimit, gasPrice, input)
	}
	return rawTx, nil
}

func NewKNContractBase(address ethereum.Address, abi abi.ABI, client *ethclient.Client) *KNContractBase {
	return &KNContractBase{
		address, abi, client,
	}
}
