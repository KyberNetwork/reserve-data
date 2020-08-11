package exchange

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/settings"
)

type StableEx struct {
	setting Setting
}

func (se *StableEx) TokenAddresses() (map[string]ethereum.Address, error) {
	// returning admin multisig. In case anyone sent dgx to this address,
	// we can still get it.
	return map[string]ethereum.Address{
		"DGX":  ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
		"WBTC": ethereum.HexToAddress("0xFDF28Bf25779ED4cA74e958d54653260af604C20"),
	}, nil
}

func (se *StableEx) MarshalText() (text []byte, err error) {
	return []byte(se.ID()), nil
}

func (se *StableEx) Address(token common.Token) (ethereum.Address, bool) {
	addrs, err := se.TokenAddresses()
	if err != nil {
		return ethereum.Address{}, false
	}
	addr, supported := addrs[token.ID]
	return addr, supported
}

func (se *StableEx) UpdateDepositAddress(token common.Token, address string) error {
	return errors.New("dgx doesn't support update deposit addresses")
}

func (se *StableEx) UpdatePairsPrecision() error {
	return errors.New("stable exchange does not support update pair precision")
}

func (se *StableEx) GetInfo() (common.ExchangeInfo, error) {
	return se.setting.GetExchangeInfo(settings.StableExchange)
}

func (se *StableEx) GetLiveExchangeInfos(tokenPairIDs []common.TokenPairID) (common.ExchangeInfo, error) {
	zap.S().Warnw("stable_exchange shouldn't come with live exchange info. Return an all 0 result...")
	result := make(common.ExchangeInfo)
	for _, tokenPairID := range tokenPairIDs {
		result[tokenPairID] = common.ExchangePrecisionLimit{
			Precision:   common.TokenPairPrecision{},
			AmountLimit: common.TokenPairAmountLimit{},
			PriceLimit:  common.TokenPairPriceLimit{},
			MinNotional: 0,
		}
	}
	return result, nil
}

func (se *StableEx) GetExchangeInfo(pair common.TokenPairID) (common.ExchangePrecisionLimit, error) {
	exInfo, err := se.setting.GetExchangeInfo(settings.StableExchange)
	if err != nil {
		return common.ExchangePrecisionLimit{}, err
	}
	data, err := exInfo.Get(pair)
	return data, err
}

func (se *StableEx) GetFee() (common.ExchangeFees, error) {
	return se.setting.GetFee(settings.StableExchange)
}

// ID must return the exact string or else simulation will fail
func (se *StableEx) ID() common.ExchangeID {
	return common.ExchangeID(settings.StableExchange.String())
}

func (se *StableEx) TokenPairs() ([]common.TokenPair, error) {
	result := []common.TokenPair{}
	exInfo, err := se.setting.GetExchangeInfo(settings.StableExchange)
	if err != nil {
		return nil, err
	}
	for pair := range exInfo.GetData() {
		pairIDs := strings.Split(string(pair), "-")
		if len(pairIDs) != 2 {
			return result, fmt.Errorf("PairID %s is malformed", string(pair))
		}
		tok1, uErr := se.setting.GetTokenByID(pairIDs[0])
		if uErr != nil {
			return result, fmt.Errorf("cant get Token %s, %s", pairIDs[0], uErr)
		}
		tok2, uErr := se.setting.GetTokenByID(pairIDs[1])
		if uErr != nil {
			return result, fmt.Errorf("cant get Token %s, %s", pairIDs[1], uErr)
		}
		tokPair := common.TokenPair{
			Base:  tok1,
			Quote: tok2,
		}
		result = append(result, tokPair)
	}
	return result, nil
}

func (se *StableEx) Name() string {
	return "stable token exchange"
}

func (se *StableEx) QueryOrder(symbol string, id uint64) (done float64, remaining float64, finished bool, err error) {
	// TODO: see if trade order (a tx to dgx contract) is successful or not
	// - successful: done = order amount, remaining = 0, finished = true, err = nil
	// - failed: done = 0, remaining = order amount, finished = false, err = some error
	// - pending: done = 0, remaining = order amount, finished = false, err = nil
	return 0, 0, false, errors.New("not supported")
}

func (se *StableEx) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (id string, done float64, remaining float64, finished bool, err error) {
	// TODO: communicate with dgx connector to do the trade
	return "not supported", 0, 0, false, errors.New("not supported")
}

func (se *StableEx) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) (string, error) {
	// TODO: communicate with dgx connector to withdraw
	return "not supported", errors.New("not supported")
}

func (se *StableEx) CancelOrder(id, symbol string) error {
	return errors.New("dgx doesn't support trade cancelling")
}

func (se *StableEx) CancelAllOrders(symbol string) error {
	return errors.New("stable exchange doesn't support this type of api")
}

func (se *StableEx) FetchPriceData(timepoint uint64, fetchBTCPrice bool) (map[common.TokenPairID]common.ExchangePrice, error) {
	result := map[common.TokenPairID]common.ExchangePrice{}
	// TODO: Get price data from dgx connector and construct valid orderbooks
	return result, nil
}

func (se *StableEx) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	result.Timestamp = common.Timestamp(fmt.Sprintf("%d", timepoint))
	result.Valid = true
	result.Status = true
	// TODO: Get balance data from dgx connector
	result.ReturnTime = common.GetTimestamp()
	result.AvailableBalance = map[string]float64{"DGX": 0, "ETH": 0}
	result.LockedBalance = map[string]float64{"DGX": 0, "ETH": 0}
	result.DepositBalance = map[string]float64{"DGX": 0, "ETH": 0}
	return result, nil
}

func (se *StableEx) GetTradeHistory(fromTime, toTime uint64) (common.ExchangeTradeHistory, error) {
	return common.ExchangeTradeHistory{}, nil
}

func (se *StableEx) FetchTradeHistory(timepoint uint64) (map[common.TokenPairID][]common.TradeHistory, error) {
	result := map[common.TokenPairID][]common.TradeHistory{}
	// TODO: get trade history
	return result, errors.New("not supported")
}

func (se *StableEx) DepositStatus(id common.ActivityID, txHash, currency string, amount float64, timepoint uint64) (string, error) {
	// TODO: checking txHash status
	return "", errors.New("not supported")
}

func (se *StableEx) WithdrawStatus(id, currency string, amount float64, timepoint uint64) (string, string, float64, error) {
	// TODO: checking id (id is the txhash) status
	return "", "", 0, errors.New("not supported")
}

func (se *StableEx) OrderStatus(id string, base, quote string) (string, error) {
	// TODO: checking id (id is the txhash) status
	return "", errors.New("not supported")
}

func (se *StableEx) GetMinDeposit() (common.ExchangesMinDeposit, error) {
	return se.setting.GetMinDeposit(settings.StableExchange)
}

func (se *StableEx) OpenOrders() ([]common.Order, error) {
	return []common.Order{}, nil
}

func NewStableEx(setting Setting) (*StableEx, error) {
	return &StableEx{
		setting,
	}, nil
}
