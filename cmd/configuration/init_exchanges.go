package configuration

import (
	"errors"
	"fmt"
	"sync"

	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/common"
	bbc "github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/common/config"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/settings"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func asyncUpdateDepositAddress(ex common.Exchange, tokenID, addr string, wait *sync.WaitGroup, setting *settings.Settings) {
	defer wait.Done()
	l := zap.S()
	token, err := setting.GetTokenByID(tokenID)
	if err != nil {
		l.Panicf("Can't get token %s. Error: %+v", tokenID, err)
	}
	if err := ex.UpdateDepositAddress(token, addr); err != nil {
		l.Warnw("Cant not update deposit address for token, this will need to be manually update", "tokenID", tokenID, "exchangeID", ex.ID(), "err", err)
	}
}

func NewExchangePool(ac config.AppConfig, blockchain *bbc.BaseBlockchain, setting *settings.Settings) (*ExchangePool, error) {
	var (
		stableEx, bin, hExchange common.Exchange
		err                      error
	)
	exchanges := map[common.ExchangeID]interface{}{}
	exparams := settings.RunningExchanges()
	l := zap.S()
	for _, exparam := range exparams {
		switch exparam {
		case "stable_exchange":
			stableEx, err = exchange.NewStableEx(
				setting,
			)
			if err != nil {
				return nil, fmt.Errorf("can not create exchange stable_exchange: (%+v)", err)
			}
			exchanges[stableEx.ID()] = stableEx
		case "binance":
			binanceSigner, err := binance.NewSigner(ac.BinanceKey, ac.BinanceSecret)
			if err != nil {
				l.Panicw("failed to init binance signer", "err", err)
			}
			client := binance.NewBinanceClient(*binanceSigner, binance.NewEndpoints(ac.ExchangeEndpoints.Binance.URL))
			storage, err := binance.NewBoltStorage(ac.BinanceDB)
			if err != nil {
				return nil, fmt.Errorf("can not create Binance storage: (%+v)", err)
			}
			bin, err = exchange.NewBinance(client, storage, setting)
			if err != nil {
				return nil, fmt.Errorf("can not create exchange Binance: (%+v)", err)
			}
			addrs, err := setting.GetDepositAddresses(settings.Binance)
			if err != nil {
				l.Infof("Can't get Binance Deposit Addresses from Storage (%+v)", err)
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go asyncUpdateDepositAddress(bin, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			if err = bin.UpdatePairsPrecision(); err != nil {
				return nil, fmt.Errorf("can not Update Binance Pairs Precision: (%+v)", err)
			}
			exchanges[bin.ID()] = bin
		case "huobi":
			huobiSigner, err := huobi.NewSigner(ac.HuobiKey, ac.HuobiSecret)
			if err != nil {
				l.Panicw("failed to init houbi signer", "err", err)
			}
			client := huobi.NewHuobiClient(*huobiSigner, huobi.NewEndpoints(ac.ExchangeEndpoints.Houbi.URL))
			storage, err := huobi.NewBoltStorage(ac.HuobiDB)
			if err != nil {
				return nil, fmt.Errorf("can not create Huobi storage: (%+v)", err)
			}
			intermediatorSigner := bbc.NewEthereumSigner(ac.HoubiKeystorePath, ac.HuobiPassphrase)
			intermediatorNonce := nonce.NewTimeWindow(intermediatorSigner.GetAddress(), 10000)
			hExchange, err = exchange.NewHuobi(client, blockchain,
				intermediatorSigner, intermediatorNonce, storage, setting)
			if err != nil {
				return nil, fmt.Errorf("can not create exchange Huobi: (%+v)", err)
			}
			addrs, err := setting.GetDepositAddresses(settings.Huobi)
			if err != nil {
				l.Infof("Can't get Huobi Deposit Addresses from Storage (%+v)", err)
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go asyncUpdateDepositAddress(hExchange, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			if err = hExchange.UpdatePairsPrecision(); err != nil {
				return nil, fmt.Errorf("can not Update Huobi Pairs Precision: (%+v)", err)
			}
			exchanges[hExchange.ID()] = hExchange
		}
	}
	return &ExchangePool{exchanges}, nil
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
