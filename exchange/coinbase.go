package exchange

import (
	"fmt"
	"math/big"
	"strconv"
	"sync"

	common3 "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

// Coinbase object definition for coibase exchange
type Coinbase struct {
	interf CoinbaseInterface
	sr     storage.Interface
	l      *zap.SugaredLogger
	//CoinbaseLive
	id common.ExchangeID
}

// Address return address of asset on cointbase
func (c *Coinbase) Address(asset commonv3.Asset) (address common3.Address, supported bool) {
	return common3.Address{}, false
}

// Withdraw withdraw asset from coinbase
func (c *Coinbase) Withdraw(asset commonv3.Asset, amount *big.Int, address common3.Address) (string, error) {
	return "", ErrNotSupport
}

// Trade create order to exchange on coinbase
func (c *Coinbase) Trade(tradeType string, pair commonv3.TradingPairSymbols, rate, amount float64) (id string, done, remaining float64, finished bool, err error) {
	return "", 0, 0, false, ErrNotSupport
}

// CancelOrder cancel open order on coinbase
func (c *Coinbase) CancelOrder(id, base, quote string) error {
	return ErrNotSupport
}

// MarshalText return coinbase exchange name instead of id
func (c *Coinbase) MarshalText() (text []byte, err error) {
	return []byte(c.ID().String()), nil
}

// GetTradeHistory return history of order created on coinbase
func (c *Coinbase) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return nil, nil
}

// GetLiveExchangeInfos return live exchange info from coinbase
func (c *Coinbase) GetLiveExchangeInfos(ps []commonv3.TradingPairSymbols) (common.ExchangeInfo, error) {
	res := make(common.ExchangeInfo) // we just fake result here so coinbase can accept any trading pair.
	for _, tp := range ps {
		res[tp.ID] = common.ExchangePrecisionLimit{
			Precision: common.TokenPairPrecision{
				Amount: 1,
				Price:  1,
			},
			AmountLimit: common.TokenPairAmountLimit{
				Min: 1,
				Max: 1,
			},
			PriceLimit: common.TokenPairPriceLimit{
				Min: 1,
				Max: 1,
			},
			MinNotional: 1,
		}
	}
	return res, nil
}

// ID return exchange id
func (c *Coinbase) ID() common.ExchangeID {
	return c.id
}

// TokenPairs return token pairs supported by exchange
func (c *Coinbase) TokenPairs() ([]commonv3.TradingPairSymbols, error) {
	pairs, err := c.sr.GetTradingPairs(uint64(c.id))
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

// FetchPriceData return orders book from coinbase
func (c *Coinbase) FetchPriceData(timepoint uint64) (map[uint64]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs, err := c.TokenPairs()
	if err != nil {
		return nil, err
	}
	var (
		i int
		x int
	)
	for i < len(pairs) {
		for x = i; x < len(pairs) && x < i+batchSize; x++ {
			wait.Add(1)
			pair := pairs[x]
			go c.FetchOnePairData(&wait, pair, &data, timepoint)
		}
		wait.Wait()
		i = x
	}
	result := map[uint64]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		//if there is conversion error, continue to next key,val
		tokenPairID, ok := key.(uint64)
		if !ok {
			err = fmt.Errorf("key (%v) cannot be asserted to TokenPairID", key)
			return false
		}
		exPrice, ok := value.(common.ExchangePrice)
		if !ok {
			err = fmt.Errorf("value (%v) cannot be asserted to ExchangePrice", value)
			return false
		}
		result[tokenPairID] = exPrice
		return true
	})
	return result, err
}

// FetchOnePairData get orders book for one pair of token
func (c *Coinbase) FetchOnePairData(
	wg *sync.WaitGroup,
	pair commonv3.TradingPairSymbols,
	data *sync.Map,
	timepoint uint64) {

	defer wg.Done()
	result := common.ExchangePrice{}

	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Timestamp = timestamp
	result.Valid = true
	respData, err := c.interf.GetOnePairOrderBook(pair.BaseSymbol, pair.QuoteSymbol)
	returnTime := common.GetTimestamp()
	result.ReturnTime = returnTime
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if respData.Code != 0 || respData.Msg != "" {
			result.Valid = false
			result.Error = fmt.Sprintf("Code: %d, Msg: %s", respData.Code, respData.Msg)
		} else {
			for _, buy := range respData.Bids {
				quantity, _ := strconv.ParseFloat(buy.Size, 64)
				rate, _ := strconv.ParseFloat(buy.Price, 64)
				result.Bids = append(
					result.Bids,
					common.NewPriceEntry(
						quantity,
						rate,
					),
				)
			}
			for _, sell := range respData.Asks {
				quantity, _ := strconv.ParseFloat(sell.Size, 64)
				rate, _ := strconv.ParseFloat(sell.Price, 64)
				result.Asks = append(
					result.Asks,
					common.NewPriceEntry(
						quantity,
						rate,
					),
				)
			}
		}
	}
	data.Store(pair.ID, result)
}

// FetchEBalanceData get balance on coinbase
func (c *Coinbase) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	return common.EBalanceEntry{}, nil // we return empty without error here so fetch won't warning
}

// FetchTradeHistory return trade history
func (c *Coinbase) FetchTradeHistory() {
	// panic("implement me")
}

// OrderStatus return status of an order
func (c *Coinbase) OrderStatus(id string, base, quote string) (string, error) {
	return "", ErrNotSupport
}

// DepositStatus return status of a deposit
func (c *Coinbase) DepositStatus(id common.ActivityID, txHash string, assetID uint64, amount float64, timepoint uint64) (string, error) {
	return "", ErrNotSupport
}

// WithdrawStatus return status of a withdrawal
func (c *Coinbase) WithdrawStatus(id string, assetID uint64, amount float64, timepoint uint64) (string, string, error) {
	return "", "", ErrNotSupport
}

// TokenAddresses return token address on coinbase
func (c *Coinbase) TokenAddresses() (map[common.AssetID]common3.Address, error) {
	return map[common.AssetID]common3.Address{}, nil // we return empty map so FetchAuthDataFromExchange does not treat as error
}

// OpenOrders return open orders on coinbase
func (c *Coinbase) OpenOrders(pair commonv3.TradingPairSymbols) ([]common.Order, error) {
	return nil, nil
}

// NewCoinbase return new Coinbase instance
func NewCoinbase(l *zap.SugaredLogger, id common.ExchangeID, interf CoinbaseInterface, sr storage.Interface) *Coinbase {
	return &Coinbase{
		l:      l,
		id:     id,
		interf: interf,
		sr:     sr,
	}
}
