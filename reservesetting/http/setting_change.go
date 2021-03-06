package http

import (
	"fmt"
	"reflect"

	"github.com/KyberNetwork/reserve-data/common/ethutil"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"

	"github.com/KyberNetwork/reserve-data/common/feed"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/common"
)

func (s *Server) validateChangeEntry(e common.SettingChangeType, changeType common.ChangeType) error {
	var (
		err error
	)

	switch changeType {
	case common.ChangeTypeCreateAsset:
		err = s.checkCreateAssetParams(*(e.(*common.CreateAssetEntry)))
	case common.ChangeTypeUpdateAsset:
		err = s.checkUpdateAssetParams(*(e.(*common.UpdateAssetEntry)))
	case common.ChangeTypeCreateAssetExchange:
		err = s.checkCreateAssetExchangeParams(*(e.(*common.CreateAssetExchangeEntry)))
	case common.ChangeTypeUpdateAssetExchange:
		err = s.checkUpdateAssetExchangeParams(*(e.(*common.UpdateAssetExchangeEntry)))
	case common.ChangeTypeCreateTradingPair:
		_, _, err = s.checkCreateTradingPairParams(*(e.(*common.CreateTradingPairEntry)))
	case common.ChangeTypeChangeAssetAddr:
		err = s.checkChangeAssetAddressParams(*e.(*common.ChangeAssetAddressEntry))
	case common.ChangeTypeUpdateExchange:
		err = s.checkUpdateExchangeParams(*e.(*common.UpdateExchangeEntry))
	case common.ChangeTypeDeleteTradingPair:
		err = s.checkDeleteTradingPairParams(*e.(*common.DeleteTradingPairEntry))
	case common.ChangeTypeDeleteAssetExchange:
		err = s.checkDeleteAssetExchangeParams(*e.(*common.DeleteAssetExchangeEntry))
	case common.ChangeTypeUpdateStableTokenParams:
		return nil
	case common.ChangeTypeSetFeedConfiguration:
		err = s.checkSetFeedConfigurationParams(*e.(*common.SetFeedConfigurationEntry))
	case common.ChangeTypeUpdateTradingPair:
		return nil
	default:
		return errors.Errorf("unknown type of setting change: %v", reflect.TypeOf(e))
	}
	return err
}

func (s *Server) fillLiveInfoSettingChange(settingChange *common.SettingChange) error {
	assets, err := s.storage.GetAssets()
	if err != nil {
		return err
	}

	for i, o := range settingChange.ChangeList {
		switch o.Type {
		case common.ChangeTypeCreateAsset:
			asset := o.Data.(*common.CreateAssetEntry)
			for index, assetExchange := range asset.Exchanges {
				err = s.fillLiveInfoAssetExchange(assets, assetExchange.ExchangeID, assetExchange.TradingPairs, assetExchange.Symbol, assetExchange.AssetID)
				if err != nil {
					return fmt.Errorf("position %d, error: %v", i, err)
				}
				withdrawFee, err := s.withdrawFeeFromExchange(assetExchange.ExchangeID, assetExchange.Symbol)
				if err != nil {
					return fmt.Errorf("position %d, error: %v", i, err)
				}
				asset.Exchanges[index].WithdrawFee = withdrawFee
			}
		case common.ChangeTypeCreateTradingPair:
			entry := o.Data.(*common.CreateTradingPairEntry)
			baseSymbol, quoteSymbol, err := s.checkCreateTradingPairParams(*entry)
			if err != nil {
				return err
			}
			tradingPairSymbol := common.TradingPairSymbols{TradingPair: entry.TradingPair}
			tradingPairSymbol.BaseSymbol = baseSymbol
			tradingPairSymbol.QuoteSymbol = quoteSymbol
			tradingPairSymbol.ID = rtypes.TradingPairID(1) // mock one
			exhID := entry.ExchangeID
			centralExh, ok := s.supportedExchanges[exhID]
			if !ok {
				return errors.Errorf("position %d, exchange %s not supported", i, exhID)
			}
			exchangeInfo, err := centralExh.GetLiveExchangeInfos([]common.TradingPairSymbols{tradingPairSymbol})
			if err != nil {
				return fmt.Errorf("position %d, error: %v", i, err)
			}
			info := exchangeInfo[tradingPairSymbol.ID]
			entry.MinNotional = info.MinNotional
			entry.AmountLimitMax = info.AmountLimit.Max
			entry.AmountLimitMin = info.AmountLimit.Min
			entry.AmountPrecision = uint64(info.Precision.Amount)
			entry.PricePrecision = uint64(info.Precision.Price)
			entry.PriceLimitMax = info.PriceLimit.Max
			entry.PriceLimitMin = info.PriceLimit.Min
		case common.ChangeTypeCreateAssetExchange:
			assetExchange := o.Data.(*common.CreateAssetExchangeEntry)
			err = s.fillLiveInfoAssetExchange(assets, assetExchange.ExchangeID, assetExchange.TradingPairs, assetExchange.Symbol, assetExchange.AssetID)
			if err != nil {
				return fmt.Errorf("position %d, error: %v", i, err)
			}
			withdrawFee, err := s.withdrawFeeFromExchange(assetExchange.ExchangeID, assetExchange.Symbol)
			if err != nil {
				return fmt.Errorf("position %d, error: %v", i, err)
			}
			assetExchange.WithdrawFee = withdrawFee
		}
	}
	return nil
}

func (s *Server) withdrawFeeFromExchange(exchangeID rtypes.ExchangeID, assetSymbol string) (float64, error) {
	exhID := exchangeID
	centralExh, ok := s.supportedExchanges[exhID]
	if !ok {
		return 0, errors.Errorf("exchange %s not supported", exhID)
	}
	return centralExh.GetLiveWithdrawFee(assetSymbol)
}

func (s *Server) fillLiveInfoAssetExchange(assets []common.Asset, exchangeID rtypes.ExchangeID, tradingPairs []common.TradingPair, assetSymbol string, assetID rtypes.AssetID) error {
	exhID := exchangeID
	centralExh, ok := s.supportedExchanges[exhID]
	if !ok {
		return errors.Errorf("exchange %s not supported", exhID)
	}
	var tps []common.TradingPairSymbols
	index := rtypes.TradingPairID(1)
	for idx, tradingPair := range tradingPairs {
		tradingPairSymbol := common.TradingPairSymbols{TradingPair: tradingPair}
		tradingPairSymbol.ID = index
		if tradingPair.Quote == 0 {
			tradingPairSymbol.QuoteSymbol = assetSymbol
			base, err := getAssetExchange(assets, tradingPair.Base, exchangeID)
			if err != nil {
				return err
			}
			tradingPairSymbol.BaseSymbol = base.Symbol
			if assetID != 0 {
				tradingPairs[idx].Quote = assetID
			}
		}
		if tradingPair.Base == 0 {
			tradingPairSymbol.BaseSymbol = assetSymbol
			quote, err := getAssetExchange(assets, tradingPair.Quote, exchangeID)
			if err != nil {
				return err
			}
			tradingPairSymbol.QuoteSymbol = quote.Symbol
			if assetID != 0 {
				tradingPairs[idx].Base = assetID
			}
		}
		tps = append(tps, tradingPairSymbol)
		index++
	}
	exchangeInfo, err := centralExh.GetLiveExchangeInfos(tps)
	if err != nil {
		return err
	}
	tradingPairID := rtypes.TradingPairID(1)
	for idx := range tradingPairs {
		if info, ok := exchangeInfo[tradingPairID]; ok {
			tradingPairs[idx].MinNotional = info.MinNotional
			tradingPairs[idx].AmountLimitMax = info.AmountLimit.Max
			tradingPairs[idx].AmountLimitMin = info.AmountLimit.Min
			tradingPairs[idx].AmountPrecision = uint64(info.Precision.Amount)
			tradingPairs[idx].PricePrecision = uint64(info.Precision.Price)
			tradingPairs[idx].PriceLimitMax = info.PriceLimit.Max
			tradingPairs[idx].PriceLimitMin = info.PriceLimit.Min
			tradingPairID++
		}
	}
	return nil
}
func (s *Server) createSettingChangeWithType(t common.ChangeCatalog) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		s.createSettingChange(ctx, t)
	}
}
func (s *Server) createSettingChange(c *gin.Context, t common.ChangeCatalog) {
	var settingChange common.SettingChange
	if err := c.ShouldBindJSON(&settingChange); err != nil {
		s.l.Warnw("cannot bind data to create setting_change from request", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if len(settingChange.ChangeList) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason("change_list must not empty"))
		return
	}
	for i, o := range settingChange.ChangeList {
		if err := binding.Validator.ValidateStruct(o.Data); err != nil {
			msg := fmt.Sprintf("verify change list failed, position %d, err=%s", i, err)
			httputil.ResponseFailure(c, httputil.WithReason(msg), httputil.WithField("failed-at", o))
			return
		}

		if err := s.validateChangeEntry(o.Data, o.Type); err != nil {
			msg := fmt.Sprintf("verify change list failed, postision %d, err=%s", i, err)
			httputil.ResponseFailure(c, httputil.WithReason(msg), httputil.WithField("failed-at", o))
			return
		}
	}
	if err := s.fillLiveInfoSettingChange(&settingChange); err != nil {
		msg := fmt.Sprintf("validate trading pair info failed, %s", err)
		s.l.Warnw(msg)
		httputil.ResponseFailure(c, httputil.WithReason(msg))
		return
	}
	keyID := s.getKeyIDFromContext(c)
	id, err := s.storage.CreateSettingChange(t, settingChange, keyID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(makeFriendlyMessage(err)))
		return
	}

	// test confirm
	_, err = s.storage.ConfirmSettingChange(id, false)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(makeFriendlyMessage(err)))
		// clean up
		if err = s.storage.RejectSettingChange(id, keyID); err != nil {
			s.l.Errorw("failed to clean up with reject setting change", "err", err)
		}
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	result, err := s.storage.GetSettingChange(rtypes.SettingChangeID(input.ID))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

type getSettingChangeWitTypeFilter struct {
	Status string `form:"status"`
}

func (s *Server) getSettingChangeWithType(t common.ChangeCatalog) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		s.getSettingChanges(ctx, t)
	}
}
func (s *Server) getSettingChanges(c *gin.Context, t common.ChangeCatalog) {
	var (
		query getSettingChangeWitTypeFilter
		err   error
	)
	if err := c.ShouldBindQuery(&query); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	status := common.ChangeStatusPending
	if query.Status != "" {
		status, err = common.ChangeStatusString(query.Status)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	result, err := s.storage.GetSettingChanges(t, status)
	if err != nil {
		s.l.Warnw("failed to get setting changes", "err", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) rejectSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.RejectSettingChange(rtypes.SettingChangeID(input.ID), s.getKeyIDFromContext(c))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) getKeyIDFromContext(c *gin.Context) string {
	return c.Request.Header.Get("UserKeyID")
}

func (s *Server) disapproveSettingChange(c *gin.Context) {
	var (
		input struct {
			ID uint64 `uri:"id" binding:"required"`
		}
	)
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if _, err := s.storage.GetSettingChange(rtypes.SettingChangeID(input.ID)); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	keyID := s.getKeyIDFromContext(c)
	if err := s.storage.DisapproveSettingChange(keyID, input.ID); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) confirmSettingChange(c *gin.Context) {
	var (
		input struct {
			ID uint64 `uri:"id" binding:"required"`
		}
	)
	if err := c.ShouldBindUri(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	keyID := s.getKeyIDFromContext(c)
	change, err := s.storage.GetSettingChange(rtypes.SettingChangeID(input.ID))
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if keyID != "" {
		if change.Proposer == keyID {
			httputil.ResponseFailure(c, httputil.WithError(errors.New("cannot approve for the change that made by yourself")))
			return
		}
		if err := s.storage.ApproveSettingChange(keyID, input.ID); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		listApprovalSettingChange, err := s.storage.GetListApprovalSettingChange(input.ID)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		if len(listApprovalSettingChange) < s.numberApprovalRequired {
			httputil.ResponseSuccess(c)
			return
		}
	}

	if change.ScheduledTime > 0 {
		httputil.ResponseSuccess(c)
		return
	}

	additionalDataReturn, err := s.storage.ConfirmSettingChange(rtypes.SettingChangeID(input.ID), true)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	if s.md != nil {
		// add pair to market data
		if err := s.md.TryToAddFeed(additionalDataReturn); err != nil {
			s.l.Errorw("try to add feed failed", "err", err)
		}
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) checkCreateTradingPairParams(createEntry common.CreateTradingPairEntry) (string, string, error) {
	var (
		ok           bool
		quoteAssetEx common.AssetExchange
		baseAssetEx  common.AssetExchange
	)

	if createEntry.AssetID != createEntry.Quote && createEntry.AssetID != createEntry.Base {
		// assetID must be quote or base
		return "", "", errors.Wrapf(common.ErrBadTradingPairConfiguration, "asset_id must is base or quote")
	}

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
		return "", "", errors.Wrap(common.ErrBaseAssetInvalid, "base asset not config on exchange")
	}

	if quoteAssetEx, ok = getAssetExchangeByExchangeID(quote, createEntry.ExchangeID); !ok {
		return "", "", errors.Wrap(common.ErrQuoteAssetInvalid, "quote asset not config on exchange")
	}
	if s.md != nil {
		if err := s.md.ValidateSymbol(createEntry.ExchangeID, baseAssetEx.Symbol, quoteAssetEx.Symbol); err != nil {
			return "", "", err
		}
	}
	return baseAssetEx.Symbol, quoteAssetEx.Symbol, nil
}

func getAssetExchangeByExchangeID(asset common.Asset, exchangeID rtypes.ExchangeID) (common.AssetExchange, bool) {
	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == exchangeID {
			return exchange,
				true
		}
	}
	return common.AssetExchange{}, false
}

func (s *Server) checkUpdateAssetParams(updateEntry common.UpdateAssetEntry) error {
	asset, err := s.storage.GetAsset(updateEntry.AssetID)
	if err != nil {
		return errors.Wrapf(err, "failed to get asset id: %v from db ", updateEntry.AssetID)
	}

	if updateEntry.SanityInfo != nil && updateEntry.SanityInfo.Path != nil && len(*updateEntry.SanityInfo.Path) == 1 {
		return errors.New("sanity path cannot be a single symbol")
	}

	if updateEntry.Rebalance != nil && *updateEntry.Rebalance {
		if asset.RebalanceQuadratic == nil && updateEntry.RebalanceQuadratic == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrRebalanceQuadraticMissing.Error(), updateEntry.AssetID)
		}

		if asset.Target == nil && updateEntry.Target == nil {
			return errors.Errorf("%v at asset id: %v", common.ErrAssetTargetMissing.Error(), updateEntry.AssetID)
		}
	}

	if updateEntry.SetRate != nil && *updateEntry.SetRate != common.SetRateNotSet && asset.PWI == nil && updateEntry.PWI == nil {
		return errors.Errorf("%v at asset id: %v", common.ErrPWIMissing.Error(), updateEntry.AssetID)
	}

	if updateEntry.SetRate != nil && *updateEntry.SetRate != common.SetRateNotSet {
		if err := s.checkTokenIndices(asset.Address); err != nil {
			return err
		}
	}

	if updateEntry.Transferable != nil && *updateEntry.Transferable {
		for _, exchange := range asset.Exchanges {
			if ethutil.IsZeroAddress(exchange.DepositAddress) {
				return errors.Errorf("%v at asset id: %v and asset_exchange: %v", common.ErrDepositAddressMissing, updateEntry.AssetID, exchange.ID)
			}
		}
	}

	if err := checkFeedWeight(updateEntry.SetRate, updateEntry.FeedWeight); err != nil {
		return err
	}
	return nil
}

func (s *Server) checkUpdateAssetExchangeParams(updateEntry common.UpdateAssetExchangeEntry) error {
	assetExchange, err := s.storage.GetAssetExchange(updateEntry.ID)
	if err != nil {
		return errors.Wrap(err, "asset exchange not found")
	}

	asset, err := s.storage.GetAsset(assetExchange.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	if asset.Transferable && updateEntry.DepositAddress != nil && ethutil.IsZeroAddress(*updateEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	return nil
}

func (s *Server) checkCreateAssetExchangeParams(createEntry common.CreateAssetExchangeEntry) error {
	asset, err := s.storage.GetAsset(createEntry.AssetID)
	if err != nil {
		return errors.Wrap(err, "asset not found")
	}

	_, err = s.storage.GetExchange(createEntry.ExchangeID)
	if err != nil {
		return errors.Wrap(err, "exchange not found")
	}

	for _, exchange := range asset.Exchanges {
		if exchange.ExchangeID == createEntry.ExchangeID {
			return common.ErrAssetExchangeAlreadyExist
		}
	}
	if asset.Transferable && ethutil.IsZeroAddress(createEntry.DepositAddress) {
		return common.ErrDepositAddressMissing
	}
	for _, tradingPair := range createEntry.TradingPairs {
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
			if s.md != nil {
				var quoteAssetExchangeSymbol string
				for _, qae := range quoteAsset.Exchanges {
					if qae.ExchangeID == createEntry.ExchangeID {
						quoteAssetExchangeSymbol = qae.Symbol
						break
					}
				}
				if quoteAssetExchangeSymbol == "" {
					return errors.New(fmt.Sprintf("quote asset didn't have asset exchange, quote id: %v", tradingPair.Quote))
				}
				if err := s.md.ValidateSymbol(createEntry.ExchangeID, createEntry.Symbol, quoteAssetExchangeSymbol); err != nil {
					return err
				}
			}
		}

		if tradingPair.Quote == 0 {
			baseAsset, err := s.storage.GetAsset(tradingPair.Base)
			if err != nil {
				return errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", tradingPair.Base)
			}

			if !asset.IsQuote {
				return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
			}
			if s.md != nil {
				var baseAssetExchangeSymbol string
				for _, bae := range baseAsset.Exchanges {
					if bae.ExchangeID == createEntry.ExchangeID {
						baseAssetExchangeSymbol = bae.Symbol
						break
					}
				}
				if baseAssetExchangeSymbol == "" {
					return errors.New(fmt.Sprintf("base asset didn't have asset exchange, base id: %v", tradingPair.Base))
				}
				if err := s.md.ValidateSymbol(createEntry.ExchangeID, baseAssetExchangeSymbol, createEntry.Symbol); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func getAssetExchange(assets []common.Asset, assetID rtypes.AssetID, exchangeID rtypes.ExchangeID) (common.AssetExchange, error) {
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

func checkFeedWeight(setrate *common.SetRate, feedWeight *common.FeedWeight) error {
	// if feedWeight is nil
	if feedWeight == nil {
		if setrate != nil && (*setrate == common.BTCFeed || *setrate == common.USDFeed) {
			return fmt.Errorf("setrate %s is required feed weight != nil", *setrate)
		}
		// if setrate == nil or setrate != BTCFeed and setrate != USDFeed
		// then we don't have to check feedWeight
		return nil
	}

	// now feedWeight != nil
	// it's possible to update feedWeight without specify Setrate
	// if setrate already set with valid value in DB
	if setrate == nil {
		return errors.New("set FeedWeight requires specify Setrate type")
	}

	// feedWeight only support for BTCFeed and USDFeed
	if *setrate != common.BTCFeed && *setrate != common.USDFeed {
		return fmt.Errorf("setrate type %s does not support feed weight", setrate.String())
	}

	// check if FeedWeight is correctly supported
	if *setrate == common.BTCFeed {
		for k := range *feedWeight {
			if _, ok := feed.AllFeeds().BTC[k]; !ok {
				return fmt.Errorf("%s feed is not supported by %s", k, common.BTCFeed.String())
			}
		}
	} else if *setrate == common.USDFeed {
		for k := range *feedWeight {
			if _, ok := feed.AllFeeds().USD[k]; !ok {
				return fmt.Errorf("%s feed is not supported by %s", k, common.USDFeed.String())
			}
		}
	}

	return nil
}

func (s *Server) checkTokenIndices(tokenAddress ethereum.Address) error {
	if s.coreClient != nil { // check to by pass test as local test does not need this
		if err := s.coreClient.CheckTokenIndice(tokenAddress); err != nil {
			return err
		}
		if err := s.coreClient.UpdateTokenIndice(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) checkCreateAssetParams(createEntry common.CreateAssetEntry) error {
	if len(createEntry.SanityInfo.Path) == 1 {
		return errors.New("sanity path cannot be a single symbol")
	}
	if createEntry.SetRate != common.SetRateNotSet {
		if err := s.checkTokenIndices(createEntry.Address); err != nil {
			return err
		}
	}
	if createEntry.Rebalance && createEntry.RebalanceQuadratic == nil {
		return common.ErrRebalanceQuadraticMissing
	}

	if createEntry.Rebalance && createEntry.Target == nil {
		return common.ErrAssetTargetMissing
	}

	if createEntry.SetRate != common.SetRateNotSet && createEntry.PWI == nil {
		return common.ErrPWIMissing
	}

	if err := checkFeedWeight(&createEntry.SetRate, createEntry.FeedWeight); err != nil {
		return err
	}

	for _, exchange := range createEntry.Exchanges {
		if ethutil.IsZeroAddress(exchange.DepositAddress) && createEntry.Transferable {
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
				if s.md != nil {
					var quoteAssetExchangeSymbol string
					for _, qae := range quoteAsset.Exchanges {
						if qae.ExchangeID == exchange.ExchangeID {
							quoteAssetExchangeSymbol = qae.Symbol
							break
						}
					}
					if quoteAssetExchangeSymbol == "" {
						return errors.New(fmt.Sprintf("quote asset didn't have asset exchange, quote id: %v", tradingPair.Quote))
					}
					if err := s.md.ValidateSymbol(exchange.ExchangeID, exchange.Symbol, quoteAssetExchangeSymbol); err != nil {
						return err
					}
				}
			}

			if tradingPair.Quote == 0 {
				baseAsset, err := s.storage.GetAsset(tradingPair.Base)
				if err != nil {
					return errors.Wrapf(common.ErrBaseAssetInvalid, "base id: %v", tradingPair.Base)
				}

				if !createEntry.IsQuote {
					return errors.Wrapf(common.ErrQuoteAssetInvalid, "quote id: %v", tradingPair.Quote)
				}
				if s.md != nil {
					var baseAssetExchangeSymbol string
					for _, bae := range baseAsset.Exchanges {
						if bae.ExchangeID == exchange.ExchangeID {
							baseAssetExchangeSymbol = bae.Symbol
							break
						}
					}
					if baseAssetExchangeSymbol == "" {
						return errors.New(fmt.Sprintf("base asset didn't have asset exchange, base id: %v", tradingPair.Base))
					}
					if err := s.md.ValidateSymbol(exchange.ExchangeID, baseAssetExchangeSymbol, exchange.Symbol); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Server) checkChangeAssetAddressParams(changeAssetAddressEntry common.ChangeAssetAddressEntry) error {
	asset, err := s.storage.GetAsset(changeAssetAddressEntry.ID)
	if err != nil {
		return err
	}
	if asset.Address == changeAssetAddressEntry.Address {
		return common.ErrAddressExists
	}
	return nil
}

func (s *Server) checkUpdateExchangeParams(updateExchangeEntry common.UpdateExchangeEntry) error {
	// check if exchange exist
	_, err := s.storage.GetExchange(updateExchangeEntry.ExchangeID)
	return err
}

func (s *Server) checkSetFeedConfigurationParams(setFeedConfigurationEntry common.SetFeedConfigurationEntry) error {
	switch setFeedConfigurationEntry.SetRate {
	case common.BTCFeed:
		if _, ok := feed.AllFeeds().BTC[setFeedConfigurationEntry.Name]; ok {
			return nil
		}
		return fmt.Errorf("set feed configuration does not match, feed name: %s, set rate value: %s", setFeedConfigurationEntry.Name, setFeedConfigurationEntry.SetRate.String())
	case common.USDFeed:
		if _, ok := feed.AllFeeds().USD[setFeedConfigurationEntry.Name]; ok {
			return nil
		}
		return fmt.Errorf("set feed configuration does not match, feed name: %s, set rate value: %s", setFeedConfigurationEntry.Name, setFeedConfigurationEntry.SetRate.String())
	case common.GoldFeed:
		if _, ok := feed.AllFeeds().Gold[setFeedConfigurationEntry.Name]; ok {
			return nil
		}
		return fmt.Errorf("set feed configuration does not match, feed name: %s, set rate value: %s", setFeedConfigurationEntry.Name, setFeedConfigurationEntry.SetRate.String())
	}

	return fmt.Errorf("feed does not exist, feed=%s", setFeedConfigurationEntry.Name)
}

func (s *Server) getNumberApprovalRequired(c *gin.Context) {
	httputil.ResponseSuccess(c, httputil.WithData(s.numberApprovalRequired))
}
