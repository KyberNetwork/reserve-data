package configuration

import (
	"errors"
	"fmt"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	blockchaincommon "github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	binanceStorage "github.com/KyberNetwork/reserve-data/exchange/binance/storage"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	huobiStorage "github.com/KyberNetwork/reserve-data/exchange/huobi/storage"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
	l         *zap.SugaredLogger
}

func updateTradingPairConf(
	assetStorage storage.Interface,
	ex common.Exchange, exchangeID uint64) {
	l := zap.S()
	pairs, err := assetStorage.GetTradingPairs(exchangeID)
	if err != nil {
		l.Warnw("failed to get trading pairs", "exchange_id", exchangeID, "err", err.Error())
		return
	}
	exInfo, err := ex.GetLiveExchangeInfos(pairs)
	if err != nil {
		l.Warnw("failed to get pair configuration for binance", "err", err.Error())
		return
	}

	for _, pair := range pairs {
		pairConf, ok := exInfo[pair.ID]
		if !ok {
			l.Warnw("no configuration found for trading pair", "pair_id", pair.ID)
			return
		}

		var (
			pricePrecision  = uint64(pairConf.Precision.Price)
			amountPrecision = uint64(pairConf.Precision.Amount)
			amountLimitMin  = pairConf.AmountLimit.Min
			amountLimitMax  = pairConf.AmountLimit.Max
			priceLimitMin   = pairConf.PriceLimit.Min
			priceLimitMax   = pairConf.PriceLimit.Max
			minNotional     = pairConf.MinNotional
		)

		l.Infof("updating pair configuration id=%d exchange_id=%d", pair.ID, exchangeID)
		err = assetStorage.UpdateTradingPair(pair.ID, storage.UpdateTradingPairOpts{
			PricePrecision:  &pricePrecision,
			AmountPrecision: &amountPrecision,
			AmountLimitMin:  &amountLimitMin,
			AmountLimitMax:  &amountLimitMax,
			PriceLimitMin:   &priceLimitMin,
			PriceLimitMax:   &priceLimitMax,
			MinNotional:     &minNotional,
		})
		if err != nil {
			l.Warn("failed to update trading pair", "pair_id", pair.ID, "exchange_id", exchangeID, "err", err.Error())
			return
		}
	}
}

func updateDepositAddress(assetStorage storage.Interface, be exchange.BinanceInterface, he exchange.HuobiInterface) {
	l := zap.S()
	assets, err := assetStorage.GetTransferableAssets()
	if err != nil {
		l.Warnw("failed to get transferable assets", "err", err.Error())
		return
	}
	for _, asset := range assets {
		for _, ae := range asset.Exchanges {
			switch ae.ExchangeID {
			case uint64(common.Binance), uint64(common.Binance2):
				l.Warnw("updating deposit address for asset", "asset_id", asset.ID,
					"exchange", common.Binance.String(), "symbol", ae.Symbol)
				if be == nil {
					l.Warnw("abort updating deposit address due binance exchange disabled")
					continue
				}
				depositAddress, err := be.GetDepositAddress(ae.Symbol)
				if err != nil {
					l.Warnw("failed to get deposit address for asset",
						"asset_id", asset.ID,
						"exchange", common.Binance.String(), "symbol", ae.Symbol, "err", err.Error())
					continue
				}
				err = assetStorage.UpdateDepositAddress(
					asset.ID,
					ae.ExchangeID,
					ethereum.HexToAddress(depositAddress.Address))
				if err != nil {
					l.Warnw("assetStorage.UpdateDepositAddress", "err", err.Error())
					continue
				}
			case uint64(common.Huobi):
				l.Warnw("updating deposit address for asset",
					"asset_id", asset.ID,
					"exchange", common.Huobi.String(),
					"symbol", ae.Symbol)
				if he == nil {
					l.Warnw("abort updating deposit address due huobi exchange disabled")
					continue
				}
				depositAddress, err := he.GetDepositAddress(ae.Symbol)
				if err != nil {
					l.Warnw("failed to get deposit address for asset",
						"asset_id", asset.ID,
						"exchange", common.Huobi.String(),
						"symbol", ae.Symbol, "err", err)
					continue
				}
				if len(depositAddress.Data) != 0 {
					err = assetStorage.UpdateDepositAddress(
						asset.ID,
						uint64(common.Huobi),
						ethereum.HexToAddress(depositAddress.Data[0].Address))
					if err != nil {
						l.Warnw("assetStorage.UpdateDepositAddress", "err", err.Error())
						continue
					}
				}
			}
		}
	}
}

func NewExchangePool(
	c *cli.Context,
	rcf common.RawConfig,
	blockchain *blockchaincommon.BaseBlockchain,
	dpl deployment.Deployment,
	bi binance.Interface,
	hi huobi.Interface,
	assetStorage storage.Interface,
) (*ExchangePool, error) {
	exchanges := map[common.ExchangeID]interface{}{}
	var (
		be      exchange.BinanceInterface
		he      exchange.HuobiInterface
		bin, hb common.Exchange
		s       = zap.S()
	)

	enabledExchanges, err := NewExchangesFromContext(c)
	if err != nil {
		return nil, err
	}

	db, err := NewDBFromContext(c)
	if err != nil {
		return nil, fmt.Errorf("can not init postgres storage: (%s)", err.Error())
	}

	for _, exparam := range enabledExchanges {
		switch exparam {
		case common.Binance, common.Binance2:
			binanceSigner := binance.NewSigner(rcf.BinanceKey, rcf.BinanceSecret)
			if exparam == common.Binance2 {
				binanceSigner = binance.NewSigner(rcf.Binance2Key, rcf.Binance2Secret)
			}
			be = binance.NewBinanceEndpoint(binanceSigner, bi, dpl, exparam)
			binancestorage, err := binanceStorage.NewPostgresStorage(db)
			if err != nil {
				return nil, fmt.Errorf("can not create Binance storage: (%s)", err.Error())
			}
			bin, err = exchange.NewBinance(
				exparam,
				be,
				binancestorage,
				assetStorage)
			if err != nil {
				return nil, fmt.Errorf("can not create exchange Binance: (%s)", err.Error())
			}
			exchanges[bin.ID()] = bin
		case common.Huobi:
			huobiSigner := huobi.NewSigner(rcf.HoubiKey, rcf.HoubiSecret)
			he = huobi.NewHuobiEndpoint(huobiSigner, hi)
			huobistorage, err := huobiStorage.NewPostgresStorage(db)
			if err != nil {
				return nil, fmt.Errorf("can not create Binance storage: (%s)", err.Error())
			}
			intermediatorSigner := blockchaincommon.NewEthereumSigner(rcf.IntermediatorKeystore, rcf.IntermediatorPassphrase)
			intermediatorNonce := nonce.NewTimeWindow(intermediatorSigner.GetAddress(), 10000)
			hb, err = exchange.NewHuobi(
				he,
				blockchain,
				intermediatorSigner,
				intermediatorNonce,
				huobistorage,
				assetStorage,
			)
			if err != nil {
				return nil, fmt.Errorf("can not create exchange Huobi: (%s)", err.Error())
			}
			exchanges[hb.ID()] = hb
		}
	}

	go updateDepositAddress(assetStorage, be, he)
	if bin != nil {
		go updateTradingPairConf(assetStorage, bin, uint64(common.Binance))
	}
	if hb != nil {
		go updateTradingPairConf(assetStorage, hb, uint64(common.Huobi))
	}
	return &ExchangePool{
		Exchanges: exchanges,
		l:         s,
	}, nil
}

func (ep *ExchangePool) FetcherExchanges() ([]fetcher.Exchange, error) {
	result := []fetcher.Exchange{}
	for _, ex := range ep.Exchanges {
		fcEx, ok := ex.(fetcher.Exchange)
		if !ok {
			return result, errors.New("ExchangePool cannot be asserted  to fetcher exchange")
		}
		result = append(result, fcEx)
	}
	return result, nil
}

func (ep *ExchangePool) CoreExchanges() ([]common.Exchange, error) {
	result := []common.Exchange{}
	for _, ex := range ep.Exchanges {
		cmEx, ok := ex.(common.Exchange)
		if !ok {
			return result, errors.New("ExchangePool cannot be asserted to core Exchange")
		}
		result = append(result, cmEx)
	}
	return result, nil
}
