package http

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	v1common "github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

type settingChangeEntry struct {
	Type common.ChangeType `json:"type"`
	Data json.RawMessage   `json:"data"`
}

type settingChange struct {
	ChangeList []settingChangeEntry `json:"change_list" binding:"required"`
}

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
	case common.ChangeTypeCreateTradingBy:
		err = s.checkCreateTradingByParams(*e.(*common.CreateTradingByEntry))
	case common.ChangeTypeChangeAssetAddr:
		err = s.checkChangeAssetAddressParams(*e.(*common.ChangeAssetAddressEntry))
	case common.ChangeTypeUpdateExchange:
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

	for _, o := range settingChange.ChangeList {
		switch o.Type {
		case common.ChangeTypeCreateAsset:
			asset := o.Data.(common.CreateAssetEntry)
			for _, assetExchange := range asset.Exchanges {
				exhID := v1common.ExchangeID(v1common.ExchangeName(assetExchange.ExchangeID).String())
				centralExh, ok := v1common.SupportedExchanges[exhID]
				if !ok {
					return errors.Errorf("exchange %s not supported", exhID)
				}
				var tps []common.TradingPairSymbols
				index := uint64(1)
				for _, tradingPair := range assetExchange.TradingPairs {
					tradingPairSymbol := common.TradingPairSymbols{TradingPair: tradingPair}
					tradingPairSymbol.ID = index
					if tradingPair.Quote == 0 {
						tradingPairSymbol.QuoteSymbol = assetExchange.Symbol
						base, err := getAssetExchange(assets, tradingPair.Base, assetExchange.ExchangeID)
						if err != nil {
							return err
						}
						tradingPairSymbol.BaseSymbol = base.Symbol
					}
					if tradingPair.Base == 0 {
						tradingPairSymbol.BaseSymbol = assetExchange.Symbol
						quote, err := getAssetExchange(assets, tradingPair.Quote, assetExchange.ExchangeID)
						if err != nil {
							return err
						}
						tradingPairSymbol.BaseSymbol = quote.Symbol
					}
					tps = append(tps, tradingPairSymbol)
					index++
				}
				exchangeInfo, err := centralExh.GetLiveExchangeInfos(tps)
				if err != nil {
					return err
				}
				tradingPairID := uint64(1)
				for idx := range assetExchange.TradingPairs {
					if info, ok := exchangeInfo[tradingPairID]; ok {
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
		case common.ChangeTypeCreateTradingPair:
			entry := o.Data.(common.CreateTradingPairEntry)
			baseSymbol, quoteSymbol, err := s.checkCreateTradingPairParams(entry)
			if err != nil {
				return err
			}
			tradingPairSymbol := common.TradingPairSymbols{TradingPair: entry.TradingPair}
			tradingPairSymbol.BaseSymbol = baseSymbol
			tradingPairSymbol.QuoteSymbol = quoteSymbol
			tradingPairSymbol.ID = uint64(1)
			exhID := v1common.ExchangeID(v1common.ExchangeName(entry.ExchangeID).String())
			centralExh, ok := v1common.SupportedExchanges[exhID]
			if !ok {
				return errors.Errorf("exchange %s not supported", exhID)
			}
			exchangeInfo, err := centralExh.GetLiveExchangeInfos([]common.TradingPairSymbols{tradingPairSymbol})
			if err != nil {
				return err
			}
			info := exchangeInfo[1]
			entry.MinNotional = info.MinNotional
			entry.AmountLimitMax = info.AmountLimit.Max
			entry.AmountLimitMin = info.AmountLimit.Min
			entry.AmountPrecision = uint64(info.Precision.Amount)
			entry.PricePrecision = uint64(info.Precision.Price)
			entry.PriceLimitMax = info.PriceLimit.Max
			entry.PriceLimitMin = info.PriceLimit.Min
		}
	}
	return nil
}

func (s *Server) createSettingChange(c *gin.Context) {
	var createSettingChange settingChange
	if err := c.ShouldBindJSON(&createSettingChange); err != nil {
		log.Printf("cannot bind data to create setting_change from request err=%s", err.Error())
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	var settingChangeRequest = common.SettingChange{
		ChangeList: []common.SettingChangeEntry{},
	}
	for i, o := range createSettingChange.ChangeList {
		obj, err := common.SettingChangeFromType(o.Type)
		if err != nil {
			msg := fmt.Sprintf("change type must set at %d\n", i)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
			return
		}
		if err = json.Unmarshal(o.Data, obj); err != nil {
			msg := fmt.Sprintf("decode error at %d, err=%s", i, err)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
			return
		}
		if err = s.validateChangeEntry(obj, o.Type); err != nil {
			msg := fmt.Sprintf("validate error at %d, err=%s", i, err)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
		}
		settingChangeRequest.ChangeList = append(settingChangeRequest.ChangeList, common.SettingChangeEntry{
			Type: o.Type,
			Data: obj,
		})
	}

	if err := s.fillLiveInfoSettingChange(&settingChangeRequest); err != nil {
		msg := fmt.Sprintf("fill live info error=%s", err)
		log.Println(msg)
		httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
	}

	id, err := s.storage.CreateSettingChange(settingChangeRequest)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	// test confirm
	err = s.storage.ConfirmSettingChange(id, false)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) getSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	result, err := s.storage.GetSettingChange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) getSettingChanges(c *gin.Context) {
	result, err := s.storage.GetSettingChanges()
	if err != nil {
		log.Printf("failed to get setting changes %v\n", err)
		httputil.ResponseFailure(c, httputil.WithError(err))
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
	err := s.storage.RejectSettingChange(input.ID)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}

func (s *Server) confirmSettingChange(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	err := s.storage.ConfirmSettingChange(input.ID, true)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
