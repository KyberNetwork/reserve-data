package blockchain

import (
	"context"
	"log"
	"strconv"

	"github.com/KyberNetwork/reserve-data/common"
	ether "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"math/big"
)

type tbindex struct {
	BulkIndex   uint64
	IndexInBulk uint64
}

const (
	FeeToWalletEvent string = "0x366bc34352215bf0bd3b527cfd6718605e1f5938777e42bcd8ed92f578368f52"
	BurnFeeEvent     string = "0xf838f6ddc89706878e3c3e698e9b5cbfbf2c0e3d3dcd0bd2e00f1ccf313e0185"
	TradeEvent       string = "0x1849bd6a030a1bca28b83437fd3de96f3d27a5d172fa7e9c78e7b61468928a39"
)

// latest version
// ethereum.HexToHash("0x366bc34352215bf0bd3b527cfd6718605e1f5938777e42bcd8ed92f578368f52"),
// ropsten version

type Blockchain struct {
	rpcClient    *rpc.Client
	client       *ethclient.Client
	wrapper      *ContractWrapper
	pricing      *Pricing
	reserve      *ReserveContract
	rm           ethereum.Address
	wrapperAddr  ethereum.Address
	pricingAddr  ethereum.Address
	burnerAddr   ethereum.Address
	networkAddr  ethereum.Address
	signer       Signer
	tokens       []common.Token
	tokenIndices map[string]tbindex
	nonce        NonceCorpus
}

func (self *Blockchain) AddToken(t common.Token) {
	self.tokens = append(self.tokens, t)
}

func (self *Blockchain) GetAddresses() *common.Addresses {
	exs := map[common.ExchangeID]common.TokenAddresses{}
	for _, ex := range common.SupportedExchanges {
		exs[ex.ID()] = ex.TokenAddresses()
	}
	tokens := map[string]ethereum.Address{}
	for _, t := range self.tokens {
		tokens[t.ID] = ethereum.HexToAddress(t.Address)
	}
	return &common.Addresses{
		Tokens: tokens,
		Exchanges: exs,
		WrapperAddress: self.wrapperAddr,
		PricingAddress: self.pricingAddr,
		ReserveAddress: self.rm,
		FeeBurnerAddress: self.burnerAddr,
		NetworkAddress: self.networkAddr,
	}
}

func (self *Blockchain) LoadAndSetTokenIndices() error {
	tokens := []ethereum.Address{}
	for _, tok := range self.tokens {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	bulkIndices, indicesInBulk, err := self.wrapper.GetTokenIndicies(
		nil,
		self.pricingAddr,
		tokens,
	)
	if err != nil {
		return err
	}
	self.tokenIndices = map[string]tbindex{}

	for i, tok := range tokens {
		self.tokenIndices[tok.Hex()] = tbindex{
			bulkIndices[i].Uint64(),
			indicesInBulk[i].Uint64(),
		}
	}
	log.Printf("Token indices: %+v", self.tokenIndices)
	return nil
}

func (self *Blockchain) CurrentBlock() (uint64, error) {
	var blockno string
	err := self.rpcClient.Call(&blockno, "eth_blockNumber")
	if err != nil {
		return 0, err
	}
	result, err := strconv.ParseUint(blockno, 0, 64)
	return result, err
}

func (self *Blockchain) IsMined(tx ethereum.Hash) (bool, error) {
	option := context.Background()
	receipt, err := self.client.TransactionReceipt(option, tx)
	if receipt != nil {
		err = nil
	}
	return receipt != nil, err
}

func (self *Blockchain) getTransactOpts(transaction string) (*bind.TransactOpts, error) {
	shared := self.signer.GetTransactOpts(transaction)
	nonce, err := self.nonce.GetNextNonce(transaction)
	if err != nil {
		return nil, err
	} else {
		result := bind.TransactOpts{
			shared.From,
			nonce,
			shared.Signer,
			shared.Value,
			shared.GasPrice,
			shared.GasLimit,
			shared.Context,
		}
		return &result, nil
	}
}

func (self *Blockchain) FetchBalanceData(reserve ethereum.Address, timepoint uint64) (map[string]common.BalanceEntry, error) {
	result := map[string]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	for _, tok := range self.tokens {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	timestamp := common.GetTimestamp()
	balances, err := self.wrapper.GetBalances(nil, reserve, tokens)
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
			result[tok.ID] = common.BalanceEntry{
				Valid:      true,
				Timestamp:  timestamp,
				ReturnTime: returnTime,
				Balance:    common.RawBalance(*balances[i]),
			}
		}
	}
	return result, nil
}

func (self *Blockchain) FetchRates(timepoint uint64) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	tokenAddrs := []ethereum.Address{}
	for _, s := range self.tokens {
		tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(s.Address))
	}
	timestamp := common.GetTimestamp()
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := self.wrapper.GetTokenRates(
		nil, self.pricingAddr, tokenAddrs,
	)
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		return result, err
	} else {
		result.Valid = true
		result.Data = map[string]common.RateEntry{}
		for i, token := range self.tokens {
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

func (self *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int) (*big.Int, error) {
	if priceType == "buy" {
		return self.pricing.GetRate(nil, token, block, true, qty)
	} else {
		return self.pricing.GetRate(nil, token, block, false, qty)
	}
}

func (self *Blockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int) (ethereum.Hash, error) {

	opts, err := self.getTransactOpts("setrate")
	// fix to 50.1 gwei
	opts.GasPrice = big.NewInt(50100000000)
	block.Add(block, big.NewInt(1))
	if err != nil {
		log.Printf("Getting transaction opts failed!!!!!!!\n")
		return ethereum.Hash{}, err
	} else {
		baseBuys, baseSells, compactBuys, compactSells, _, err := self.wrapper.GetTokenRates(
			nil, self.pricingAddr, tokens,
		)
		if err != nil {
			return ethereum.Hash{}, err
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
				newBSells = append(newBSells, compactSell.Base)
				newBBuys = append(newBBuys, compactBuy.Base)
			} else {
				if compactSell.Compact != byte(compactSells[i]) ||
					compactBuy.Compact != byte(compactBuys[i]) {
					newCSells[token] = compactSell.Compact
					newCBuys[token] = compactBuy.Compact
				}
			}
		}
		buys, sells, indices := BuildCompactBulk(
			newCBuys,
			newCSells,
			self.tokenIndices,
		)
		var tx *types.Transaction
		if len(baseTokens) > 0 {
			// set base tx
			tx, err = self.pricing.SetBaseRate(
				opts, baseTokens, newBBuys, newBSells,
				buys, sells, block, indices)
			log.Printf("Setting base rates: tx(%s), err(%v) with baseTokens(%+v), basebuys(%+v), basesells(%+v), buys(%+v), sells(%+v), block(%s), indices(%+v)",
				tx.Hash().Hex(), err, baseTokens, newBBuys, newBSells, buys, sells, block.Text(10), indices,
			)
		} else {
			// update compact tx
			tx, err = self.pricing.SetCompactData(
				opts, buys, sells, block, indices)
			log.Printf("Setting compact rates: tx(%s), err(%v) with basesells(%+v), buys(%+v), sells(%+v), block(%s), indices(%+v)",
				tx.Hash().Hex(), err, baseTokens, buys, sells, block.Text(10), indices,
			)
		}
		if err != nil {
			log.Printf("Broadcasting transaction failed!!!!!!!\n")
			return ethereum.Hash{}, err
		} else {
			return tx.Hash(), err
		}
	}
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

func (self *Blockchain) GetRawLogs(fromBlock uint64, toBlock uint64, timepoint uint64) ([]types.Log, error) {
	result := []types.Log{}
	var to *big.Int
	if toBlock != 0 {
		to = big.NewInt(int64(toBlock))
	}
	param := ether.FilterQuery{
		big.NewInt(int64(fromBlock)),
		to,
		[]ethereum.Address{
			self.networkAddr,
			self.burnerAddr,
		},
		[][]ethereum.Hash{
			[]ethereum.Hash{
				ethereum.HexToHash(TradeEvent),
				ethereum.HexToHash(BurnFeeEvent),
				ethereum.HexToHash(FeeToWalletEvent),
			},
		},
	}
	err := self.rpcClient.Call(&result, "eth_getLogs", toFilterArg(param))
	return result, err
}

// return timestamp increasing array of trade log
func (self *Blockchain) GetLogs(fromBlock uint64, timepoint uint64) ([]common.TradeLog, error) {
	result := []common.TradeLog{}
	// get all logs from fromBlock to best block
	logs, err := self.GetRawLogs(fromBlock, 0, timepoint)
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
			if len(l.Topics) == 0 {
				log.Printf("Getting empty zero topic list. This shouldn't happen and is Ethereum responsibility.")
			} else {
				topic := l.Topics[0]
				switch topic.Hex() {
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

func (self *Blockchain) Send(
	token common.Token,
	amount *big.Int,
	dest ethereum.Address) (ethereum.Hash, error) {

	opts, err := self.getTransactOpts("deposit")
	if err != nil {
		return ethereum.Hash{}, err
	} else {
		tx, err := self.reserve.Withdraw(
			opts,
			ethereum.HexToAddress(token.Address),
			amount, dest)
		if err != nil {
			return ethereum.Hash{}, err
		} else {
			return tx.Hash(), err
		}
	}
}

func (self *Blockchain) SetImbalanceStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := self.getTransactOpts("setrate")
	if err != nil {
		log.Printf("Getting transaction opts failed!!!!!!!\n")
		return &types.Transaction{}, err
	} else {
		return self.pricing.SetImbalanceStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
	}
}

func (self *Blockchain) SetQtyStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := self.getTransactOpts("setrate")
	if err != nil {
		log.Printf("Getting transaction opts failed!!!!!!!\n")
		return &types.Transaction{}, err
	} else {
		return self.pricing.SetQtyStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
	}
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
	ethereum *ethclient.Client,
	wrapperAddr, pricingAddr, burnerAddr, networkAddr, reserveAddr ethereum.Address,
	signer Signer, nonceCorpus NonceCorpus) (*Blockchain, error) {
	log.Printf("wrapper address: %s", wrapperAddr.Hex())
	wrapper, err := NewContractWrapper(wrapperAddr, ethereum)
	if err != nil {
		return nil, err
	}
	log.Printf("reserve owner address: %s", signer.GetAddress("setrate").Hex())
	log.Printf("reserve address: %s", reserveAddr.Hex())
	reserve, err := NewReserveContract(reserveAddr, ethereum)
	if err != nil {
		return nil, err
	}
	log.Printf("pricing address: %s", pricingAddr.Hex())
	pricing, err := NewPricing(pricingAddr, ethereum)
	if err != nil {
		return nil, err
	}
	log.Printf("burner address: %s", burnerAddr.Hex())
	log.Printf("network address: %s", networkAddr.Hex())
	return &Blockchain{
		rpcClient:   client,
		client:      ethereum,
		wrapper:     wrapper,
		pricing:     pricing,
		reserve:     reserve,
		rm:          reserveAddr,
		wrapperAddr: wrapperAddr,
		pricingAddr: pricingAddr,
		burnerAddr:  burnerAddr,
		networkAddr: networkAddr,
		signer:      signer,
		tokens:      []common.Token{},
		nonce:       nonceCorpus,
	}, nil
}
