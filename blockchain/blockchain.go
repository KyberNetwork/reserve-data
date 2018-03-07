package blockchain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type tbindex struct {
	BulkIndex   uint64
	IndexInBulk uint64
}

const (
	FeeToWalletEvent string = "0x366bc34352215bf0bd3b527cfd6718605e1f5938777e42bcd8ed92f578368f52"
	BurnFeeEvent     string = "0xf838f6ddc89706878e3c3e698e9b5cbfbf2c0e3d3dcd0bd2e00f1ccf313e0185"
	TradeEvent       string = "0x1849bd6a030a1bca28b83437fd3de96f3d27a5d172fa7e9c78e7b61468928a39"
	UserCatEvent     string = "0x0aeb0f7989a09b8cccf58cea1aefa196ccf738cb14781d6910448dd5649d0e6e"
)

var (
	Big0   *big.Int = big.NewInt(0)
	BigMax *big.Int = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(33), nil)
)

type Blockchain struct {
	rpcClient          *rpc.Client
	client             *ethclient.Client
	wrapper            *KNWrapperContract
	pricing            *KNPricingContract
	reserve            *KNReserveContract
	rm                 ethereum.Address
	wrapperAddr        ethereum.Address
	pricingAddr        ethereum.Address
	burnerAddr         ethereum.Address
	networkAddr        ethereum.Address
	whitelistAddr      ethereum.Address
	oldNetworks        []ethereum.Address
	oldBurners         []ethereum.Address
	signer             Signer
	depositSigner      Signer
	intermediateSigner Signer
	tokens             []common.Token
	tokenIndices       map[string]tbindex
	nonce              NonceCorpus
	nonceDeposit       NonceCorpus
	nonceIntermediate  NonceCorpus
	broadcaster        *Broadcaster
	chainType          string
}

func (self *Blockchain) AddOldNetwork(addr ethereum.Address) {
	self.oldNetworks = append(self.oldNetworks, addr)
}

func (self *Blockchain) AddOldBurners(addr ethereum.Address) {
	self.oldBurners = append(self.oldBurners, addr)
}

func (self *Blockchain) AddToken(t common.Token) {
	self.tokens = append(self.tokens, t)
}

func (self *Blockchain) GetAddresses() *common.Addresses {
	exs := map[common.ExchangeID]common.TokenAddresses{}
	for _, ex := range common.SupportedExchanges {
		exs[ex.ID()] = ex.TokenAddresses()
	}
	tokens := map[string]common.TokenInfo{}
	for _, t := range self.tokens {
		tokens[t.ID] = common.TokenInfo{
			Address:  ethereum.HexToAddress(t.Address),
			Decimals: t.Decimal,
		}
	}
	return &common.Addresses{
		Tokens:               tokens,
		Exchanges:            exs,
		WrapperAddress:       self.wrapperAddr,
		PricingAddress:       self.pricingAddr,
		ReserveAddress:       self.rm,
		FeeBurnerAddress:     self.burnerAddr,
		NetworkAddress:       self.networkAddr,
		PricingOperator:      self.signer.GetAddress(),
		DepositOperator:      self.depositSigner.GetAddress(),
		IntermediateOperator: self.intermediateSigner.GetAddress(),
	}
}

func (self *Blockchain) LoadAndSetTokenIndices() error {
	tokens := []ethereum.Address{}
	self.tokenIndices = map[string]tbindex{}

	for _, tok := range self.tokens {
		if tok.ID != "ETH" {
			tokens = append(tokens, ethereum.HexToAddress(tok.Address))
		} else {
			// this is not really needed. Just a safe guard
			self.tokenIndices[ethereum.HexToAddress(tok.Address).Hex()] = tbindex{1000000, 1000000}
		}
	}
	bulkIndices, indicesInBulk, err := self.wrapper.GetTokenIndicies(
		nil, nil,
		self.pricingAddr,
		tokens,
	)
	if err != nil {
		return err
	}
	for i, tok := range tokens {
		self.tokenIndices[tok.Hex()] = tbindex{
			bulkIndices[i].Uint64(),
			indicesInBulk[i].Uint64(),
		}
	}
	log.Printf("Token indices: %+v", self.tokenIndices)
	return nil
}

func getNextNonce(n NonceCorpus) (*big.Int, error) {
	var nonce *big.Int
	var err error
	for i := 0; i < 3; i++ {
		nonce, err = n.GetNextNonce()
		if err == nil {
			return nonce, nil
		}
	}
	return nonce, err
}

func donothing() {}

func (self *Blockchain) getTransactOpts(nonce *big.Int, gasPrice *big.Int) (*bind.TransactOpts, context.CancelFunc, error) {
	shared := self.signer.GetTransactOpts()
	var err error
	if nonce == nil {
		nonce, err = getNextNonce(self.nonce)
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

func (self *Blockchain) getDepositTransactOpts(nonce, gasPrice *big.Int) (*bind.TransactOpts, context.CancelFunc, error) {
	shared := self.depositSigner.GetTransactOpts()
	var err error
	if nonce == nil {
		nonce, err = getNextNonce(self.nonceDeposit)
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

func toBlockNumArg(number *big.Int) string {
	if number == nil {
		return "latest"
	}
	return hexutil.EncodeBig(number)
}

func toFilterArg(q ether.FilterQuery) interface{} {
	arg := map[string]interface{}{
		"fromBlock": toBlockNumArg(q.FromBlock),
		"toBlock":   toBlockNumArg(q.ToBlock),
		"address":   q.Addresses,
		"topics":    q.Topics,
	}
	if q.FromBlock == nil {
		arg["fromBlock"] = "0x0"
	}
	return arg
}

func (self *Blockchain) signAndBroadcast(tx *types.Transaction, singer Signer) (*types.Transaction, error) {
	if tx == nil {
		panic(errors.New("Nil tx is forbidden here"))
	} else {
		signedTx, err := singer.Sign(tx)
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

func (self *Blockchain) SetRateMinedNonce() (uint64, error) {
	nonce, err := self.nonce.MinedNonce()
	if err != nil {
		return 0, err
	} else {
		return nonce.Uint64(), err
	}
}

//====================== Write calls ===============================

// TODO: Need better test coverage
// we got a bug when compact is not set to old compact
// or when one of buy/sell got overflowed, it discards
// the other's compact
func (self *Blockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {

	opts, cancel, err := self.getTransactOpts(nonce, gasPrice)

	defer cancel()
	block.Add(block, big.NewInt(1))
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		baseBuys, baseSells, _, _, _, err := self.wrapper.GetTokenRates(
			nil, nil, self.pricingAddr, tokens,
		)
		if err != nil {
			return nil, err
		}
		baseTokens := []ethereum.Address{}
		newBSells := []*big.Int{}
		newBBuys := []*big.Int{}
		newCSells := map[ethereum.Address]byte{}
		newCBuys := map[ethereum.Address]byte{}
		for i, token := range tokens {
			compactSell, overflow1 := BigIntToCompactRate(sells[i], baseSells[i])
			compactBuy, overflow2 := BigIntToCompactRate(buys[i], baseBuys[i])
			if overflow1 || overflow2 {
				baseTokens = append(baseTokens, token)
				newBSells = append(newBSells, sells[i])
				newBBuys = append(newBBuys, buys[i])
				newCSells[token] = 0
				newCBuys[token] = 0
			} else {
				newCSells[token] = compactSell.Compact
				newCBuys[token] = compactBuy.Compact
			}
		}
		bbuys, bsells, indices := BuildCompactBulk(
			newCBuys,
			newCSells,
			self.tokenIndices,
		)
		var tx *types.Transaction
		if len(baseTokens) > 0 {
			// set base tx
			tx, err = self.pricing.SetBaseRate(
				opts, baseTokens, newBBuys, newBSells,
				bbuys, bsells, block, indices)
			if tx != nil {
				log.Printf(
					"broadcasting setbase tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new base buy(%s) || new base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
					tx.Hash().Hex(),
					buys, sells,
					baseBuys, baseSells,
					newBBuys, newBSells,
					newCBuys, newCSells,
					bbuys, bsells, indices,
				)
			}
		} else {
			// update compact tx
			tx, err = self.pricing.SetCompactData(
				opts, bbuys, bsells, block, indices)
			if tx != nil {
				log.Printf(
					"broadcasting setcompact tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
					tx.Hash().Hex(),
					buys, sells,
					baseBuys, baseSells,
					newCBuys, newCSells,
					bbuys, bsells, indices,
				)
			}
			// log.Printf("Setting compact rates: tx(%s), err(%v) with basesells(%+v), buys(%+v), sells(%+v), block(%s), indices(%+v)",
			// 	tx.Hash().Hex(), err, baseTokens, buys, sells, block.Text(10), indices,
			// )
		}
		if err != nil {
			return nil, err
		} else {
			return self.signAndBroadcast(tx, self.signer)
		}
	}
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
	return data, nil
}

func (self *Blockchain) SendTokenFromAccountToExchange(amount *big.Int, exchangeAddress ethereum.Address) (*types.Transaction, error) {

	opts, cancel, err := self.getIntermediateTransactOpts(nil, nil)
	ctx := opts.Context
	defer cancel()
	//build msg and get gas limit
	data, err := packData("transfer", exchangeAddress, amount)
	if err != nil {
		return nil, err
	}
	log.Printf("Intermediator: Data is %v", data)
	msg := ether.CallMsg{From: opts.From, To: &exchangeAddress, Value: big.NewInt(0), Data: data}
	gasLimit, err := self.client.EstimateGas(ensureContext(opts.Context), msg)
	if err != nil {
		log.Printf("Intermediator: Can not estimate gas %v", err)
		return nil, err
	}
	log.Println("Intermediator: gas limit estimated is : %d", gasLimit)
	if gasLimit.Cmp(big.NewInt(21000)) < 0 {
		gasLimit = big.NewInt(21000)
	}
	//build tx, sign and send
	tx := types.NewTransaction(opts.Nonce.Uint64(), exchangeAddress, amount, gasLimit, opts.GasPrice, nil)
	signTX, err := self.intermediateSigner.Sign(tx)
	if err != nil {
		log.Println("Intermediator: Can not sign the transaction")
		return nil, err
	}
	err = self.client.SendTransaction(ctx, signTX)
	if err != nil {
		log.Println("ERROR: Can't send the transaction")
		return nil, err
	}
	return tx, nil
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
	err = self.client.SendTransaction(ctx, signTX)
	if err != nil {
		log.Println("ERROR: Can't send the transaction")
		return nil, err
	}
	return tx, nil
}

func (self *Blockchain) Send(
	token common.Token,
	amount *big.Int,
	dest ethereum.Address) (*types.Transaction, error) {

	opts, cancel, err := self.getDepositTransactOpts(nil, nil)
	defer cancel()
	if err != nil {
		return nil, err
	} else {
		tx, err := self.reserve.Withdraw(
			opts,
			ethereum.HexToAddress(token.Address),
			amount, dest)
		if err != nil {
			return nil, err
		} else {
			return self.signAndBroadcast(tx, self.depositSigner)
		}
	}
}

func (self *Blockchain) SetImbalanceStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, cancel, err := self.getTransactOpts(nil, nil)
	defer cancel()
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		tx, err := self.pricing.SetImbalanceStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
		if err != nil {
			return nil, err
		}
		return self.signAndBroadcast(tx, self.signer)
	}
}

func (self *Blockchain) SetQtyStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, cancel, err := self.getTransactOpts(nil, nil)
	defer cancel()
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		tx, err := self.pricing.SetQtyStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
		if err != nil {
			return nil, err
		}
		return self.signAndBroadcast(tx, self.signer)
	}
}

//====================== Readonly calls ============================
func (self *Blockchain) CurrentBlock() (uint64, error) {
	var blockno string
	err := self.rpcClient.Call(&blockno, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseUint(blockno, 0, 64)
	return result, err
}

func (self *Blockchain) TransactionByHash(ctx context.Context, hash ethereum.Hash) (tx *rpcTransaction, isPending bool, err error) {
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
	return json, json.BlockNumber == nil, nil
}

func (self *Blockchain) TxStatus(hash ethereum.Hash) (string, uint64, error) {
	option := context.Background()
	tx, pending, err := self.TransactionByHash(option, hash)
	if err == nil {
		// tx exist
		if pending {
			return "", 0, nil
		} else {
			receipt, err := self.client.TransactionReceipt(option, hash)
			if err != nil {
				log.Println("Get receipt err: ", err.Error())
				log.Printf("Receipt: %+v", receipt)
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

func (self *Blockchain) FetchBalanceData(reserve ethereum.Address, atBlock *big.Int, timepoint uint64) (map[string]common.BalanceEntry, error) {
	result := map[string]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	for _, tok := range self.tokens {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	timestamp := common.GetTimestamp()
	balances, err := self.wrapper.GetBalances(nil, atBlock, reserve, tokens)
	returnTime := common.GetTimestamp()
	log.Printf("Fetcher ------> balances: %v, err: %s", balances, err)
	if err != nil {
		for tokenID, _ := range common.SupportedTokens {
			result[tokenID] = common.BalanceEntry{
				Valid:      false,
				Error:      err.Error(),
				Timestamp:  timestamp,
				ReturnTime: returnTime,
			}
		}
	} else {
		for i, tok := range self.tokens {
			if balances[i].Cmp(Big0) == 0 || balances[i].Cmp(BigMax) > 0 {
				log.Printf("Fetcher ------> balances of token %s is invalid", tok.ID)
				result[tok.ID] = common.BalanceEntry{
					Valid:      false,
					Error:      "Got strange balances from node. It equals to 0 or is bigger than 10^33",
					Timestamp:  timestamp,
					ReturnTime: returnTime,
					Balance:    common.RawBalance(*balances[i]),
				}
			} else {
				result[tok.ID] = common.BalanceEntry{
					Valid:      true,
					Timestamp:  timestamp,
					ReturnTime: returnTime,
					Balance:    common.RawBalance(*balances[i]),
				}
			}
		}
	}
	return result, nil
}

func (self *Blockchain) FetchRates(timepoint uint64, block uint64) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	tokenAddrs := []ethereum.Address{}
	validTokens := []common.Token{}
	for _, s := range self.tokens {
		if s.ID != "ETH" {
			tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(s.Address))
			validTokens = append(validTokens, s)
		}
	}
	timestamp := common.GetTimestamp()
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := self.wrapper.GetTokenRates(
		nil, big.NewInt(int64(block)), self.pricingAddr, tokenAddrs,
	)
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.BlockNumber = block
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		return result, err
	} else {
		result.Valid = true
		result.Data = map[string]common.RateEntry{}
		for i, token := range validTokens {
			result.Data[token.ID] = common.RateEntry{
				baseBuys[i],
				int8(compactBuys[i]),
				baseSells[i],
				int8(compactSells[i]),
				blocks[i].Uint64(),
			}
		}
		return result, nil
	}
}

func (self *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int, atBlock *big.Int) (*big.Int, error) {
	if priceType == "buy" {
		return self.pricing.GetRate(nil, atBlock, token, block, true, qty)
	} else {
		return self.pricing.GetRate(nil, atBlock, token, block, false, qty)
	}
}

func (self *Blockchain) GetRawLogs(fromBlock uint64, toBlock uint64, timepoint uint64) ([]types.Log, error) {
	result := []types.Log{}
	var to *big.Int
	if toBlock != 0 {
		to = big.NewInt(int64(toBlock))
	}
	// we have to track events from network and fee burner contracts
	// including their old contracts
	addresses := []ethereum.Address{}
	addresses = append(addresses, self.networkAddr, self.burnerAddr, self.whitelistAddr)
	addresses = append(addresses, self.oldNetworks...)
	addresses = append(addresses, self.oldBurners...)
	param := ether.FilterQuery{
		big.NewInt(int64(fromBlock)),
		to,
		addresses,
		[][]ethereum.Hash{
			[]ethereum.Hash{
				ethereum.HexToHash(TradeEvent),
				ethereum.HexToHash(BurnFeeEvent),
				ethereum.HexToHash(FeeToWalletEvent),
				ethereum.HexToHash(UserCatEvent),
			},
		},
	}
	err := self.rpcClient.Call(&result, "eth_getLogs", toFilterArg(param))
	return result, err
}

// return timestamp increasing array of trade log
func (self *Blockchain) GetLogs(fromBlock uint64, toBlock uint64, timepoint uint64, ethRate float64) ([]common.KNLog, error) {
	result := []common.KNLog{}
	// get all logs from fromBlock to best block
	logs, err := self.GetRawLogs(fromBlock, toBlock, timepoint)
	if err != nil {
		return result, err
	}
	var prevLog *types.Log
	var tradeLog *common.TradeLog
	for i, l := range logs {
		if l.Removed {
			log.Printf("Log is ignored because it is removed due to chain reorg")
		} else {
			if prevLog == nil || l.TxHash != prevLog.TxHash {
				if tradeLog != nil {
					result = append(result, *tradeLog)
				}
				if len(l.Topics) > 0 && l.Topics[0].Hex() != UserCatEvent {
					// start new TradeLog
					tradeLog = &common.TradeLog{}
					tradeLog.BlockNumber = l.BlockNumber
					tradeLog.TransactionHash = l.TxHash
					tradeLog.TransactionIndex = l.TxIndex
					tradeLog.Timestamp, err = self.InterpretTimestamp(
						tradeLog.BlockNumber,
						tradeLog.TransactionIndex,
					)
					if err != nil {
						return result, err
					}
				}
			}
			if len(l.Topics) == 0 {
				log.Printf("Getting empty zero topic list. This shouldn't happen and is Ethereum responsibility.")
			} else {
				topic := l.Topics[0]
				switch topic.Hex() {
				case UserCatEvent:
					addr, cat := LogDataToCatLog(l.Data)
					t, err := self.InterpretTimestamp(
						l.BlockNumber,
						l.TxIndex,
					)
					if err != nil {
						return result, err
					}
					result = append(result, common.SetCatLog{
						Timestamp:   t,
						BlockNumber: l.BlockNumber,
						Address:     addr,
						Category:    cat,
					})
				case FeeToWalletEvent:
					reserveAddr, walletAddr, walletFee := LogDataToFeeWalletParams(l.Data)
					tradeLog.ReserveAddress = reserveAddr
					tradeLog.WalletAddress = walletAddr
					tradeLog.WalletFee = walletFee.Big()
				case BurnFeeEvent:
					reserveAddr, burnFees := LogDataToBurnFeeParams(l.Data)
					tradeLog.ReserveAddress = reserveAddr
					tradeLog.BurnFee = burnFees.Big()
				case TradeEvent:
					srcAddr, destAddr, srcAmount, destAmount := LogDataToTradeParams(l.Data)
					tradeLog.SrcAddress = srcAddr
					tradeLog.DestAddress = destAddr
					tradeLog.SrcAmount = srcAmount.Big()
					tradeLog.DestAmount = destAmount.Big()
					tradeLog.UserAddress = ethereum.BytesToAddress(l.Topics[1].Bytes())

					if ethRate != 0 {
						// fiatAmount = amount * ethRate
						eth := common.SupportedTokens["ETH"]
						f := new(big.Float)
						if strings.ToLower(eth.Address) == strings.ToLower(srcAddr.String()) {
							f.SetInt(tradeLog.SrcAmount)
						} else {
							f.SetInt(tradeLog.DestAmount)
						}

						f = f.Mul(f, new(big.Float).SetFloat64(ethRate))
						f.Quo(f, new(big.Float).SetFloat64(math.Pow10(18)))
						tradeLog.FiatAmount, _ = f.Float64()
					}
				}
			}
			prevLog = &logs[i]
		}
	}
	if tradeLog != nil {
		result = append(result, *tradeLog)
	}
	return result, nil
}

// func (self *Blockchain) sendToken(token common.Token, amount *big.Int, address ethereum.Address) (ethereum.Hash, error) {
// 	erc20, err := NewErc20Contract(
// 		ethereum.HexToAddress(token.Address),
// 		self.ethclient,
// 	)
// 	fmt.Printf("address: %s\n", token.Address)
// 	if err != nil {
// 		return ethereum.Hash{}, err
// 	}
// 	tx, err := erc20.Transfer(
// 		self.signer.GetTransactOpts(),
// 		address, amount)
// 	if err != nil {
// 		return ethereum.Hash{}, err
// 	} else {
// 		return tx.Hash(), nil
// 	}
// }
//
// func (self *Blockchain) sendETH(
// 	amount *big.Int,
// 	address ethereum.Address) (ethereum.Hash, error) {
// 	// nonce, gasLimit, gasPrice gets from ethclient
//
// 	option := context.Background()
// 	rm := self.signer.GetAddress()
// 	nonce, err := self.ethclient.PendingNonceAt(
// 		option, rm)
// 	if err != nil {
// 		return ethereum.Hash{}, err
// 	}
// 	gasLimit := big.NewInt(1000000)
// 	gasPrice := big.NewInt(20000000000)
// 	rawTx := types.NewTransaction(
// 		nonce, address, amount, gasLimit, gasPrice, []byte{})
// 	signedTx, err := self.signer.Sign(rm, rawTx)
// 	if err != nil {
// 		return ethereum.Hash{}, err
// 	}
// 	if err = self.ethclient.SendTransaction(option, signedTx); err != nil {
// 		return ethereum.Hash{}, err
// 	}
// 	return signedTx.Hash(), nil
// }

func NewBlockchain(
	client *rpc.Client,
	etherCli *ethclient.Client,
	clients map[string]*ethclient.Client,
	wrapperAddr, pricingAddr, burnerAddr, networkAddr, reserveAddr, whitelistAddr ethereum.Address,
	signer Signer, depositSigner Signer, intermediateSigner Signer, nonceCorpus NonceCorpus,
	nonceDeposit NonceCorpus, nonceIntermediate NonceCorpus,
	chainType string) (*Blockchain, error) {
	log.Printf("wrapper address: %s", wrapperAddr.Hex())
	wrapper, err := NewKNWrapperContract(wrapperAddr, etherCli)
	if err != nil {
		return nil, err
	}
	log.Printf("reserve owner address: %s", signer.GetAddress().Hex())
	log.Printf("reserve address: %s", reserveAddr.Hex())
	reserve, err := NewKNReserveContract(reserveAddr, etherCli)
	if err != nil {
		return nil, err
	}
	log.Printf("pricing address: %s", pricingAddr.Hex())
	pricing, err := NewKNPricingContract(pricingAddr, etherCli)
	if err != nil {
		return nil, err
	}
	log.Printf("burner address: %s", burnerAddr.Hex())
	log.Printf("network address: %s", networkAddr.Hex())
	log.Printf("whitelist address: %s", whitelistAddr.Hex())
	log.Printf("intermediate address: %s", intermediateSigner.GetAddress().Hex())
	return &Blockchain{
		rpcClient:          client,
		client:             etherCli,
		wrapper:            wrapper,
		pricing:            pricing,
		reserve:            reserve,
		rm:                 reserveAddr,
		wrapperAddr:        wrapperAddr,
		pricingAddr:        pricingAddr,
		burnerAddr:         burnerAddr,
		networkAddr:        networkAddr,
		whitelistAddr:      whitelistAddr,
		oldNetworks:        []ethereum.Address{},
		oldBurners:         []ethereum.Address{},
		signer:             signer,
		depositSigner:      depositSigner,
		intermediateSigner: intermediateSigner,
		tokens:             []common.Token{},
		nonce:              nonceCorpus,
		nonceDeposit:       nonceDeposit,
		nonceIntermediate:  nonceIntermediate,
		broadcaster:        NewBroadcaster(clients),
	}, nil
}
