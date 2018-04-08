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

	"github.com/KyberNetwork/reserve-data/common"
	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

const (
	ZeroAddress string = "0x0000000000000000000000000000000000000000"
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
	ethRate        EthUSDRate
	chainType      string
	contractCaller *ContractCaller
	erc20abi       abi.ABI
}

func (self *BaseBlockchain) OperatorAddresses() map[string]ethereum.Address {
	result := map[string]ethereum.Address{}
	for name, op := range self.operators {
		result[name] = op.Address
	}
	return result
}

func (self *BaseBlockchain) RegisterOperator(name string, op *Operator) {
	if _, found := self.operators[name]; found {
		panic(fmt.Sprintf("Operator name %s already exist", name))
	}
	self.operators[name] = op
}

func (self *BaseBlockchain) GetOperator(name string) *Operator {
	op, found := self.operators[name]
	if !found {
		panic(fmt.Sprintf("operator %s is not found. you have to register it before using it", name))
	}
	return op
}

func (self *BaseBlockchain) GetMinedNonce(operator string) (uint64, error) {
	nonce, err := self.GetOperator(operator).NonceCorpus.MinedNonce(self.client)
	if err != nil {
		return 0, err
	} else {
		return nonce.Uint64(), err
	}
}

func (self *BaseBlockchain) GetNextNonce(operator string) (*big.Int, error) {
	n := self.GetOperator(operator).NonceCorpus
	var nonce *big.Int
	var err error
	for i := 0; i < 3; i++ {
		nonce, err = n.GetNextNonce(self.client)
		if err == nil {
			return nonce, nil
		}
	}
	return nonce, err
}

func (self *BaseBlockchain) SignAndBroadcast(tx *types.Transaction, from string) (*types.Transaction, error) {
	signer := self.GetOperator(from).Signer
	if tx == nil {
		panic(errors.New("Nil tx is forbidden here"))
	} else {
		signedTx, err := signer.Sign(tx)
		if err != nil {
			return nil, err
		}
		failures, ok := self.broadcaster.Broadcast(signedTx)
		log.Printf("Rebroadcasting failures: %s", failures)
		if !ok {
			log.Printf("Broadcasting transaction failed!!!!!!!, err: %s, retry failures: %s", err, failures)
			if signedTx != nil {
				return signedTx, errors.New(fmt.Sprintf("Broadcasting transaction %s failed, err: %s, retry failures: %s", tx.Hash().Hex(), err, failures))
			} else {
				return signedTx, errors.New(fmt.Sprintf("Broadcasting transaction failed, err: %s, retry failures: %s", err, failures))
			}
		} else {
			return signedTx, nil
		}
	}
}

func (self *BaseBlockchain) Call(timeOut time.Duration, opts CallOpts, contract *Contract, result interface{}, method string, params ...interface{}) error {
	// Pack the input, call and unpack the results
	input, err := contract.ABI.Pack(method, params...)
	if err != nil {
		return err
	}
	var (
		msg    = ether.CallMsg{From: ethereum.HexToAddress(ZeroAddress), To: &contract.Address, Data: input}
		code   []byte
		output []byte
	)
	if opts.Block == nil || opts.Block.Cmp(ethereum.Big0) == 0 {
		// calling in pending state
		output, err = self.contractCaller.CallContract(msg, nil, timeOut)
	} else {
		output, err = self.contractCaller.CallContract(msg, opts.Block, timeOut)
	}
	if err == nil && len(output) == 0 {
		ctx := context.Background()
		// Make sure we have a contract to operate on, and bail out otherwise.
		if opts.Block == nil || opts.Block.Cmp(ethereum.Big0) == 0 {
			code, err = self.client.CodeAt(ctx, contract.Address, nil)
		} else {
			code, err = self.client.CodeAt(ctx, contract.Address, opts.Block)
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

func (self *BaseBlockchain) BuildTx(context context.Context, opts TxOpts, contract *Contract, method string, params ...interface{}) (*types.Transaction, error) {
	input, err := contract.ABI.Pack(method, params...)
	if err != nil {
		return nil, err
	}
	return self.transactTx(context, opts, contract.Address, input)
}

func (self *BaseBlockchain) transactTx(context context.Context, opts TxOpts, contract ethereum.Address, input []byte) (*types.Transaction, error) {
	var err error
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
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := opts.GasLimit
	if gasLimit == nil {
		// Gas estimation cannot succeed without code for method invocations
		if contract.Big().Cmp(ethereum.Big0) == 0 {
			if code, err := self.client.PendingCodeAt(ensureContext(context), contract); err != nil {
				return nil, err
			} else if len(code) == 0 {
				return nil, bind.ErrNoCode
			}
		}
		// If the contract surely has code (or code is not needed), estimate the transaction
		msg := ether.CallMsg{From: opts.Operator.Address, To: &contract, Value: value, Data: input}
		gasLimit, err = self.client.EstimateGas(ensureContext(context), msg)
		if err != nil {
			return nil, fmt.Errorf("failed to estimate gas needed: %v", err)
		}
		// add gas limit by 50K gas
		gasLimit.Add(gasLimit, big.NewInt(50000))
	}
	// Create the transaction, sign it and schedule it for execution
	var rawTx *types.Transaction
	if contract.Big().Cmp(ethereum.Big0) == 0 {
		rawTx = types.NewContractCreation(nonce, value, gasLimit, opts.GasPrice, input)
	} else {
		rawTx = types.NewTransaction(nonce, contract, value, gasLimit, opts.GasPrice, input)
	}
	return rawTx, nil
}

func (self *BaseBlockchain) GetCallOpts(block uint64) CallOpts {
	var blockBig *big.Int
	if block != 0 {
		blockBig = big.NewInt(int64(block))
	}
	return CallOpts{
		Block: blockBig,
	}
}

func (self *BaseBlockchain) GetTxOpts(op string, nonce *big.Int, gasPrice *big.Int, value *big.Int) (TxOpts, error) {
	result := TxOpts{}
	operator := self.GetOperator(op)
	var err error
	if nonce == nil {
		nonce, err = self.GetNextNonce(op)
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
	result.GasLimit = nil
	return result, nil
}

func (self *BaseBlockchain) GetLogs(param ether.FilterQuery) ([]types.Log, error) {
	result := []types.Log{}
	// log.Printf("LogFetcher - fetching logs data from block %d, to block %d", opts.Block, to.Uint64())
	err := self.rpcClient.Call(&result, "eth_getLogs", toFilterArg(param))
	return result, err
}

func (self *BaseBlockchain) CurrentBlock() (uint64, error) {
	var blockno string
	err := self.rpcClient.Call(&blockno, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseUint(blockno, 0, 64)
	return result, err
}

func (self *BaseBlockchain) PackERC20Data(method string, params ...interface{}) ([]byte, error) {
	return self.erc20abi.Pack(method, params...)
}

func (self *BaseBlockchain) BuildSendERC20Tx(opts TxOpts, amount *big.Int, to ethereum.Address, tokenAddress ethereum.Address) (*types.Transaction, error) {
	var err error
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
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	data, err := self.PackERC20Data("transfer", to, amount)
	if err != nil {
		return nil, err
	}
	msg := ether.CallMsg{From: opts.Operator.Address, To: &tokenAddress, Value: value, Data: data}
	timeout, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	gasLimit, err := self.client.EstimateGas(timeout, msg)
	if err != nil {
		log.Printf("Cannot estimate gas limit: %v", err)
		return nil, err
	}
	gasLimit.Add(gasLimit, big.NewInt(50000))
	rawTx := types.NewTransaction(nonce, tokenAddress, value, gasLimit, opts.GasPrice, data)
	return rawTx, nil
}

func (self *BaseBlockchain) BuildSendETHTx(opts TxOpts, amount *big.Int, to ethereum.Address) (*types.Transaction, error) {
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
	if opts.GasPrice == nil {
		return nil, errors.New("gas price must be specified")
	}
	gasLimit := big.NewInt(50000)
	rawTx := types.NewTransaction(nonce, to, value, gasLimit, opts.GasPrice, nil)
	return rawTx, nil
}

func (self *BaseBlockchain) TransactionByHash(ctx context.Context, hash ethereum.Hash) (tx *rpcTransaction, isPending bool, err error) {
	var json *rpcTransaction
	err = self.rpcClient.CallContext(ctx, &json, "eth_getTransactionByHash", hash)
	if err != nil {
		return nil, false, err
	} else if json == nil {
		return nil, false, ether.NotFound
	} else if _, r, _ := json.tx.RawSignatureValues(); r == nil {
		return nil, false, fmt.Errorf("server returned transaction without signature")
	}
	setSenderFromServer(json.tx, json.From, json.BlockHash)
	return json, json.BlockNumber().Cmp(ethereum.Big0) == 0, nil
}

func (self *BaseBlockchain) TxStatus(hash ethereum.Hash) (string, uint64, error) {
	option := context.Background()
	tx, pending, err := self.TransactionByHash(option, hash)
	if err == nil {
		// tx exist
		if pending {
			return "", 0, nil
		} else {
			receipt, err := self.client.TransactionReceipt(option, hash)
			if err != nil {
				// incompatibily between geth and parity
				// so even err is not nil, receipt is still there
				// and have valid fields
				if receipt != nil {
					// only byzantium has status field at the moment
					// mainnet, ropsten are byzantium, other chains such as
					// devchain, kovan are not
					if self.chainType == "byzantium" {
						if receipt.Status == 1 {
							// successful tx
							return "mined", tx.BlockNumber().Uint64(), nil
						} else {
							// failed tx
							return "failed", tx.BlockNumber().Uint64(), nil
						}
					} else {
						return "mined", tx.BlockNumber().Uint64(), nil
					}
				} else {
					// networking issue
					return "", 0, err
				}
			} else {
				if receipt.Status == 1 {
					// successful tx
					return "mined", tx.BlockNumber().Uint64(), nil
				} else {
					// failed tx
					return "failed", tx.BlockNumber().Uint64(), nil
				}
			}
		}
	} else {
		if err == ether.NotFound {
			// tx doesn't exist. it failed
			return "lost", 0, nil
		} else {
			// networking issue
			return "", 0, err
		}
	}
}

func (self *BaseBlockchain) GetEthRate(timepoint uint64) float64 {
	rate := self.ethRate.GetUSDRate(timepoint)
	log.Printf("ETH-USD rate: %f", rate)
	return rate
}

func NewMinimalBaseBlockchain(
	endpoints []string, operators map[string]*Operator, chainType string) (*BaseBlockchain, error) {

	if len(endpoints) == 0 {
		return nil, errors.New("At least one endpoint is required to init a blockchain")
	}
	endpoint := endpoints[0]
	rpcClient, err := rpc.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	ethClient := ethclient.NewClient(rpcClient)
	bkclients := map[string]*ethclient.Client{}
	callClients := []*ethclient.Client{}
	for _, ep := range endpoints {
		bkclient, err := ethclient.Dial(ep)
		if err != nil {
			log.Printf("Cannot connect to %s, err %s. Ignore it.", ep, err)
		} else {
			bkclients[ep] = bkclient
			callClients = append(callClients, bkclient)
		}
	}
	return NewBaseBlockchain(
		rpcClient, ethClient, operators,
		NewBroadcaster(bkclients),
		NewCMCEthUSDRate(),
		chainType,
		NewContractCaller(callClients, endpoints),
	), nil
}

func NewBaseBlockchain(
	rpcClient *rpc.Client,
	client *ethclient.Client,
	operators map[string]*Operator,
	broadcaster *Broadcaster,
	ethRate EthUSDRate,
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
		ethRate:        ethRate,
		chainType:      chainType,
		erc20abi:       packabi,
		contractCaller: contractcaller,
	}
}
