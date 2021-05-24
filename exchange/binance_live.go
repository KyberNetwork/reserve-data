package exchange

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/bcnetwork"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"
)

//BinanceLive implement live info for binance
type BinanceLive struct {
	sugar       *zap.SugaredLogger
	mu          *sync.RWMutex
	interf      BinanceInterface
	allCoinInfo map[string]CoinInfo
}

// NewBinanceLive return new BinanceLive instance
func NewBinanceLive(interf BinanceInterface) *BinanceLive {
	return &BinanceLive{
		sugar:       zap.S(),
		mu:          &sync.RWMutex{},
		interf:      interf,
		allCoinInfo: make(map[string]CoinInfo),
	}
}

// RunUpdateAssetDetails run interval get asset detail
func (bl *BinanceLive) RunUpdateAssetDetails(interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		func() {
			var (
				allCoinInfo map[string]CoinInfo
				err         error
			)
			for i := 0; i < 2; i++ {
				allCoinInfo, err = bl.interf.AllCoinInfo()
				if err != nil {
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}
			if err != nil {
				bl.sugar.Errorw("cannot get asset detail", "err", err)
				return
			}
			bl.mu.Lock()
			bl.allCoinInfo = allCoinInfo
			bl.mu.Unlock()
		}()
		<-t.C
	}
}

// GetLiveWithdrawFee ...
func (bl *BinanceLive) GetLiveWithdrawFee(asset string) (float64, error) {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	coinInfo, ok := bl.allCoinInfo[asset]
	if !ok {
		return 0, fmt.Errorf("coin info of asset is not available, asset: %s", asset)
	}
	nw := bcnetwork.GetPreConfig().Network
	for _, listed := range coinInfo.NetworkList {
		if listed.Network == nw {
			fee, err := decimal.NewFromString(listed.WithdrawFee)
			if err != nil {
				return 0, fmt.Errorf("parse withdraw fee[%s] failed: %w", listed.WithdrawFee, err)
			}
			v, _ := fee.Float64()
			return v, nil
		}
	}
	return 0, fmt.Errorf("no entry withdraw fee for %s network %s", asset, nw)
}

// GetLiveExchangeInfos queries the Exchange Endpoint for exchange precision and limit of a certain pair ID
// It return error if occurs.
func (bl *BinanceLive) GetLiveExchangeInfos(pairs []commonv3.TradingPairSymbols) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := bl.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	symbols := exchangeInfo.Symbols
	for _, pair := range pairs {
		exchangePrecisionLimit, ok := bl.getPrecisionLimitFromSymbols(pair, symbols)
		if !ok {
			return result, fmt.Errorf("binance exchange reply doesn't contain token pair '%s'",
				strings.ToUpper(fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol)))
		}
		result[pair.ID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (bl *BinanceLive) getPrecisionLimitFromSymbols(pair commonv3.TradingPairSymbols, symbols []BinanceSymbol) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol))
	for _, symbol := range symbols {
		if strings.ToUpper(symbol.Symbol) == pairName {
			//update precision
			result.Precision.Amount = symbol.BaseAssetPrecision
			result.Precision.Price = symbol.QuoteAssetPrecision
			// update limit
			for _, filter := range symbol.Filters {
				if filter.FilterType == "LOT_SIZE" {
					// update amount min
					minQuantity, _ := strconv.ParseFloat(filter.MinQuantity, 64)
					result.AmountLimit.Min = minQuantity
					// update amount max
					maxQuantity, _ := strconv.ParseFloat(filter.MaxQuantity, 64)
					result.AmountLimit.Max = maxQuantity
					result.Precision.Amount = bl.precisionFromStepSize(filter.StepSize)
				}

				if filter.FilterType == "PRICE_FILTER" {
					// update price min
					minPrice, _ := strconv.ParseFloat(filter.MinPrice, 64)
					result.PriceLimit.Min = minPrice
					// update price max
					maxPrice, _ := strconv.ParseFloat(filter.MaxPrice, 64)
					result.PriceLimit.Max = maxPrice
					result.Precision.Price = bl.precisionFromStepSize(filter.TickSize)
				}

				if filter.FilterType == "MIN_NOTIONAL" {
					minNotional, _ := strconv.ParseFloat(filter.MinNotional, 64)
					result.MinNotional = minNotional
				}
			}
			return result, true
		}
	}
	return result, false
}

func (bl *BinanceLive) precisionFromStepSize(stepSize string) int {
	re := regexp.MustCompile("0*$")
	parts := strings.Split(re.ReplaceAllString(stepSize, ""), ".")
	if len(parts) > 1 {
		return len(parts[1])
	}
	return 0
}
