package blockchain

import (
	"fmt"
	"log"
	"math"
	"math/big"
	"strings"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	ether "github.com/ethereum/go-ethereum"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	PRICING_OP string = "pricingOP"
	DEPOSIT_OP string = "depositOP"
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
	*blockchain.BaseBlockchain
	wrapper       *blockchain.Contract
	pricing       *blockchain.Contract
	reserve       *blockchain.Contract
	rm            ethereum.Address
	wrapperAddr   ethereum.Address
	pricingAddr   ethereum.Address
	burnerAddr    ethereum.Address
	networkAddr   ethereum.Address
	whitelistAddr ethereum.Address
	oldNetworks   []ethereum.Address
	oldBurners    []ethereum.Address
	tokens        []common.Token
	tokenIndices  map[string]tbindex
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
	opAddrs := self.OperatorAddresses()
	return &common.Addresses{
		Tokens:           tokens,
		Exchanges:        exs,
		WrapperAddress:   self.wrapperAddr,
		PricingAddress:   self.pricingAddr,
		ReserveAddress:   self.rm,
		FeeBurnerAddress: self.burnerAddr,
		NetworkAddress:   self.networkAddr,
		PricingOperator:  opAddrs[PRICING_OP],
		DepositOperator:  opAddrs[DEPOSIT_OP],
	}
}

func (self *Blockchain) LoadAndSetTokenIndices() error {
	tokens := []ethereum.Address{}
	self.tokenIndices = map[string]tbindex{}

	log.Printf("tokens: %v", self.tokens)
	for _, tok := range self.tokens {
		if tok.ID != "ETH" {
			tokens = append(tokens, ethereum.HexToAddress(tok.Address))
		} else {
			// this is not really needed. Just a safe guard
			self.tokenIndices[ethereum.HexToAddress(tok.Address).Hex()] = tbindex{1000000, 1000000}
		}
	}
	opts := self.GetCallOpts(0)
	log.Printf("tokens: %v", tokens)
	bulkIndices, indicesInBulk, err := self.GeneratedGetTokenIndicies(
		opts,
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

func (self *Blockchain) RegisterPricingOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	log.Printf("reserve pricing address: %s", signer.GetAddress().Hex())
	self.RegisterOperator(PRICING_OP, blockchain.NewOperator(signer, nonceCorpus))
}

func (self *Blockchain) RegisterDepositOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	log.Printf("reserve depositor address: %s", signer.GetAddress().Hex())
	self.RegisterOperator(DEPOSIT_OP, blockchain.NewOperator(signer, nonceCorpus))
}

func readablePrint(data map[ethereum.Address]byte) string {
	result := ""
	for addr, b := range data {
		result = result + "|" + fmt.Sprintf("%s-%d", addr.Hex(), b)
	}
	return result
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

	block.Add(block, big.NewInt(1))
	copts := self.GetCallOpts(0)
	baseBuys, baseSells, _, _, _, err := self.GeneratedGetTokenRates(
		copts, self.pricingAddr, tokens,
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
	opts, err := self.GetTxOpts(PRICING_OP, nonce, gasPrice, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		var tx *types.Transaction
		if len(baseTokens) > 0 {
			// set base tx
			tx, err = self.GeneratedSetBaseRate(
				opts, baseTokens, newBBuys, newBSells,
				bbuys, bsells, block, indices)
			if tx != nil {
				log.Printf(
					"broadcasting setbase tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new base buy(%s) || new base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
					tx.Hash().Hex(),
					buys, sells,
					baseBuys, baseSells,
					newBBuys, newBSells,
					readablePrint(newCBuys), readablePrint(newCSells),
					bbuys, bsells, indices,
				)
			}
		} else {
			// update compact tx
			tx, err = self.GeneratedSetCompactData(
				opts, bbuys, bsells, block, indices)
			if tx != nil {
				log.Printf(
					"broadcasting setcompact tx %s, target buys(%s), target sells(%s), old base buy(%s) || old base sell(%s) || new compact buy(%s) || new compact sell(%s) || new buy bulk(%v) || new sell bulk(%v) || indices(%v)",
					tx.Hash().Hex(),
					buys, sells,
					baseBuys, baseSells,
					readablePrint(newCBuys), readablePrint(newCSells),
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
			return self.SignAndBroadcast(tx, PRICING_OP)
		}
	}
}

func (self *Blockchain) Send(
	token common.Token,
	amount *big.Int,
	dest ethereum.Address) (*types.Transaction, error) {

	opts, err := self.GetTxOpts(DEPOSIT_OP, nil, nil, nil)
	if err != nil {
		return nil, err
	} else {
		tx, err := self.GeneratedWithdraw(
			opts,
			ethereum.HexToAddress(token.Address),
			amount, dest)
		if err != nil {
			return nil, err
		} else {
			return self.SignAndBroadcast(tx, DEPOSIT_OP)
		}
	}
}

func (self *Blockchain) SetImbalanceStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := self.GetTxOpts(PRICING_OP, nil, nil, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		tx, err := self.GeneratedSetImbalanceStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
		if err != nil {
			return nil, err
		}
		return self.SignAndBroadcast(tx, PRICING_OP)
	}
}

func (self *Blockchain) SetQtyStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := self.GetTxOpts(PRICING_OP, nil, nil, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		tx, err := self.GeneratedSetQtyStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
		if err != nil {
			return nil, err
		}
		return self.SignAndBroadcast(tx, PRICING_OP)
	}
}

//====================== Readonly calls ============================
func (self *Blockchain) FetchBalanceData(reserve ethereum.Address, atBlock uint64) (map[string]common.BalanceEntry, error) {
	result := map[string]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	for _, tok := range self.tokens {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	timestamp := common.GetTimestamp()
	opts := self.GetCallOpts(atBlock)
	balances, err := self.GeneratedGetBalances(opts, reserve, tokens)
	returnTime := common.GetTimestamp()
	log.Printf("Fetcher ------> balances: %v, err: %s", balances, err)
	if err != nil {
		for _, token := range common.InternalTokens() {
			result[token.ID] = common.BalanceEntry{
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

func (self *Blockchain) FetchRates(atBlock uint64, currentBlock uint64) (common.AllRateEntry, error) {
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
	opts := self.GetCallOpts(atBlock)
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := self.GeneratedGetTokenRates(
		opts, self.pricingAddr, tokenAddrs,
	)
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.BlockNumber = currentBlock
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

func (self *Blockchain) GetReserveRates(
	atBlock, currentBlock uint64, reserveAddress ethereum.Address,
	tokens []common.Token) (common.ReserveRates, error) {
	result := common.ReserveTokenRateEntry{}
	rates := common.ReserveRates{}
	rates.Timestamp = common.GetTimepoint()

	ETH := common.ETHToken()
	srcAddresses := []ethereum.Address{}
	destAddresses := []ethereum.Address{}
	for _, token := range tokens {
		srcAddresses = append(srcAddresses, ethereum.HexToAddress(token.Address), ethereum.HexToAddress(ETH.Address))
		destAddresses = append(destAddresses, ethereum.HexToAddress(ETH.Address), ethereum.HexToAddress(token.Address))
	}

	opts := self.GetCallOpts(atBlock)
	reserveRate, sanityRate, err := self.GeneratedGetReserveRates(opts, reserveAddress, srcAddresses, destAddresses)
	if err != nil {
		return rates, err
	}

	rates.BlockNumber = atBlock
	rates.ToBlockNumber = currentBlock
	rates.ReturnTime = common.GetTimepoint()
	for index, token := range tokens {
		rateEntry := common.ReserveRateEntry{}
		rateEntry.BuyReserveRate = common.BigToFloat(reserveRate[index*2+1], 18)
		rateEntry.BuySanityRate = common.BigToFloat(sanityRate[index*2+1], 18)
		rateEntry.SellReserveRate = common.BigToFloat(reserveRate[index*2], 18)
		rateEntry.SellSanityRate = common.BigToFloat(sanityRate[index*2], 18)
		result[fmt.Sprintf("ETH-%s", token.ID)] = rateEntry
	}
	rates.Data = result

	return rates, err
}

func (self *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int, atBlock uint64) (*big.Int, error) {
	opts := self.GetCallOpts(atBlock)
	if priceType == "buy" {
		return self.GeneratedGetRate(opts, token, block, true, qty)
	} else {
		return self.GeneratedGetRate(opts, token, block, false, qty)
	}
}

func (self *Blockchain) GetRawLogs(fromBlock uint64, toBlock uint64) ([]types.Log, error) {
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
	log.Printf("LogFetcher - fetching logs data from block %d, to block %d", fromBlock, to.Uint64())
	return self.BaseBlockchain.GetLogs(param)
}

// return timestamp increasing array of trade log
func (self *Blockchain) GetLogs(fromBlock uint64, toBlock uint64) ([]common.KNLog, error) {
	result := []common.KNLog{}
	noCatLog := 0
	noTradeLog := 0
	// get all logs from fromBlock to best block
	logs, err := self.GetRawLogs(fromBlock, toBlock)
	if err != nil {
		return result, err
	}
	var prevLog *types.Log
	var tradeLog *common.TradeLog
	for i, l := range logs {
		if l.Removed {
			log.Printf("LogFetcher - Log is ignored because it is removed due to chain reorg")
		} else {
			if prevLog == nil || (l.TxHash != prevLog.TxHash && l.Topics[0].Hex() != UserCatEvent) {
				if tradeLog != nil {
					result = append(result, *tradeLog)
					noTradeLog += 1
					// log.Printf(
					// 	"LogFetcher - Fetched logs: TxHash(%s), TxIndex(%d), blockno(%d)",
					// 	tradeLog.TransactionHash.Hex(),
					// 	tradeLog.TransactionIndex,
					// 	tradeLog.BlockNumber,
					// )
				}
				if len(l.Topics) > 0 && l.Topics[0].Hex() != UserCatEvent {
					// start new TradeLog
					tradeLog = &common.TradeLog{}
					tradeLog.BlockNumber = l.BlockNumber
					tradeLog.TransactionHash = l.TxHash
					tradeLog.Index = l.Index
					tradeLog.Timestamp, err = self.InterpretTimestamp(
						tradeLog.BlockNumber,
						tradeLog.Index,
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
						l.Index,
					)
					if err != nil {
						return result, err
					}
					// log.Printf(
					// 	"LogFetcher - raw log entry: removed(%s), txhash(%s), timestamp(%d)",
					// 	l.Removed, l.TxHash.Hex(), t,
					// )
					result = append(result, common.SetCatLog{
						Timestamp:       t,
						BlockNumber:     l.BlockNumber,
						TransactionHash: l.TxHash,
						Index:           l.Index,
						Address:         addr,
						Category:        cat,
					})
					noCatLog += 1
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

					if ethRate := self.GetEthRate(tradeLog.Timestamp / 1000000); ethRate != 0 {
						// fiatAmount = amount * ethRate
						eth := common.ETHToken()
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
			if len(l.Topics) > 0 && l.Topics[0].Hex() != UserCatEvent {
				prevLog = &logs[i]
			}
		}
	}
	if tradeLog != nil && (len(result) == 0 || tradeLog.TransactionHash != result[len(result)-1].TxHash()) {
		result = append(result, *tradeLog)
		noTradeLog += 1
	}
	log.Printf("LogFetcher - Fetched %d trade logs, %d cat logs", noTradeLog, noCatLog)
	return result, nil
}

func (self *Blockchain) SetRateMinedNonce() (uint64, error) {
	return self.GetMinedNonce(PRICING_OP)
}

func NewBlockchain(
	base *blockchain.BaseBlockchain,
	wrapperAddr, pricingAddr, burnerAddr,
	networkAddr, reserveAddr, whitelistAddr ethereum.Address) (*Blockchain, error) {
	log.Printf("wrapper address: %s", wrapperAddr.Hex())
	wrapper := blockchain.NewContract(
		wrapperAddr,
		"/go/src/github.com/KyberNetwork/reserve-data/blockchain/wrapper.abi",
	)
	log.Printf("reserve address: %s", reserveAddr.Hex())
	reserve := blockchain.NewContract(
		reserveAddr,
		"/go/src/github.com/KyberNetwork/reserve-data/blockchain/reserve.abi",
	)
	log.Printf("pricing address: %s", pricingAddr.Hex())
	pricing := blockchain.NewContract(
		pricingAddr,
		"/go/src/github.com/KyberNetwork/reserve-data/blockchain/pricing.abi",
	)

	log.Printf("burner address: %s", burnerAddr.Hex())
	log.Printf("network address: %s", networkAddr.Hex())
	log.Printf("whitelist address: %s", whitelistAddr.Hex())

	return &Blockchain{
		BaseBlockchain: base,
		// blockchain.NewBaseBlockchain(
		// 	client, etherCli, operators, blockchain.NewBroadcaster(clients),
		// 	ethUSDRate, chainType,
		// ),
		wrapper:       wrapper,
		pricing:       pricing,
		reserve:       reserve,
		rm:            reserveAddr,
		wrapperAddr:   wrapperAddr,
		pricingAddr:   pricingAddr,
		burnerAddr:    burnerAddr,
		networkAddr:   networkAddr,
		whitelistAddr: whitelistAddr,
		oldNetworks:   []ethereum.Address{},
		oldBurners:    []ethereum.Address{},
		tokens:        []common.Token{},
	}, nil
}
