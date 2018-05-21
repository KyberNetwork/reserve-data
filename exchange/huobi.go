package exchange

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	huobiblockchain "github.com/KyberNetwork/reserve-data/exchange/huobi/blockchain"
	huobihttp "github.com/KyberNetwork/reserve-data/exchange/huobi/http"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	HUOBI_EPSILON float64 = 0.0000000001 // 10e-10
)

type Huobi struct {
	interf            HuobiInterface
	pairs             []common.TokenPair
	tokens            []common.Token
	addresses         *common.ExchangeAddresses
	exchangeInfo      *common.ExchangeInfo
	fees              common.ExchangeFees
	blockchain        HuobiBlockchain
	intermediatorAddr ethereum.Address
	storage           HuobiStorage
	minDeposit        common.ExchangesMinDeposit
}

func (self *Huobi) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Huobi) TokenAddresses() map[string]ethereum.Address {
	return self.addresses.GetData()
}

func (self *Huobi) Address(token common.Token) (ethereum.Address, bool) {

	_, supported := self.addresses.Get(token.ID)
	addr := self.intermediatorAddr
	return addr, supported
}

func (self *Huobi) UpdateAllDepositAddresses(address string, timepoint uint64) {
	data := self.addresses.GetData()
	for k, _ := range data {
		self.addresses.Update(k, ethereum.HexToAddress(address))
	}
}

func (self *Huobi) UpdateDepositAddress(token common.Token, address string) {
	liveAddress, _ := self.interf.GetDepositAddress(strings.ToLower(token.ID))
	if liveAddress.Address != "" {
		self.addresses.Update(token.ID, ethereum.HexToAddress(liveAddress.Address))
	} else {
		self.addresses.Update(token.ID, ethereum.HexToAddress(address))
	}
}

func (self *Huobi) UpdatePrecisionLimit(pair common.TokenPair, symbols HuobiExchangeInfo) {
	pairName := strings.ToLower(pair.Base.ID) + strings.ToLower(pair.Quote.ID)
	for _, symbol := range symbols.Data {
		if symbol.Base+symbol.Quote == pairName {
			exchangePrecisionLimit := common.ExchangePrecisionLimit{}
			exchangePrecisionLimit.Precision.Amount = symbol.AmountPrecision
			exchangePrecisionLimit.Precision.Price = symbol.PricePrecision
			exchangePrecisionLimit.MinNotional = 0.02
			self.exchangeInfo.Update(pair.PairID(), exchangePrecisionLimit)
			break
		}
	}
}

func (self *Huobi) UpdatePairsPrecision() {
	exchangeInfo, err := self.interf.GetExchangeInfo()
	if err != nil {
		log.Printf("Get exchange info failed: %s\n", err)
	} else {
		for _, pair := range self.pairs {
			self.UpdatePrecisionLimit(pair, exchangeInfo)
		}
	}
}

func (self *Huobi) GetInfo() (common.ExchangeInfo, error) {
	return *self.exchangeInfo, nil
}

func (self *Huobi) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	data, err := self.exchangeInfo.Get(pair)
	return data, err
}

func (self *Huobi) GetFee() common.ExchangeFees {
	return self.fees
}

func (self *Huobi) GetMinDeposit() common.ExchangesMinDeposit {
	return self.minDeposit
}

func (self *Huobi) ID() common.ExchangeID {
	return common.ExchangeID("huobi")
}

func (self *Huobi) TokenPairs() []common.TokenPair {
	return self.pairs
}

func (self *Huobi) Name() string {
	return "huobi"
}

func (self *Huobi) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.OrderStatus(symbol, id)
	if err != nil {
		return 0, 0, false, err
	} else {
		done, _ := strconv.ParseFloat(result.Data.ExecutedQty, 64)
		total, _ := strconv.ParseFloat(result.Data.OrigQty, 64)
		return done, total - done, total-done < HUOBI_EPSILON, nil
	}
}

func (self *Huobi) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	result, err := self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)

	if err != nil {
		return "", 0, 0, false, err
	} else {
		orderID, _ := strconv.ParseUint(result.OrderID, 10, 64)
		done, remaining, finished, err := self.QueryOrder(
			base.ID+quote.ID,
			orderID,
		)
		return result.OrderID, done, remaining, finished, err
	}
}

func (self *Huobi) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	withdrawID, err := self.interf.Withdraw(token, amount, address)
	if err != nil {
		return "", err
	}
	// this magical logic base on inspection on huobi website
	result := withdrawID + "01"
	return result, err
}

func (self *Huobi) CancelOrder(id, base, quote string) error {
	idNo, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return err
	}
	symbol := base + quote
	result, err := self.interf.CancelOrder(symbol, idNo)
	if err != nil {
		return err
	}
	if result.Status != "ok" {
		return fmt.Errorf("Couldn't cancel order id " + id)
	}
	return nil
}

func (self *Huobi) FetchOnePairData(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	resp_data, err := self.interf.GetDepthOnePair(pair)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if resp_data.Status != "ok" {
			result.Valid = false
		} else {
			for _, buy := range resp_data.Tick.Bids {
				quantity := buy[1]
				rate := buy[0]
				result.Bids = append(
					result.Bids,
					common.PriceEntry{
						quantity,
						rate,
					},
				)
			}
			for _, sell := range resp_data.Tick.Asks {
				quantity := sell[1]
				rate := sell[0]
				result.Asks = append(
					result.Asks,
					common.PriceEntry{
						quantity,
						rate,
					},
				)
			}
		}
	}
	data.Store(pair.PairID(), result)
}

func (self *Huobi) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.(common.ExchangePrice)
		return true
	})
	return result, nil
}

func (self *Huobi) OpenOrdersForOnePair(
	wg *sync.WaitGroup,
	pair common.TokenPair,
	data *sync.Map,
	timepoint uint64) {

	// defer wg.Done()

	// result, err := self.interf.OpenOrdersForOnePair(pair, timepoint)

	//TODO: complete open orders for one pair
}

func (self *Huobi) FetchOrderData(timepoint uint64) (common.OrderEntry, error) {
	result := common.OrderEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Data = []common.Order{}

	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.OpenOrdersForOnePair(&wait, pair, &data, timepoint)
	}
	wait.Wait()

	result.ReturnTime = common.GetTimestamp()

	data.Range(func(key, value interface{}) bool {
		orders := value.([]common.Order)
		result.Data = append(result.Data, orders...)
		return true
	})
	return result, nil
}

func (self *Huobi) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Error = ""
	resp_data, err := self.interf.GetInfo()
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
		result.Status = false
	} else {
		result.AvailableBalance = map[string]float64{}
		result.LockedBalance = map[string]float64{}
		result.DepositBalance = map[string]float64{}
		result.Status = true
		if resp_data.Status != "ok" {
			result.Valid = false
			result.Error = fmt.Sprintf("Cannot fetch ebalance")
			result.Status = false
		} else {
			balances := resp_data.Data.List
			for _, b := range balances {
				tokenID := strings.ToUpper(b.Currency)
				_, err := common.GetInternalToken(tokenID)
				if err == nil {
					balance, _ := strconv.ParseFloat(b.Balance, 64)
					if b.Type == "trade" {
						result.AvailableBalance[tokenID] = balance
					} else {
						result.LockedBalance[tokenID] = balance
					}
					result.DepositBalance[tokenID] = 0
				}
			}
		}
	}
	return result, nil
}

func (self *Huobi) FetchOnePairTradeHistory(
	wait *sync.WaitGroup,
	data *sync.Map,
	pair common.TokenPair) {

	defer wait.Done()
	result := []common.TradeHistory{}
	resp, err := self.interf.GetAccountTradeHistory(pair.Base, pair.Quote)
	if err != nil {
		log.Printf("Cannot fetch data for pair %s%s: %s", pair.Base.ID, pair.Quote.ID, err.Error())
	}
	pairString := pair.PairID()
	for _, trade := range resp.Data {
		price, _ := strconv.ParseFloat(trade.Price, 64)
		quantity, _ := strconv.ParseFloat(trade.Amount, 64)
		historyType := "sell"
		if trade.Type == "buy-limit" {
			historyType = "buy"
		}
		tradeHistory := common.TradeHistory{
			strconv.FormatUint(trade.ID, 10),
			price,
			quantity,
			historyType,
			trade.Timestamp,
		}
		result = append(result, tradeHistory)
	}
	data.Store(pairString, result)
}

func (self *Huobi) FetchTradeHistory() {
	t := time.NewTicker(10 * time.Second)
	go func() {
		for {
			result := map[common.TokenPairID][]common.TradeHistory{}
			data := sync.Map{}
			pairs := self.pairs
			wait := sync.WaitGroup{}
			for _, pair := range pairs {
				wait.Add(1)
				go self.FetchOnePairTradeHistory(&wait, &data, pair)
			}
			wait.Wait()
			data.Range(func(key, value interface{}) bool {
				result[key.(common.TokenPairID)] = value.([]common.TradeHistory)
				return true
			})
			self.storage.StoreTradeHistory(result)
			<-t.C
		}
	}()
}

func (self *Huobi) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return self.storage.GetTradeHistory(fromTime, toTime)
}

func getDepositInfo(id common.ActivityID) (string, float64, string) {
	idParts := strings.Split(id.EID, "|")
	txID := idParts[0]
	sentAmount, err := strconv.ParseFloat(idParts[2], 64)
	if err != nil {
		log.Println("The ID is malform, cannot get Amount from EID")
	}
	tokenID := idParts[1]
	return txID, sentAmount, tokenID
}

func (self *Huobi) Send2ndTransaction(amount float64, token common.Token, exchangeAddress ethereum.Address) (*types.Transaction, error) {
	IAmount := common.FloatToBigInt(amount, token.Decimal)
	// Check balance, removed from huobi's blockchain object.
	// currBalance := self.blockchain.CheckBalance(token)
	// log.Printf("current balance of token %s is %d", token.ID, currBalance)
	// //self.blockchain.
	// if currBalance.Cmp(IAmount) < 0 {
	// 	log.Printf("balance is not enough, wait till next check")
	// 	return nil, errors.New("balance is not enough")
	// }
	var tx *types.Transaction
	var err error
	if token.ID == "ETH" {
		tx, err = self.blockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress)
	} else {
		tx, err = self.blockchain.SendTokenFromAccountToExchange(IAmount, exchangeAddress, ethereum.HexToAddress(token.Address))
	}
	if err != nil {
		log.Printf("ERROR: Can not send transaction to exchange: %v", err)
		return nil, err
	}
	log.Printf("Transaction submitted. Tx is: %v", tx)
	return tx, nil

}

func (self *Huobi) PendingIntermediateTxs() (map[common.ActivityID]common.TXEntry, error) {
	result, err := self.storage.GetPendingIntermediateTXs()
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (self *Huobi) FindTx2InPending(id common.ActivityID) (common.TXEntry, bool) {
	pendings, err := self.storage.GetPendingIntermediateTXs()
	if err != nil {
		log.Printf("can't get pendings tx2 records: %v", err)
		return common.TXEntry{}, false
	}
	for actID, txentry := range pendings {
		if actID == id {
			return txentry, true
		}
	}
	return common.TXEntry{}, false
}

//FindTx2 : find Tx2 Record associates with activity ID, return
func (self *Huobi) FindTx2(id common.ActivityID) (Tx2 common.TXEntry, found bool) {
	found = true
	//first look it up in permanent bucket
	Tx2, err := self.storage.GetIntermedatorTx(id)
	if err != nil {
		//couldn't look for it in permanent bucket, look for it in pending bucket
		Tx2, found = self.FindTx2InPending(id)
	}
	return Tx2, found
}

func (self *Huobi) DepositStatus(id common.ActivityID, tx1Hash, currency string, sentAmount float64, timepoint uint64) (string, error) {

	var data common.TXEntry
	tx2Entry, found := self.FindTx2(id)
	if !found {
		//if not found, meaning there is no tx2 yet, check the 1st Tx to process.

		status, blockno, err := self.blockchain.TxStatus(ethereum.HexToHash(tx1Hash))
		if err != nil {
			log.Printf("Can not get TX status (%s)", err.Error())
			return "", nil
		}
		log.Printf("Status for Tx1 was %s at block %d ", status, blockno)
		if status == "mined" {
			//if it is mined, send 2nd tx.
			log.Printf("Found a new deposit status, which deposit %f %s. Procceed to send it to Huobi", sentAmount, currency)
			//check if the token is supported
			token, err := common.GetInternalToken(currency)
			if err != nil {
				return "", err
			}
			exchangeAddress, ok := self.addresses.Get(currency)
			if !ok {
				return "", fmt.Errorf("Wrong token address configuration")
			}
			tx2, err := self.Send2ndTransaction(sentAmount, token, exchangeAddress)
			if err != nil {
				log.Printf("Trying to send 2nd tx failed, error: %s. Will retry next time", err.Error())
				return "", nil
			}
			//store tx2 to pendingIntermediateTx
			data = common.TXEntry{tx2.Hash().Hex(), self.Name(), currency, "submitted", "", sentAmount, common.GetTimestamp()}
			err = self.storage.StorePendingIntermediateTx(id, data)
			if err != nil {
				log.Printf("Trying to store 2nd tx to pending tx storage failed, error: %s. It will be ignored and can make us to send to huobi again and the deposit will be marked as failed because the fund is not efficient", err.Error())
			}
			return "", nil
		} else {
			//No need to handle other blockchain status of TX1 here, since Fetcher will handle it from blockchain Status.
			return "", nil
		}
	} else {
		// if there is tx2Entry, check it blockchain status first:
		status, _, err := self.blockchain.TxStatus(ethereum.HexToHash(tx2Entry.Hash))
		if err != nil {
			return "", err
		}
		if status == "mined" {
			log.Println("2nd Transaction is mined. Processed to store it and check the Huobi Deposit history")
			data = common.TXEntry{tx2Entry.Hash, self.Name(), currency, "mined", "", sentAmount, common.GetTimestamp()}
			err = self.storage.StorePendingIntermediateTx(id, data)
			if err != nil {
				log.Printf("Trying to store intermediate tx to huobi storage, error: %s. Ignore it and try later", err.Error())
				return "", nil
			}
			deposits, err := self.interf.DepositHistory()
			if err != nil || deposits.Status != "ok" {
				log.Printf("Getting deposit history from huobi failed, error: %v, status: %s", err, deposits.Status)
				return "", nil
			}
			//check tx2 deposit status from Huobi
			for _, deposit := range deposits.Data {
				// log.Printf("deposit tx is %s, with token %s", deposit.TxHash, deposit.Currency)
				if deposit.TxHash == tx2Entry.Hash {
					if deposit.State == "safe" || deposit.State == "confirmed" {
						data = common.TXEntry{tx2Entry.Hash, self.Name(), currency, "mined", "done", sentAmount, common.GetTimestamp()}
						err = self.storage.StoreIntermediateTx(id, data)
						if err != nil {
							log.Printf("Trying to store intermediate tx to huobi storage, error: %s. Ignore it and try later", err.Error())
							return "", nil
						}
						err = self.storage.RemovePendingIntermediateTx(id)
						if err != nil {
							log.Printf("Trying to remove pending intermediate tx from huobi storage, error: %s. Ignore it and treat it like it is still pending", err.Error())
							return "", nil
						}
						return "done", nil
					} else {
						//TODO : handle other states following https://github.com/huobiapi/API_Docs_en/wiki/REST_Reference#deposit-states
						log.Printf("Tx %s is found but the status was not safe but %s", deposit.TxHash, deposit.State)
						return "", nil
					}
				}
			}
			log.Printf("Deposit doesn't exist. Huobi hasn't recognized the deposit yet or in theory, you have more than %d deposits at the same time.", len(common.InternalTokens())*2)
			return "", nil
		} else if status == "failed" {
			data = common.TXEntry{tx2Entry.Hash, self.Name(), currency, "failed", "failed", sentAmount, common.GetTimestamp()}

			return "failed", nil
		} else if status == "lost" {
			elapsed := common.GetTimepoint() - tx2Entry.Timestamp.ToUint64()
			if elapsed > uint64(15*time.Minute/time.Millisecond) {
				data = common.TXEntry{tx2Entry.Hash, self.Name(), currency, "lost", "lost", sentAmount, common.GetTimestamp()}
				err = self.storage.StoreIntermediateTx(id, data)
				if err != nil {
					log.Printf("Trying to store intermediate tx failed, error: %s. Ignore it and treat it like it is still pending", err.Error())
					return "", nil
				}
				err = self.storage.RemovePendingIntermediateTx(id)
				if err != nil {
					log.Printf("Trying to remove pending intermediate tx from huobi storage, error: %s. Ignore it and treat it like it is still pending", err.Error())
					return "", nil
				}
				log.Printf("The tx is not found for over 15mins, it is considered as lost and the deposit failed")
				return "failed", nil
			} else {
				return "", nil
			}
		}
	}
	log.Printf("should not be here")
	return "", nil
}

func (self *Huobi) WithdrawStatus(
	id, currency string, amount float64, timepoint uint64) (string, string, error) {
	withdrawID, _ := strconv.ParseUint(id, 10, 64)
	withdraws, err := self.interf.WithdrawHistory()
	if err != nil {
		return "", "", nil
	}
	log.Printf("Withdrawal id: %d", withdrawID)
	for _, withdraw := range withdraws.Data {
		if withdraw.TxID == withdrawID {
			if withdraw.State == "confirmed" {
				return "done", withdraw.TxHash, nil
			}
			return "", withdraw.TxHash, nil
		}
	}
	return "", "", fmt.Errorf("Withdrawal doesn't exist. This shouldn't happen unless tx returned from withdrawal from huobi and activity ID are not consistently designed")
}

func (self *Huobi) OrderStatus(id string, base, quote string) (string, error) {
	orderID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return "", err
	}
	symbol := base + quote
	order, err := self.interf.OrderStatus(symbol, orderID)
	if err != nil {
		return "", err
	}
	if order.Data.State == "pre-submitted" || order.Data.State == "submitting" || order.Data.State == "submitted" || order.Data.State == "partial-filled" || order.Data.State == "partial-canceled" {
		return "", nil
	} else {
		return "done", nil
	}
}

func NewHuobi(
	addressConfig map[string]string,
	feeConfig common.ExchangeFees,
	interf HuobiInterface, blockchain *blockchain.BaseBlockchain,
	signer blockchain.Signer, nonce blockchain.NonceCorpus, storage HuobiStorage,
	minDepositConfig common.ExchangesMinDeposit) *Huobi {

	tokens, pairs, fees, minDeposit := getExchangePairsAndFeesFromConfig(addressConfig, feeConfig, minDepositConfig, "huobi")
	bc, err := huobiblockchain.NewBlockchain(blockchain, signer, nonce)
	if err != nil {
		log.Printf("Cant create Huobi's blockchain: %v", err)
		panic(err)
	}

	huobiObj := Huobi{
		interf,
		pairs,
		tokens,
		common.NewExchangeAddresses(),
		common.NewExchangeInfo(),
		fees,
		bc,
		signer.GetAddress(),
		storage,
		minDeposit,
	}
	huobiObj.FetchTradeHistory()
	huobiServer := huobihttp.NewHuobiHTTPServer(&huobiObj)
	go huobiServer.Run()
	return &huobiObj
}
