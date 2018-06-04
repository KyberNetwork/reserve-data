package configuration

import (
	"fmt"
	"log"
	"path/filepath"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/settings"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func AsyncUpdateDepositAddress(ex common.Exchange, tokenID, addr string, wait *sync.WaitGroup, setting *settings.Settings) {
	defer wait.Done()
	token, err := setting.GetInternalTokenByID(tokenID)
	if err != nil {
		log.Panicf("ERROR: Can't get internal token %s. Error: %s", tokenID, err)
	}
	ex.UpdateDepositAddress(token, addr)
}

func getBittrexInterface(kyberENV string) bittrex.Interface {
	envInterface, ok := BittrexInterfaces[kyberENV]
	if !ok {
		envInterface = BittrexInterfaces[common.DEV_MODE]
	}
	return envInterface
}

func getBinanceInterface(kyberENV string) binance.Interface {
	envInterface, ok := BinanceInterfaces[kyberENV]
	if !ok {
		envInterface = BinanceInterfaces[common.DEV_MODE]
	}
	return envInterface
}

func getHuobiInterface(kyberENV string) huobi.Interface {
	envInterface, ok := HuobiInterfaces[kyberENV]
	if !ok {
		envInterface = HuobiInterfaces[common.DEV_MODE]
	}
	return envInterface
}

func NewExchangePool(
	settingPaths SettingPaths,
	blockchain *blockchain.BaseBlockchain,
	kyberENV string, setting *settings.Settings) (*ExchangePool, error) {
	exchanges := map[common.ExchangeID]interface{}{}
	exparams := settings.RunningExchange()
	for _, exparam := range exparams {
		switch exparam {
		case "stable_exchange":
			stableEx, err := exchange.NewStableEx(
				setting,
			)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange stable_exchange: (%s)", err)
			}
			exchanges[stableEx.ID()] = stableEx
		case "bittrex":
			bittrexSigner := bittrex.NewSignerFromFile(settingPaths.secretPath)
			endpoint := bittrex.NewBittrexEndpoint(bittrexSigner, getBittrexInterface(kyberENV))
			bittrexStorage, err := bittrex.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "bittrex.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Bittrex storage: (%s)", err)
			}
			bit, err := exchange.NewBittrex(
				endpoint,
				bittrexStorage,
				setting)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange Bittrex: (%s)", err)
			}
			addrs, err := setting.GetDepositAddress(settings.Bittrex)
			if err != nil {
				log.Printf("INFO: Can't get Bittrex Deposit Addresses from Storage (%s), attempt to init it from Bittrex Endpoint", err)
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			err = bit.UpdatePairsPrecision()
			if err != nil {
				return nil, fmt.Errorf("Can not Update Bittrex Pairs Precision : (%s)", err)
			}
			exchanges[bit.ID()] = bit
		case "binance":
			binanceSigner := binance.NewSignerFromFile(settingPaths.secretPath)
			endpoint := binance.NewBinanceEndpoint(binanceSigner, getBinanceInterface(kyberENV))
			storage, err := huobi.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "binance.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Binance storage: (%s)", err)
			}
			bin, err := exchange.NewBinance(
				endpoint,
				storage,
				setting)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange Binance: (%s)", err)
			}
			addrs, err := setting.GetDepositAddress(settings.Binance)
			if err != nil {
				log.Printf("INFO: Can't get Binance Deposit Addresses from Storage (%s), attempt to init it from Binance Endpoint", err.Error())
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			err = bin.UpdatePairsPrecision()
			if err != nil {
				return nil, fmt.Errorf("Can not Update Binance Pairs Precision: (%s)", err)
			}
			exchanges[bin.ID()] = bin
		case "huobi":
			huobiSigner := huobi.NewSignerFromFile(settingPaths.secretPath)
			endpoint := huobi.NewHuobiEndpoint(huobiSigner, getHuobiInterface(kyberENV))
			storage, err := huobi.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "huobi.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Huobi storage: (%s)", err)
			}
			intermediatorSigner := HuobiIntermediatorSignerFromFile(settingPaths.secretPath)
			intermediatorNonce := nonce.NewTimeWindow(intermediatorSigner.GetAddress(), 10000)
			huobi, err := exchange.NewHuobi(
				endpoint,
				blockchain,
				intermediatorSigner,
				intermediatorNonce,
				storage,
				setting,
			)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange Huobi: (%s)", err)
			}
			addrs, err := setting.GetDepositAddress(settings.Huobi)
			if err != nil {
				log.Printf("INFO: Can't get Huobi Deposit Addresses from Storage (%s), attempt to init it from Huobi Endpoint", err.Error())
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(huobi, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			err = huobi.UpdatePairsPrecision()
			if err != nil {
				return nil, fmt.Errorf("Can not Update Huobi Pairs Precision: (%s)", err)
			}
			exchanges[huobi.ID()] = huobi
		}
	}
	return &ExchangePool{exchanges}, nil
}

func (self *ExchangePool) FetcherExchanges() []fetcher.Exchange {
	result := []fetcher.Exchange{}
	for _, ex := range self.Exchanges {
		result = append(result, ex.(fetcher.Exchange))
	}
	return result
}

func (self *ExchangePool) CoreExchanges() []common.Exchange {
	result := []common.Exchange{}
	for _, ex := range self.Exchanges {
		result = append(result, ex.(common.Exchange))
	}
	return result
}
