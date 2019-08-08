package http

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"

	v1common "github.com/KyberNetwork/reserve-data/common"
)

func getAssetExchange(assets []common.Asset, assetID, exchangeID uint64) (common.AssetExchange, error) {
	for _, asset := range assets {
		if asset.ID == assetID {
			for _, assetExchange := range asset.Exchanges {
				if assetExchange.ExchangeID == exchangeID {
					return assetExchange, nil
				}
			}
		}
	}
	return common.AssetExchange{}, fmt.Errorf("AssetExchange not found, asset=%d exchange=%d", assetID, exchangeID)
}

func (s *Server) fillTradingPair(ccAsset *common.CreateCreateAsset) error {
	// skip if we have no exchanges enabled.
	if len(v1common.SupportedExchanges) == 0 {
		return nil
	}
	// we fill trading pair parameters that receive from exchange here

	assets, err := s.storage.GetAssets()
	if err != nil {
		return err
	}
	for _, asset := range ccAsset.AssetInputs {
		for _, assetExchange := range asset.Exchanges {
			var tps []common.TradingPairSymbols
			var tradingPairID = uint64(1)
			// a pseudo tradingPairID, because the requested trading pair has not created yet.
			// and keep increase for each pair,
			// we fetch info for all trading pair in AssetExchange, mean one centralExh per round.
			for _, tp := range assetExchange.TradingPairs {
				tradingPairSymbol := common.TradingPairSymbols{TradingPair: tp}
				if tp.Quote == 0 { // current asset is quote in trading pair, so need to fill the Base
					tradingPairSymbol.ID = tradingPairID
					tradingPairID++
					tradingPairSymbol.QuoteSymbol = assetExchange.Symbol
					base, err := getAssetExchange(assets, tp.Base, assetExchange.ExchangeID)
					if err != nil {
						return err
					}
					tradingPairSymbol.BaseSymbol = base.Symbol
				}
				if tp.Base == 0 { // current asset is Base in trading pair, so need to fill the Quote
					tradingPairSymbol.ID = tradingPairID
					tradingPairID++
					tradingPairSymbol.BaseSymbol = assetExchange.Symbol
					quote, err := getAssetExchange(assets, tp.Quote, assetExchange.ExchangeID)
					if err != nil {
						return err
					}
					tradingPairSymbol.QuoteSymbol = quote.Symbol
				}
				tps = append(tps, tradingPairSymbol)
			}
			exhID := v1common.ExchangeID(v1common.ExchangeName(assetExchange.ExchangeID).String())
			centralExh, ok := v1common.SupportedExchanges[exhID]
			if !ok {
				return fmt.Errorf("exchange %s not supported", exhID)
			}

			exInfo, err := centralExh.GetLiveExchangeInfos(tps)
			if err != nil {
				return err
			}
			// because we generate pseudo tradingPairID, like 1,2,3,4
			// fill what we received into request trading pair.
			tradingPairID = uint64(1)
			for idx := range assetExchange.TradingPairs {
				// TODO what if live info does not return all we asked?
				if info, ok := exInfo[tradingPairID]; ok {
					assetExchange.TradingPairs[idx].MinNotional = info.MinNotional
					assetExchange.TradingPairs[idx].AmountLimitMax = info.AmountLimit.Max
					assetExchange.TradingPairs[idx].AmountLimitMin = info.AmountLimit.Min
					assetExchange.TradingPairs[idx].AmountPrecision = uint64(info.Precision.Amount)
					assetExchange.TradingPairs[idx].PricePrecision = uint64(info.Precision.Price)
					assetExchange.TradingPairs[idx].PriceLimitMax = info.PriceLimit.Max
					assetExchange.TradingPairs[idx].PriceLimitMin = info.PriceLimit.Min
					tradingPairID++
				}
			}
		}
	}
	return nil
}

func (s *Server) createCreateAsset(c *gin.Context) {
	var createAsset common.CreateCreateAsset

	err := c.ShouldBindJSON(&createAsset)

	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for id, entry := range createAsset.AssetInputs {
		if err = s.checkCreateAssetEntry(entry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithField("index", id))
			return
		}
	}

	err = s.fillTradingPair(&createAsset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	id, err := s.storage.CreateCreateAsset(createAsset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkCreateAssetEntry(createEntry common.CreateAssetEntry) error {
	if createEntry.Rebalance && createEntry.RebalanceQuadratic == nil {
		return common.ErrRebalanceQuadraticMissing
	}

	if createEntry.Rebalance && createEntry.Target == nil {
		return common.ErrAssetTargetMissing
	}

	if createEntry.SetRate != common.SetRateNotSet && createEntry.PWI == nil {
		return common.ErrPWIMissing
	}

	for _, exchange := range createEntry.Exchanges {
		if common.IsZeroAddress(exchange.DepositAddress) && createEntry.Transferable {
			return errors.Wrapf(common.ErrDepositAddressMissing, "exchange %v", exchange.Symbol)
		}

		for _, tradingPair := range exchange.TradingPairs {

			if tradingPair.Base != 0 && tradingPair.Quote != 0 {
				return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
			}

			if tradingPair.Base == 0 && tradingPair.Quote == 0 {
				return errors.Wrapf(common.ErrBadTradingPairConfiguration, "base id:%v quote id:%v", tradingPair.Base, tradingPair.Quote)
			}

			if tradingPair.Base == 0 {
				quoteAsset, err := s.storage.GetAsset(tradingPair.Quote)
				if err != nil {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
				if !quoteAsset.IsQuote {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
			}

			if tradingPair.Quote == 0 {
				_, err := s.storage.GetAsset(tradingPair.Base)
				if err != nil {
					return errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", tradingPair.Base)
				}

				if !createEntry.IsQuote {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
			}
		}
	}

	return nil
}

func (s *Server) getCreateAssets(c *gin.Context) {
	result, err := s.storage.GetCreateAssets()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getCreateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetCreateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) confirmCreateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmCreateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) rejectCreateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"exists"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectCreateAsset(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
