package exchange

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	commonv3 "github.com/KyberNetwork/reserve-data/reservesetting/common"
	"go.uber.org/zap"
)

// HuobiLive is LiveExchange for huobi
type HuobiLive struct {
	interf          HuobiInterface
	sugar           *zap.SugaredLogger
	mu              *sync.RWMutex
	allAssetDetails map[string]HuobiChain
}

// NewHuobiLive return new HuobiLive instance
func NewHuobiLive(interf HuobiInterface) *HuobiLive {
	return &HuobiLive{
		sugar:           zap.S(),
		interf:          interf,
		mu:              &sync.RWMutex{},
		allAssetDetails: make(map[string]HuobiChain),
	}
}

// RunUpdateAssetDetails just update asset info of eth and erc20 tokens
// eth has chain is "ETH" and erc20 tokens has base chain is ETH
func (hl *HuobiLive) RunUpdateAssetDetails(interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		func() {
			var (
				rawAllAssetDetails []HuobiAssetDetail
				err                error
			)
			for i := 0; i < 2; i++ {
				rawAllAssetDetails, err = hl.interf.GetAllAssetDetail()
				if err != nil {
					time.Sleep(3 * time.Second)
					continue
				}
				break
			}
			if err != nil {
				hl.sugar.Errorw("cannot get asset detail", "err", err)
				return
			}
			allAssetDetails := make(map[string]HuobiChain)
			for _, rawAssetDetail := range rawAllAssetDetails {
				for _, c := range rawAssetDetail.Chains { // TODO: this won't work for BSC chain, assume not support
					if c.Chain == "eth" || c.BaseChain == "ETH" {
						allAssetDetails[strings.ToUpper(rawAssetDetail.Currency)] = c
						break
					}
				}
			}
			hl.mu.Lock()
			hl.allAssetDetails = allAssetDetails
			hl.mu.Unlock()
		}()
		<-t.C
	}
}

// GetLiveWithdrawFee ...
func (hl *HuobiLive) GetLiveWithdrawFee(asset string) (float64, error) {
	hl.mu.RLock()
	defer hl.mu.RUnlock()
	assetDetail, ok := hl.allAssetDetails[asset]
	if !ok {
		return 0, fmt.Errorf("asset detail is not available, asset: %s", asset)
	}
	withdrawFee, err := strconv.ParseFloat(assetDetail.TransactFeeWithdraw, 64)
	if err != nil {
		return 0, fmt.Errorf("withdraw fee from huobi api is not valid, asset: %s, fee: %s", asset, assetDetail.TransactFeeWithdraw)
	}
	return withdrawFee, nil
}

// GetLiveExchangeInfos querry the Exchange Endpoint for exchange precision and limit of a list of tokenPairIDs
// It return error if occurs.
func (hl *HuobiLive) GetLiveExchangeInfos(pairs []commonv3.TradingPairSymbols) (common.ExchangeInfo, error) {
	result := make(common.ExchangeInfo)
	exchangeInfo, err := hl.interf.GetExchangeInfo()
	if err != nil {
		return result, err
	}
	for _, pair := range pairs {
		exchangePrecisionLimit, ok := hl.getPrecisionLimitFromSymbols(pair, exchangeInfo)
		if !ok {
			return result, fmt.Errorf("huobi exchange reply doesn't contain token pair %s",
				strings.ToUpper(fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol)))
		}
		result[pair.ID] = exchangePrecisionLimit
	}
	return result, nil
}

// getPrecisionLimitFromSymbols find the pairID amongs symbols from exchanges,
// return ExchangePrecisionLimit of that pair and true if the pairID exist amongs symbols, false if otherwise
func (hl *HuobiLive) getPrecisionLimitFromSymbols(pair commonv3.TradingPairSymbols, symbols HuobiExchangeInfo) (common.ExchangePrecisionLimit, bool) {
	var result common.ExchangePrecisionLimit
	pairName := strings.ToUpper(fmt.Sprintf("%s%s", pair.BaseSymbol, pair.QuoteSymbol))
	for _, symbol := range symbols.Data {
		symbolName := strings.ToUpper(symbol.Base + symbol.Quote)
		if symbolName == pairName {
			result.Precision.Amount = symbol.AmountPrecision
			result.Precision.Price = symbol.PricePrecision
			result.MinNotional = symbol.MinOrderValue
			return result, true
		}
	}
	return result, false
}
