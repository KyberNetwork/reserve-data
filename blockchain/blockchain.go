package blockchain

import (
	"fmt"
	"log"
	"math/big"
	"path/filepath"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/settings"
	"github.com/ethereum/go-ethereum/accounts/abi"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	pricingOP = "pricingOP"
	depositOP = "depositOP"

	// feeToWalletEvent is the topic of event AssignFeeToWallet(address reserve, address wallet, uint walletFee).
	feeToWalletEvent = "0x366bc34352215bf0bd3b527cfd6718605e1f5938777e42bcd8ed92f578368f52"
	// burnFeeEvent is the topic of event AssignBurnFees(address reserve, uint burnFee).
	burnFeeEvent = "0xf838f6ddc89706878e3c3e698e9b5cbfbf2c0e3d3dcd0bd2e00f1ccf313e0185"
	// tradeEvent is the topic of event
	// ExecuteTrade(address indexed sender, ERC20 src, ERC20 dest, uint actualSrcAmount, uint actualDestAmount).
	tradeEvent = "0x1849bd6a030a1bca28b83437fd3de96f3d27a5d172fa7e9c78e7b61468928a39"
	// etherReceivalEvent is the topic of event EtherReceival(address indexed sender, uint amount).
	etherReceivalEvent = "0x75f33ed68675112c77094e7c5b073890598be1d23e27cd7f6907b4a7d98ac619"
	// userCatEvent is the topic of event UserCategorySet(address user, uint category).
	userCatEvent = "0x0aeb0f7989a09b8cccf58cea1aefa196ccf738cb14781d6910448dd5649d0e6e"
)

var (
	Big0   = big.NewInt(0)
	BigMax = big.NewInt(10).Exp(big.NewInt(10), big.NewInt(33), nil)
)

// tbindex is where the token data stored in blockchain.
// In blockchain, data of a token (sell/buy rates) is stored in an array of 32 bytes values called (tokenRatesCompactData).
// Each data is stored in a byte.
// https://github.com/KyberNetwork/smart-contracts/blob/fed8e09dc6e4365e1597474d9b3f53634eb405d2/contracts/ConversionRates.sol#L48
type tbindex struct {
	// BulkIndex is the index of bytes32 value that store data of multiple tokens.
	BulkIndex uint64
	// IndexInBulk is the index in the above BulkIndex value where the sell/buy rates are stored following structure:
	// sell: IndexInBulk + 4
	// buy: IndexInBulk + 8
	IndexInBulk uint64
}

// newTBIndex creates new tbindex instance with given parameters.
func newTBIndex(bulkIndex, indexInBulk uint64) tbindex {
	return tbindex{BulkIndex: bulkIndex, IndexInBulk: indexInBulk}
}

type Blockchain struct {
	*blockchain.BaseBlockchain
	wrapper      *blockchain.Contract
	pricing      *blockchain.Contract
	reserve      *blockchain.Contract
	tokenIndices map[string]tbindex
	mu           sync.RWMutex

	localSetRateNonce     uint64
	setRateNonceTimestamp uint64
	setting               Setting
}

func (bc *Blockchain) StandardGasPrice() float64 {
	// we use node's recommended gas price because gas station is not returning
	// correct gas price now
	price, err := bc.RecommendedGasPriceFromNode()
	if err != nil {
		return 0
	}
	return common.BigToFloat(price, 9)
}

func (bc *Blockchain) AddOldNetwork(addr ethereum.Address) error {
	return bc.setting.AddAddressToSet(settings.OldNetWorks, addr)
}

func (bc *Blockchain) AddOldBurners(addr ethereum.Address) error {
	return bc.setting.AddAddressToSet(settings.OldBurners, addr)
}

func (bc *Blockchain) CheckTokenIndices(tokenAddr ethereum.Address) error {
	opts := bc.GetCallOpts(0)
	pricingAddr, err := bc.setting.GetAddress(settings.Pricing)
	if err != nil {
		return err
	}
	tokenAddrs := []ethereum.Address{}
	tokenAddrs = append(tokenAddrs, tokenAddr)
	_, _, err = bc.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	if err != nil {
		return err
	}
	return nil
}

func (bc *Blockchain) LoadAndSetTokenIndices(tokenAddrs []ethereum.Address) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.tokenIndices = map[string]tbindex{}
	// this is not really needed. Just a safe guard. Use a very big indices so it would not exist.
	bc.tokenIndices[ethereum.HexToAddress(bc.setting.ETHToken().Address).Hex()] = tbindex{1000000, 1000000}
	opts := bc.GetCallOpts(0)
	pricingAddr, err := bc.setting.GetAddress(settings.Pricing)
	if err != nil {
		return err
	}
	bulkIndices, indicesInBulk, err := bc.GeneratedGetTokenIndicies(
		opts,
		pricingAddr,
		tokenAddrs,
	)
	if err != nil {
		return err
	}
	for i, tok := range tokenAddrs {
		bc.tokenIndices[tok.Hex()] = newTBIndex(
			bulkIndices[i].Uint64(),
			indicesInBulk[i].Uint64(),
		)
	}
	log.Printf("Token indices: %+v", bc.tokenIndices)
	return nil
}

func (bc *Blockchain) RegisterPricingOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	log.Printf("reserve pricing address: %s", signer.GetAddress().Hex())
	bc.MustRegisterOperator(pricingOP, blockchain.NewOperator(signer, nonceCorpus))
}

func (bc *Blockchain) RegisterDepositOperator(signer blockchain.Signer, nonceCorpus blockchain.NonceCorpus) {
	log.Printf("reserve depositor address: %s", signer.GetAddress().Hex())
	bc.MustRegisterOperator(depositOP, blockchain.NewOperator(signer, nonceCorpus))
}

func readablePrint(data map[ethereum.Address]byte) string {
	result := ""
	for addr, bc := range data {
		result = result + "|" + fmt.Sprintf("%s-%d", addr.Hex(), bc)
	}
	return result
}

//====================== Write calls ===============================

// TODO: Need better test coverage
// we got a bug when compact is not set to old compact
// or when one of buy/sell got overflowed, it discards
// the other's compact
func (bc *Blockchain) SetRates(
	tokens []ethereum.Address,
	buys []*big.Int,
	sells []*big.Int,
	block *big.Int,
	nonce *big.Int,
	gasPrice *big.Int) (*types.Transaction, error) {
	pricingAddr, err := bc.setting.GetAddress(settings.Pricing)
	if err != nil {
		return nil, err
	}
	block.Add(block, big.NewInt(1))
	copts := bc.GetCallOpts(0)
	baseBuys, baseSells, _, _, _, err := bc.GeneratedGetTokenRates(
		copts, pricingAddr, tokens,
	)
	if err != nil {
		return nil, err
	}

	// This is commented out because we dont want to make too much of change. Don't remove
	// this check, it can be useful in the future.
	//
	// Don't submit any txs if it is just trying to set all tokens to 0 when they are already 0
	// if common.AllZero(buys, sells, baseBuys, baseSells) {
	// 	return nil, errors.New("Trying to set all rates to 0 but they are already 0. Skip the tx.")
	// }

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
		bc.tokenIndices,
	)
	opts, err := bc.GetTxOpts(pricingOP, nonce, gasPrice, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	} else {
		var tx *types.Transaction
		if len(baseTokens) > 0 {
			// set base tx
			tx, err = bc.GeneratedSetBaseRate(
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
			tx, err = bc.GeneratedSetCompactData(
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
		}
		return bc.SignAndBroadcast(tx, pricingOP)
	}
}

func (bc *Blockchain) Send(
	token common.Token,
	amount *big.Int,
	dest ethereum.Address) (*types.Transaction, error) {

	opts, err := bc.GetTxOpts(depositOP, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	tx, err := bc.GeneratedWithdraw(
		opts,
		ethereum.HexToAddress(token.Address),
		amount, dest)
	if err != nil {
		return nil, err
	}
	return bc.SignAndBroadcast(tx, depositOP)
}

func (bc *Blockchain) SetImbalanceStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := bc.GetTxOpts(pricingOP, nil, nil, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	}
	tx, err := bc.GeneratedSetImbalanceStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
	if err != nil {
		return nil, err
	}
	return bc.SignAndBroadcast(tx, pricingOP)
}

func (bc *Blockchain) SetQtyStepFunction(token ethereum.Address, xBuy []*big.Int, yBuy []*big.Int, xSell []*big.Int, ySell []*big.Int) (*types.Transaction, error) {
	opts, err := bc.GetTxOpts(pricingOP, nil, nil, nil)
	if err != nil {
		log.Printf("Getting transaction opts failed, err: %s", err)
		return nil, err
	}
	tx, err := bc.GeneratedSetQtyStepFunction(opts, token, xBuy, yBuy, xSell, ySell)
	if err != nil {
		return nil, err
	}
	return bc.SignAndBroadcast(tx, pricingOP)
}

//====================== Readonly calls ============================
func (bc *Blockchain) FetchBalanceData(reserve ethereum.Address, atBlock uint64) (map[string]common.BalanceEntry, error) {
	result := map[string]common.BalanceEntry{}
	tokens := []ethereum.Address{}
	tokensSetting, err := bc.setting.GetInternalTokens()
	if err != nil {
		return result, err
	}
	for _, tok := range tokensSetting {
		tokens = append(tokens, ethereum.HexToAddress(tok.Address))
	}
	timestamp := common.GetTimestamp()
	opts := bc.GetCallOpts(atBlock)
	balances, err := bc.GeneratedGetBalances(opts, reserve, tokens)
	returnTime := common.GetTimestamp()
	log.Printf("Fetcher ------> balances: %v, err: %s", balances, common.ErrorToString(err))
	if err != nil {
		for _, token := range tokensSetting {
			result[token.ID] = common.BalanceEntry{
				Valid:      false,
				Error:      err.Error(),
				Timestamp:  timestamp,
				ReturnTime: returnTime,
			}
		}
	} else {
		for i, tok := range tokensSetting {
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

func (bc *Blockchain) FetchRates(atBlock uint64, currentBlock uint64) (common.AllRateEntry, error) {
	result := common.AllRateEntry{}
	tokenAddrs := []ethereum.Address{}
	validTokens := []common.Token{}
	tokenSettings, err := bc.setting.GetInternalTokens()
	if err != nil {
		return result, err
	}
	for _, s := range tokenSettings {
		if s.ID != "ETH" {
			tokenAddrs = append(tokenAddrs, ethereum.HexToAddress(s.Address))
			validTokens = append(validTokens, s)
		}
	}
	timestamp := common.GetTimestamp()
	opts := bc.GetCallOpts(atBlock)
	pricingAddr, err := bc.setting.GetAddress(settings.Pricing)
	if err != nil {
		return result, err
	}
	baseBuys, baseSells, compactBuys, compactSells, blocks, err := bc.GeneratedGetTokenRates(
		opts, pricingAddr, tokenAddrs,
	)
	if err != nil {
		return result, err
	}
	returnTime := common.GetTimestamp()
	result.Timestamp = timestamp
	result.ReturnTime = returnTime
	result.BlockNumber = currentBlock

	result.Data = map[string]common.RateEntry{}
	for i, token := range validTokens {
		result.Data[token.ID] = common.NewRateEntry(
			baseBuys[i],
			int8(compactBuys[i]),
			baseSells[i],
			int8(compactSells[i]),
			blocks[i].Uint64(),
		)
	}
	return result, nil
}

func (bc *Blockchain) GetReserveRates(
	atBlock, currentBlock uint64, reserveAddress ethereum.Address,
	tokens []common.Token) (common.ReserveRates, error) {
	result := common.ReserveTokenRateEntry{}
	rates := common.ReserveRates{}
	rates.Timestamp = common.GetTimepoint()

	ETH := bc.setting.ETHToken()
	srcAddresses := []ethereum.Address{}
	destAddresses := []ethereum.Address{}
	for _, token := range tokens {
		srcAddresses = append(srcAddresses, ethereum.HexToAddress(token.Address), ethereum.HexToAddress(ETH.Address))
		destAddresses = append(destAddresses, ethereum.HexToAddress(ETH.Address), ethereum.HexToAddress(token.Address))
	}

	opts := bc.GetCallOpts(atBlock)
	reserveRate, sanityRate, err := bc.GeneratedGetReserveRates(opts, reserveAddress, srcAddresses, destAddresses)
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

func (bc *Blockchain) GetPrice(token ethereum.Address, block *big.Int, priceType string, qty *big.Int, atBlock uint64) (*big.Int, error) {
	opts := bc.GetCallOpts(atBlock)
	if priceType == "buy" {
		return bc.GeneratedGetRate(opts, token, block, true, qty)
	}
	return bc.GeneratedGetRate(opts, token, block, false, qty)
}

func (bc *Blockchain) GetRawLogs(fromBlock uint64, toBlock uint64) ([]types.Log, error) {
	var (
		from      = big.NewInt(int64(fromBlock))
		to        = big.NewInt(int64(toBlock))
		addresses []ethereum.Address
	)
	// we have to track events from network and fee burner contracts
	// including their old contracts
	networkAddr, err := bc.setting.GetAddress(settings.Network)
	if err != nil {
		return nil, err
	}
	burnerAddr, err := bc.setting.GetAddress(settings.Burner)
	if err != nil {
		return nil, err
	}
	whitelistAddr, err := bc.setting.GetAddress(settings.Whitelist)
	if err != nil {
		return nil, err
	}

	addresses = append(addresses, networkAddr, burnerAddr, whitelistAddr)
	oldNetworks, err := bc.setting.GetAddresses(settings.OldNetWorks)
	if err != nil {
		log.Printf("WARNING: can't get old network addresses (%s)", err)
	}
	oldBurners, err := bc.setting.GetAddresses(settings.OldBurners)
	if err != nil {
		log.Printf("WARNING: can't get old burners addresses (%s)", err)
	}
	addresses = append(addresses, oldNetworks...)
	addresses = append(addresses, oldBurners...)
	param := common.NewFilterQuery(
		from,
		to,
		addresses,
		[][]ethereum.Hash{
			{
				ethereum.HexToHash(tradeEvent),
				ethereum.HexToHash(burnFeeEvent),
				ethereum.HexToHash(feeToWalletEvent),
				ethereum.HexToHash(userCatEvent),
				ethereum.HexToHash(etherReceivalEvent),
			},
		},
	)

	log.Printf("LogFetcher - fetching logs data from block %d, to block %d", fromBlock, to.Uint64())
	return bc.BaseBlockchain.GetLogs(param)
}

// GetLogs gets raw logs from blockchain and process it before returning.
func (bc *Blockchain) GetLogs(fromBlock uint64, toBlock uint64) ([]common.KNLog, error) {
	var (
		err      error
		result   []common.KNLog
		noCatLog = 0
		eth      = bc.setting.ETHToken()
	)

	// get all logs from fromBlock to best block
	logs, err := bc.GetRawLogs(fromBlock, toBlock)
	if err != nil {
		return result, err
	}

	for _, logItem := range logs {
		if logItem.Removed {
			log.Printf("LogFetcher - Log is ignored because it is removed due to chain reorg")
			continue
		}

		if len(logItem.Topics) == 0 {
			log.Printf("Getting empty zero topic list. This shouldn't happen and is Ethereum responsibility.")
			continue
		}

		ts, err := bc.InterpretTimestamp(
			logItem.BlockNumber,
			logItem.Index,
		)
		if err != nil {
			return result, err
		}

		topic := logItem.Topics[0]
		switch topic.Hex() {
		case userCatEvent:
			addr, cat := logDataToCatLog(logItem.Data)
			result = append(result, common.SetCatLog{
				Timestamp:       ts,
				BlockNumber:     logItem.BlockNumber,
				TransactionHash: logItem.TxHash,
				Index:           logItem.Index,
				Address:         addr,
				Category:        cat,
			})
			noCatLog++
		case feeToWalletEvent, burnFeeEvent, etherReceivalEvent, tradeEvent:
			if result, err = updateTradeLogs(result, logItem, ts); err != nil {
				return result, err
			}
		default:
			log.Printf("Unknown topic: %s", topic.Hex())
		}
	}

	for i, logItem := range result {
		tradeLog, ok := logItem.(common.TradeLog)
		if !ok {
			continue
		}

		ethRate := bc.GetEthRate(tradeLog.Timestamp / 1000000)
		if ethRate != 0 {
			result[i] = calculateFiatAmount(tradeLog, ethRate, eth)
		}
	}

	log.Printf("LogFetcher - Fetched %d trade logs, %d cat logs", len(result)-noCatLog, noCatLog)
	return result, nil
}

// SetRateMinedNonce returns nonce of the pricing operator in confirmed
// state (not pending state).
//
// Getting mined nonce is not simple because there might be lag between
// node leading us to get outdated mined nonce from an unsynced node.
// To overcome this situation, we will keep a local nonce and require
// the nonce from node to be equal or greater than it.
// If the nonce from node is smaller than the local one, we will use
// the local one. However, if the local one stay with the same value
// for more than 15mins, the local one is considered incorrect
// because the chain might be reorganized so we will invalidate it
// and assign it to the nonce from node.
func (bc *Blockchain) SetRateMinedNonce() (uint64, error) {
	nonceFromNode, err := bc.GetMinedNonce(pricingOP)
	if err != nil {
		return nonceFromNode, err
	}
	if nonceFromNode < bc.localSetRateNonce {
		if common.GetTimepoint()-bc.setRateNonceTimestamp > uint64(15*time.Minute) {
			bc.localSetRateNonce = nonceFromNode
			bc.setRateNonceTimestamp = common.GetTimepoint()
			return nonceFromNode, nil
		} else {
			return bc.localSetRateNonce, nil
		}
	} else {
		bc.localSetRateNonce = nonceFromNode
		bc.setRateNonceTimestamp = common.GetTimepoint()
		return nonceFromNode, nil
	}
}

func (bc *Blockchain) GetPricingMethod(inputData string) (*abi.Method, error) {
	abiPricing := &bc.pricing.ABI
	inputDataByte, err := hexutil.Decode(inputData)
	if err != nil {
		log.Printf("Cannot decode data: %v", err)
		return nil, err
	}
	method, err := abiPricing.MethodById(inputDataByte)
	if err != nil {
		return nil, err
	}
	return method, nil
}

func NewBlockchain(base *blockchain.BaseBlockchain, setting Setting) (*Blockchain, error) {
	wrapperAddr, err := setting.GetAddress(settings.Wrapper)
	if err != nil {
		return nil, err
	}
	log.Printf("wrapper address: %s", wrapperAddr.Hex())
	wrapper := blockchain.NewContract(
		wrapperAddr,
		filepath.Join(common.CurrentDir(), "wrapper.abi"),
	)
	reserveAddr, err := setting.GetAddress(settings.Reserve)
	if err != nil {
		return nil, err
	}
	log.Printf("reserve address: %s", reserveAddr.Hex())
	reserve := blockchain.NewContract(
		reserveAddr,
		filepath.Join(common.CurrentDir(), "reserve.abi"),
	)
	pricingAddr, err := setting.GetAddress(settings.Pricing)
	if err != nil {
		return nil, err
	}
	log.Printf("pricing address: %s", pricingAddr.Hex())
	pricing := blockchain.NewContract(
		pricingAddr,
		filepath.Join(common.CurrentDir(), "pricing.abi"),
	)
	burnerAddr, err := setting.GetAddress(settings.Burner)
	if err != nil {
		return nil, err
	}
	log.Printf("burner address: %s", burnerAddr.Hex())
	networkAddr, err := setting.GetAddress(settings.Network)
	if err != nil {
		return nil, err
	}
	log.Printf("network address: %s", networkAddr.Hex())
	whitelistAddr, err := setting.GetAddress(settings.Whitelist)
	if err != nil {
		return nil, err
	}
	log.Printf("whitelist address: %s", whitelistAddr.Hex())

	return &Blockchain{
		BaseBlockchain: base,
		// blockchain.NewBaseBlockchain(
		// 	client, etherCli, operators, blockchain.NewBroadcaster(clients),
		// 	ethUSDRate, chainType,
		// ),
		wrapper: wrapper,
		pricing: pricing,
		reserve: reserve,
		setting: setting,
	}, nil
}
