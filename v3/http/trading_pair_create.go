package http

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

func fillLiveInfoToTradingPair(
	ccTradingPair *common.CreateCreateTradingPair,
	exchangeTpsMap map[uint64][]common.TradingPairSymbols) error {

	for exchangeID, tps := range exchangeTpsMap {
		exhID := v1common.ExchangeID(v1common.ExchangeName(exchangeID).String())
		centralExh, ok := v1common.SupportedExchanges[exhID]
		if !ok {
			return errors.Errorf("exchange %s not supported", exhID)
		}

		exchangeInfo, err := centralExh.GetLiveExchangeInfos(tps)
		if err != nil {
			return errors.Wrapf(err, "fail to get live exchange info %v", exchangeID)
		}
		for id, info := range exchangeInfo {
			ccTradingPair.TradingPairs[id].MinNotional = info.MinNotional
			ccTradingPair.TradingPairs[id].AmountLimitMax = info.AmountLimit.Max
			ccTradingPair.TradingPairs[id].AmountLimitMin = info.AmountLimit.Min
			ccTradingPair.TradingPairs[id].AmountPrecision = uint64(info.Precision.Amount)
			ccTradingPair.TradingPairs[id].PricePrecision = uint64(info.Precision.Price)
			ccTradingPair.TradingPairs[id].PriceLimitMax = info.PriceLimit.Max
			ccTradingPair.TradingPairs[id].PriceLimitMin = info.PriceLimit.Min
		}
	}
	return nil
}

func (s *Server) createCreateTradingPair(c *gin.Context) {
	var (
		createTradingPair common.CreateCreateTradingPair
		exchangeTpsMap    = make(map[uint64][]common.TradingPairSymbols)
	)

	err := c.ShouldBindJSON(&createTradingPair)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for index, entry := range createTradingPair.TradingPairs {
		var (
			baseSymbol  string
			quoteSymbol string
		)
		if baseSymbol, quoteSymbol, err = s.checkCreateTradingPairEntry(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", index),
				httputil.WithField("quote", entry.Quote), httputil.WithField("base", entry.Base))
			return
		}

		tradingPairSymbol := common.TradingPairSymbols{TradingPair: entry.TradingPair}
		tradingPairSymbol.BaseSymbol = baseSymbol
		tradingPairSymbol.QuoteSymbol = quoteSymbol
		tradingPairSymbol.ID = uint64(index)

		exchangeTpsMap[entry.ExchangeID] = append(exchangeTpsMap[entry.ExchangeID], tradingPairSymbol)
	}

	err = fillLiveInfoToTradingPair(&createTradingPair, exchangeTpsMap)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	id, err := s.storage.CreateCreateTradingPair(createTradingPair)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkCreateTradingPairEntry(createEntry common.CreateTradingPairEntry) (string, string, error) {
	var (
		ok           bool
		quoteAssetEx common.AssetExchange
		baseAssetEx  common.AssetExchange
	)

	base, err := s.storage.GetAsset(createEntry.Base)
	if err != nil {
		return "", "", errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", createEntry.Base)
	}
	quote, err := s.storage.GetAsset(createEntry.Quote)
	if err != nil {
		return "", "", errors.Wrapf(common.ErrBaseAssetInvalid, "quote id: %v", createEntry.Quote)
	}

	if !quote.IsQuote {
		return "", "", errors.Wrap(common.ErrQuoteAssetInvalid, "quote asset should have is_quote=true")
	}

	if baseAssetEx, ok = getAssetExchangeByExchangeID(base, createEntry.ExchangeID); !ok {
		return "", "", errors.Wrap(common.ErrBaseAssetInvalid, "exchange id not found")
	}

	if quoteAssetEx, ok = getAssetExchangeByExchangeID(base, createEntry.ExchangeID); !ok {
		return "", "", errors.Wrap(common.ErrQuoteAssetInvalid, "exchange id not found")
	}
	return baseAssetEx.Symbol, quoteAssetEx.Symbol, nil
}

func getAssetExchangeByExchangeID(asset common.Asset, exchangeID uint64) (common.AssetExchange, bool) {
	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == exchangeID {
			return exchange, true
		}
	}
	return common.AssetExchange{}, false
}

func (s *Server) getCreateTradingPairs(c *gin.Context) {
	result, err := s.storage.GetCreateTradingPairs()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getCreateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetCreateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmCreateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmCreateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectCreateTradingPair(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectCreateTradingPair(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
