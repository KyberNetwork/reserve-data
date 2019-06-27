package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"

	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	zeroAddress string = "0x0000000000000000000000000000000000000000"
)

// BaseBlockchain interact with the blockchain in a way that eases
// other blockchain types in KyberNetwork.
// It manages multiple operators (address, signer and nonce)
// It has convenient logic of each operator nonce so users dont have
// to care about nonce management.
// It has convenient logic of broadcasting tx to multiple nodes at once.
// It has convenient functions to init proper CallOpts and TxOpts.
// It has eth usd rate lookup function.

type BaseBlockchain struct {
	client         *ethclient.Client
	rpcClient      *rpc.Client
	operators      map[string]*Operator
	broadcaster    *Broadcaster
	chainType      string
	contractCaller *ContractCaller
	erc20abi       abi.ABI
}

func (b *BaseBlockchain) OperatorAddresses() map[string]ethereum.Address {
	result := map[string]ethereum.Address{}
	for name, op := range b.operators {
		result[name] = op.Address
	}
	return result
}

func (b *BaseBlockchain) MustRegisterOperator(name string, op *Operator) {
	//This shouldn't happen, each operator get registered only once.
	if _, found := b.operators[name]; found {
		panic(fmt.Sprintf("Operator name %s already exist", name))
	}
	b.operators[name] = op
}

func (b *BaseBlockchain) RecommendedGasPriceFromNode() (*big.Int, error) {
	timeout, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	return b.client.SuggestGasPrice(timeout)
}

// MustGetOperator returns the operator if avail, panic if the operator can't be found
func (b *BaseBlockchain) MustGetOperator(name string) *Operator {
	op, found := b.operators[name]
	if !found {
		panic(fmt.Sprintf("operator %s is not found. you have to register it before using it", name))
	}
	return op
}

func (b *BaseBlockchain) GetMinedNonce(operator string) (uint64, error) {
	nonce, err := b.MustGetOperator(operator).NonceCorpus.MinedNonce(b.client)
	if err != nil {
		return 0, err
	}
	return nonce.Uint64(), err
}

func (b *BaseBlockchain) GetNextNonce(operator string) (*big.Int, error) {
	var nonce *big.Int
	n := b.MustGetOperator(operator).NonceCorpus
	var err error
	for i := 0; i < 3; i++ {
		nonce, err = n.GetNextNonce(b.client)
		if err == nil {
			return nonce, nil
		}
	}
	return nonce, err
}

func (b *BaseBlockchain) SignAndBroadcast(tx *types.Transaction, from string) (*types.Transaction, error) {
	signer := b.MustGetOperator(from).Signer
	if tx == nil {
		return nil, errors.New("nil tx is forbidden here")
	}
	signedTx, err := signer.Sign(tx)
	if err != nil {
		return nil, err
	}
	failures, ok := b.broadcaster.Broadcast(signedTx)
	log.Printf("Rebroadcasting failures: %s", failures)
	if !ok {
		log.Printf("Broadcasting transaction failed! nonce: %d, gas price: %s, retry failures: %s", tx.Nonce(), tx.GasPrice().Text(10), failures)
		if signedTx != nil {
			return signedTx, fmt.Errorf("broadcasting transaction %s failed, retry failures: %s", tx.Hash().Hex(), failures)
		}
		return signedTx, fmt.Errorf("broadcasting transaction failed, retry failures: %s", failures)
	}
	return signedTx, nil
}

func (b *BaseBlockchain) Call(timeOut time.Duration, opts CallOpts, contract *Contract, result interface{}, method string, params ...interface{}) error {
	// Pack the input, call and unpack the results
	input, err := contract.ABI.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = ether.CallMsg{From: ethereum.HexToAddress(zeroAddress), To: &contract.Address, Data: input}
		code   []byte
		output []byte
	)
	if opts.Block == nil || opts.Block.Cmp(ethereum.Big0) == 0 {
		// calling in pending state
		output, err = b.contractCaller.CallContract(msg, nil, timeOut)
	} else {
		output, err = b.contractCaller.CallContract(msg, opts.Block, timeOut)
	}
	if err == nil && len(output) == 0 {
		ctx := context.Background()
		// Make sure we have a contract to operate on, and bail out otherwise.
		if opts.Block == nil || opts.Block.Cmp(ethereum.Big0) == 0 {
			code, err = b.client.CodeAt(ctx, contract.Address, nil)
		} else {
			code, err = b.client.CodeAt(ctx, contract.Address, opts.Block)
		}
		if err != nil {
			return err
		} else if len(code) == 0 {
			return bind.ErrNoCode
		}
	}
	if err != nil {
		return err
	}
	return contract.ABI.Unpack(result, method, output)
}

func (b *BaseBlockchain) BuildTx(context context.Context, opts TxOpts, contract *Contract, method string, params ...interface{}) (*types.Transaction, error) {
	input, err := contract.ABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	return b.transactTx(context, opts, contract.Address, input)
}

func (b *BaseBlockchain) transactTx(context context.Context, opts TxOpts, contract ethereum.Address, input []byte) (*types.Transaction, error) {
	var err error
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		return nil, errors.New("nonce must be specified")
	}
	nonce = opts.Nonce.Uint64()
	// Figure out the gas allowance and gas price values
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := opts.GasLimit
	if gasLimit == 0 {
		// Gas estimation cannot succeed without code for method invocations
		if contract.Hash().Big().Cmp(ethereum.Big0) == 0 {
			if code, pErr := b.client.PendingCodeAt(ensureContext(context), contract); pErr != nil {
				return nil, pErr
			} else if len(code) == 0 {
				return nil, bind.ErrNoCode
			}
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ether.CallMsg{From: opts.Operator.Address, To: &contract, Value: value, Data: input}
		gasLimit, err = b.client.EstimateGas(ensureContext(context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
		// add gas limit by 50K gas
		gasLimit += 50000
	}
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if contract.Hash().Big().Cmp(ethereum.Big0) == 0 {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, opts.GasPrice, input)
	} else {
		rawTx = types.NewTransaction(nonce, contract, value, gasLimit, opts.GasPrice, input)
	}
	return rawTx, nil
}

func (b *BaseBlockchain) GetCallOpts(block uint64) CallOpts {
	var blockBig *big.Int
	if block != 0 {
		blockBig = big.NewInt(int64(block))
	}
	return CallOpts{
		Block: blockBig,
	}
}

func (b *BaseBlockchain) GetTxOpts(op string, nonce *big.Int, gasPrice *big.Int, value *big.Int) (TxOpts, error) {
	result := TxOpts{}
	operator := b.MustGetOperator(op)
	var err error
	if nonce == nil {
		nonce, err = b.GetNextNonce(op)
	}
	if err != nil {
		return result, err
	}
	if gasPrice == nil {
		gasPrice = big.NewInt(50100000000)
	}
	// timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	result.Operator = operator
	result.Nonce = nonce
	result.Value = value
	result.GasPrice = gasPrice
	result.GasLimit = 0
	return result, nil
}

func (b *BaseBlockchain) GetLogs(param ether.FilterQuery) ([]types.Log, error) {
	var result []types.Log
	err := b.rpcClient.Call(&result, "eth_getLogs", toFilterArg(param))
	return result, err
}

func (b *BaseBlockchain) CurrentBlock() (uint64, error) {
	var blockno string
	err := b.rpcClient.Call(&blockno, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseUint(blockno, 0, 64)
	return result, err
}

func (b *BaseBlockchain) PackERC20Data(method string, params ...interface{}) ([]byte, error) {
	return b.erc20abi.Pack(method, params...)
}

func (b *BaseBlockchain) BuildSendERC20Tx(opts TxOpts, amount *big.Int, to ethereum.Address, tokenAddress ethereum.Address) (*types.Transaction, error) {
	var err error
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		return nil, errors.New("nonce must be specified")
	}
	nonce = opts.Nonce.Uint64()

	// Figure out the gas allowance and gas price values
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	data, err := b.PackERC20Data("transfer", to, amount)
	if err != nil {
		return nil, err
	}
	msg := ether.CallMsg{From: opts.Operator.Address, To: &tokenAddress, Value: value, Data: data}
	timeout, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()
	gasLimit, err := b.client.EstimateGas(timeout, msg)
	if err != nil {
		log.Printf("Cannot estimate gas limit: %v", err)
		return nil, err
	}
	gasLimit += 50000
	rawTx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, opts.GasPrice, data)
	return rawTx, nil
}

func (b *BaseBlockchain) BuildSendETHTx(opts TxOpts, to ethereum.Address) (*types.Transaction, error) {
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	var nonce uint64
	if opts.Nonce == nil {
		return nil, errors.New("nonce must be specified")
	}
	nonce = opts.Nonce.Uint64()
	// Figure out the gas allowance and gas price values
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := uint64(50000)
	rawTx := types.NewTransaction(nonce, to, value, gasLimit, opts.GasPrice, nil)
	return rawTx, nil
}

func (b *BaseBlockchain) TransactionByHash(ctx context.Context, hash ethereum.Hash) (tx *RPCTransaction, isPending bool, err error) {
	var json *RPCTransaction
	err = b.rpcClient.CallContext(ctx, &json, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if json == nil {
		return nil, false, ether.NotFound
	} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
		return nil, false, errors.New("server returned transaction without signature")
	}
	setSenderFromServer(json.tx, json.From, json.BlockHash)
	return json, json.BlockNumber().Cmp(ethereum.Big0) == 0, nil
}

func (b *BaseBlockchain) TxStatus(hash ethereum.Hash) (string, uint64, error) {
	option := context.Background()
	tx, pending, err := b.TransactionByHash(option, hash)
	if err != nil {
		if err == ether.NotFound {
			// tx doesn't exist. it failed
			return common.MiningStatusLost, 0, nil
		}
		// networking issue
		return "", 0, err
	}
	// tx exist
	if pending {
		return common.MiningStatusPending, 0, nil
	}
	var receipt *types.Receipt
	receipt, err = b.client.TransactionReceipt(option, hash)
	if err != nil {
		// incompatibily between geth and parity
		// so even err is not nil, receipt is still there
		// and have valid fields
		if receipt != nil {
			// only byzantium has status field at the moment
			// mainnet, ropsten are byzantium, other chains such as
			// devchain, kovan are not
			if b.chainType == "byzantium" {
				if receipt.Status == 1 {
					// successful tx
					return common.MiningStatusMined, tx.BlockNumber().Uint64(), nil
				}
				// failed tx
				return common.MiningStatusFailed, tx.BlockNumber().Uint64(), nil
			}
			return common.MiningStatusMined, tx.BlockNumber().Uint64(), nil
		}
		// networking issue
		return "", 0, err
	}
	if receipt.Status == 1 {
		// successful tx
		return common.MiningStatusMined, tx.BlockNumber().Uint64(), nil
	}
	// failed tx
	return common.MiningStatusFailed, tx.BlockNumber().Uint64(), nil
}

func NewBaseBlockchain(
	rpcClient *rpc.Client,
	client *ethclient.Client,
	operators map[string]*Operator,
	broadcaster *Broadcaster,
	chainType string,
	contractcaller *ContractCaller) *BaseBlockchain {

	file, err := os.Open(
		filepath.Join(common.CurrentDir(), "ERC20.abi"))
	if err != nil {
		panic(err)
	}
	packabi, err := abi.JSON(file)
	if err != nil {
		panic(err)
	}

	return &BaseBlockchain{
		client:         client,
		rpcClient:      rpcClient,
		operators:      operators,
		broadcaster:    broadcaster,
		chainType:      chainType,
		erc20abi:       packabi,
		contractCaller: contractcaller,
	}
}
