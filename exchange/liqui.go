package exchange

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
	uuid "github.com/satori/go.uuid"
)

const (
	DEPOSIT_WAITING_TIME  uint64 = 10 * 60 * 1000
	WITHDRAW_WAITING_TIME uint64 = 5 * 60 * 1000
)

type Liqui struct {
	interf      LiquiInterface
	pairs       []common.TokenPair
	addresses   map[string]ethereum.Address
	databusType string
}

func (self *Liqui) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Liqui) Address(token common.Token) (ethereum.Address, bool) {
	addr, supported := self.addresses[token.ID]
	return addr, supported
}

func (self *Liqui) UpdateAllDepositAddresses(address string) {
	for k, _ := range self.addresses {
		self.addresses[k] = ethereum.HexToAddress(address)
	}
}

func (self *Liqui) UpdateDepositAddress(token common.Token, address string) {
	self.addresses[token.ID] = ethereum.HexToAddress(address)
}

func (self *Liqui) UpdateFetcherDatabusType(databusType string) {
	self.databusType = databusType
}

func (self *Liqui) ID() common.ExchangeID {
	return common.ExchangeID("liqui")
}

func (self *Liqui) DatabusType() string {
	return self.databusType
}

func (self *Liqui) TokenPairs() []common.TokenPair {
	return self.pairs
}

func (self *Liqui) Name() string {
	return "liqui"
}

func (self *Liqui) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	return self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)
}

func (self *Liqui) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	err := self.interf.Withdraw(token, amount, address, timepoint)
	return uuid.NewV4().String(), err
}

func (self *Liqui) CancelOrder(id common.ActivityID) error {
	result, err := self.interf.CancelOrder(id.EID)
	if err != nil {
		return err
	}
	if result.Success != 1 {
		return errors.New("Couldn't cancel order id " + id.EID + " err: " + result.Error)
	}
	return nil
}

func (self *Liqui) FetchOrderData(timepoint uint64) (common.OrderEntry, error) {
	result := common.OrderEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Data = []common.Order{}
	resp_data, err := self.interf.ActiveOrders(timepoint)
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if resp_data.Success == 1 {
			for id, order := range resp_data.Return {
				tokens := strings.Split(order.Pair, "_")
				result.Data = append(result.Data, common.Order{
					ID:          id,
					Base:        strings.ToUpper(tokens[0]),
					Quote:       strings.ToUpper(tokens[1]),
					OrderId:     id,
					Price:       order.Rate,
					OrigQty:     order.Amount,
					ExecutedQty: 0,
					TimeInForce: "GTC",
					Type:        "LIMIT",
					Side:        order.Type,
					Time:        order.Timestamp,
				})
			}
		} else {
			result.Valid = false
			result.Error = resp_data.Error
		}
	}
	return result, nil
}

func (self *Liqui) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	resp_data, err := self.interf.GetInfo(timepoint)
	result.ReturnTime = common.GetTimestamp()
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	} else {
		if resp_data.Success == 1 {
			balances := resp_data.Return["funds"]
			result.AvailableBalance = map[string]float64{}
			result.LockedBalance = map[string]float64{}
			result.DepositBalance = map[string]float64{}
			for tokenID, _ := range common.SupportedTokens {
				result.AvailableBalance[tokenID] = balances[strings.ToLower(tokenID)]
				// TODO: need to take open order into account
				result.LockedBalance[tokenID] = 0
				result.DepositBalance[tokenID] = 0
			}
		} else {
			result.Valid = false
			result.Error = resp_data.Error
		}
	}
	return result, nil
}

func (self *Liqui) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	result := map[common.TokenPairID]common.ExchangePrice{}
	pairs_str := []string{}
	for _, pair := range self.pairs {
		pairs_str = append(pairs_str, fmt.Sprintf("%s_%s", pair.Base.ID, pair.Quote.ID))
	}
	timestamp := common.Timestamp(fmt.Sprintf("%d", timepoint))
	log.Printf("depth: %s - %s\n",
		strings.ToLower(strings.Join(pairs_str, "-")),
		timepoint,
	)
	resp_data, err := self.interf.Depth(
		strings.ToLower(strings.Join(pairs_str, "-")),
		timepoint,
	)
	returnTime := common.GetTimestamp()
	if err != nil {
		for _, pair := range self.pairs {
			one_pair_result := common.ExchangePrice{}
			one_pair_result.Timestamp = timestamp
			one_pair_result.ReturnTime = returnTime
			one_pair_result.Valid = false
			one_pair_result.Error = err.Error()
			result[pair.PairID()] = one_pair_result
		}
	} else {
		for _, pair := range self.pairs {
			one_pair_result := common.ExchangePrice{}
			one_pair_result.Timestamp = timestamp
			one_pair_result.ReturnTime = returnTime
			one_pair_result.Valid = true
			one_data := resp_data[fmt.Sprintf(
				"%s_%s",
				strings.ToLower(pair.Base.ID),
				strings.ToLower(pair.Quote.ID),
			)]
			for _, buy := range one_data["bids"] {
				one_pair_result.Bids = append(
					one_pair_result.Bids,
					common.PriceEntry{
						buy[1],
						buy[0],
					},
				)
			}
			for _, sell := range one_data["asks"] {
				one_pair_result.Asks = append(
					one_pair_result.Asks,
					common.PriceEntry{
						sell[1],
						sell[0],
					},
				)
			}
			result[pair.PairID()] = one_pair_result
		}
	}
	return result, err
}

func (self *Liqui) FetchPriceDataUsingSocket() (map[common.TokenPairID]common.ExchangePrice, error) {
	// Liqui not support for socket
	panic("Liqui does not support socket")
}

func (self *Liqui) DepositStatus(id common.ActivityID, timepoint uint64) (string, error) {
	timestamp := id.Timepoint
	if timepoint-timestamp/uint64(time.Millisecond) > DEPOSIT_WAITING_TIME {
		return "done", nil
	} else {
		return "", nil
	}
	return "", errors.New("Not implemented yet")
}

// Liqui should not work properly because of lack of txid
func (self *Liqui) WithdrawStatus(id common.ActivityID, timepoint uint64) (string, string, error) {
	timestamp := id.Timepoint
	if timepoint-timestamp/uint64(time.Millisecond) > WITHDRAW_WAITING_TIME {
		return "done", "", nil
	} else {
		return "", "", nil
	}
	return "", "", errors.New("Not implemented yet")
}

func (self *Liqui) OrderStatus(id common.ActivityID, timepoint uint64) (string, error) {
	result, err := self.interf.OrderInfo(id.EID, timepoint)
	if err != nil {
		return "", err
	} else {
		if result.Success != 1 {
			for _, v := range result.Return {
				if v.Status == 0 {
					return "", nil
				} else if v.Status == 1 {
					return "done", nil
				} else if v.Status == 2 {
					return "failed", nil
				} else if v.Status == 3 {
					return "failed", nil
				}
			}
			panic("Malformed response from liqui")
		} else {
			return "", errors.New(result.Error)
		}
	}
}

func NewLiqui(interf LiquiInterface) *Liqui {
	return &Liqui{
		interf,
		[]common.TokenPair{
			common.MustCreateTokenPair("OMG", "ETH"),
			common.MustCreateTokenPair("DGD", "ETH"),
			common.MustCreateTokenPair("CVC", "ETH"),
			common.MustCreateTokenPair("MCO", "ETH"),
			common.MustCreateTokenPair("GNT", "ETH"),
			common.MustCreateTokenPair("ADX", "ETH"),
			common.MustCreateTokenPair("EOS", "ETH"),
			common.MustCreateTokenPair("PAY", "ETH"),
			common.MustCreateTokenPair("BAT", "ETH"),
			common.MustCreateTokenPair("KNC", "ETH"),
		},
		map[string]ethereum.Address{},
		"http",
	}
}
