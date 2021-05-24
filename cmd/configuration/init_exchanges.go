package configuration

import (
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

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
	binblockchain "github.com/KyberNetwork/reserve-data/exchange/blockchain"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	huobiStorage "github.com/KyberNetwork/reserve-data/exchange/huobi/storage"
	authhttp "github.com/KyberNetwork/reserve-data/lib/auth-http"
	rtypes "github.com/KyberNetwork/reserve-data/lib/rtypes"
	"github.com/KyberNetwork/reserve-data/reservesetting/storage"
)

type ExchangePool struct {
	Exchanges map[rtypes.ExchangeID]interface{}
	l         *zap.SugaredLogger
}

func updateTradingPairConf(
	assetStorage storage.Interface,
	ex common.Exchange, exchangeID rtypes.ExchangeID) {
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

func updateDepositAddress(assetStorage storage.Interface, exchanges map[rtypes.ExchangeID]interface{}) {
	var (
		bin *exchange.Binance
		hb  *exchange.Huobi
		ok  bool
	)
	l := zap.S()
	assets, err := assetStorage.GetTransferableAssets()
	if err != nil {
		l.Warnw("failed to get transferable assets", "err", err.Error())
		return
	}
	for _, asset := range assets {
		for _, ae := range asset.Exchanges {
			switch ae.ExchangeID {
			case rtypes.Binance, rtypes.Binance2:
				l.Infow("updating deposit address for asset", "asset_id", asset.ID,
					"exchange", ae.ExchangeID.String(), "symbol", ae.Symbol)
				if bin, ok = exchanges[ae.ExchangeID].(*exchange.Binance); !ok {
					l.Warnw("exchange does not exist", "exchange id", ae.ExchangeID)
				}
				depositAddress, ok := bin.Address(asset)
				if !ok {
					l.Warnw("failed to get deposit address for asset",
						"asset_id", asset.ID,
						"exchange", ae.ExchangeID.String(), "symbol", ae.Symbol, "err", err)
					continue
				}
				if err = assetStorage.UpdateDepositAddress(
					asset.ID,
					ae.ExchangeID,
					depositAddress); err != nil {
					l.Warnw("failed to update deposit address", "err", err)
				}
				l.Infow("updated deposit address", "address", depositAddress.Hex())
			case rtypes.Huobi:
				l.Infow("updating deposit address for asset", "asset_id", asset.ID,
					"exchange", ae.ExchangeID.String(),
					"symbol", ae.Symbol)
				if hb, ok = exchanges[ae.ExchangeID].(*exchange.Huobi); !ok {
					l.Warnw("exchange does not exist", "exchange id", ae.ExchangeID)
				}
				depositAddress, ok := hb.Address(asset)
				if !ok {
					l.Warnw("failed to get deposit address for asset",
						"asset_id", asset.ID,
						"exchange", ae.ExchangeID.String(),
						"symbol", ae.Symbol, "err", err)
					continue
				}
				err = assetStorage.UpdateDepositAddress(
					asset.ID,
					rtypes.Huobi,
					depositAddress)
				if err != nil {
					l.Warnw("assetStorage.UpdateDepositAddress", "err", err.Error())
					continue
				}
				l.Infow("updated deposit address", "address", depositAddress.Hex())
			}
		}
	}
}

func registerOP(bc *blockchaincommon.BaseBlockchain, op, file, pass string, chainID *big.Int) {
	txSigner := blockchaincommon.NewEthereumSigner(file, pass, chainID)
	nonceIns := nonce.NewTimeWindow(txSigner.GetAddress(), 10000)
	bc.MustRegisterOperator(op, blockchaincommon.NewOperator(txSigner, nonceIns))
}

func NewExchangePool(
	c *cli.Context,
	rcf common.RawConfig,
	blockchain *blockchaincommon.BaseBlockchain,
	dpl deployment.Deployment,
	bi binance.Interface,
	hi huobi.Interface,
	assetStorage storage.Interface,
	chainID *big.Int,
) (*ExchangePool, error) {
	exchanges := map[rtypes.ExchangeID]interface{}{}
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
	httpClient := &http.Client{Transport: exchange.NewTransportRateLimiter(&http.Client{Timeout: time.Second * 30})}
	marketDataBaseURL := strings.TrimSuffix(rcf.MarketDataBaseURL, "/")

	if rcf.BinanceIntermediateKeystore != "" && rcf.BinanceIntermediatePassphrase != "" {
		registerOP(blockchain, blockchaincommon.BinanceOP, rcf.BinanceIntermediateKeystore, rcf.BinanceIntermediatePassphrase, chainID)
		s.Debug("register binance op")
	}
	if rcf.IntermediatorKeystore != "" && rcf.IntermediatorPassphrase != "" {
		registerOP(blockchain, blockchaincommon.HuobiOP, rcf.IntermediatorKeystore, rcf.IntermediatorPassphrase, chainID)
		s.Debug("register huobi op")
	}
	for _, exparam := range enabledExchanges {
		switch exparam {
		case rtypes.Binance, rtypes.Binance2:
			accountID := rcf.BinanceAccountID
			binanceSigner := binance.NewSigner(rcf.BinanceKey, rcf.BinanceSecret)
			if exparam == rtypes.Binance2 {
				accountID = rcf.BinanceAccount2ID
				binanceSigner = binance.NewSigner(rcf.Binance2Key, rcf.Binance2Secret)
			}
			accountDataBaseURL := strings.TrimSuffix(rcf.AccountData.BaseURL, "/")
			be = binance.NewBinanceEndpoint(binanceSigner, bi, dpl, httpClient, exparam,
				marketDataBaseURL, accountDataBaseURL, accountID, authhttp.NewAuthHTTP(rcf.AccountData.AccessKey, rcf.AccountData.AccessSecret))
			binancestorage, err := binanceStorage.NewPostgresStorage(db)
			if err != nil {
				return nil, fmt.Errorf("cannot create Binance storage: (%s)", err.Error())
			}
			bc, err := binblockchain.NewBlockchain(blockchain)
			if err != nil {
				return nil, fmt.Errorf("cannot create binance blockchain: %s", err.Error())
			}
			bin, err = exchange.NewBinance(exparam, be, binancestorage, assetStorage, bc, rcf.OverrideTxPeriodSeconds)
			if err != nil {
				return nil, fmt.Errorf("cannot create exchange Binance: (%s)", err.Error())
			}
			exchanges[bin.ID()] = bin
		case rtypes.Huobi:
			huobiSigner := huobi.NewSigner(rcf.HoubiKey, rcf.HoubiSecret)
			he = huobi.NewHuobiEndpoint(huobiSigner, hi, httpClient, marketDataBaseURL)
			huobistorage, err := huobiStorage.NewPostgresStorage(db)
			if err != nil {
				return nil, fmt.Errorf("cannot create Binance storage: (%s)", err.Error())
			}
			hb, err = exchange.NewHuobi(he, blockchain, huobistorage, assetStorage)
			if err != nil {
				return nil, fmt.Errorf("cannot create exchange Huobi: (%s)", err.Error())
			}
			exchanges[hb.ID()] = hb
		}
	}

	go updateDepositAddress(assetStorage, exchanges)
	if bin, ok := exchanges[rtypes.Binance].(*exchange.Binance); ok {
		go updateTradingPairConf(assetStorage, bin, bin.ID())
	}
	if bin2, ok := exchanges[rtypes.Binance2].(*exchange.Binance); ok {
		go updateTradingPairConf(assetStorage, bin2, bin2.ID())
	}
	if hb != nil {
		go updateTradingPairConf(assetStorage, hb, rtypes.Huobi)
	}
	return &ExchangePool{
		Exchanges: exchanges,
		l:         s,
	}, nil
}

func (ep *ExchangePool) FetcherExchanges() ([]fetcher.Exchange, error) {
	var result []fetcher.Exchange
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
