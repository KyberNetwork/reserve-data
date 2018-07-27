package configuration

import (
	"errors"
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
	token, err := setting.GetTokenByID(tokenID)
	if err != nil {
		log.Panicf("ERROR: Can't get token %s. Error: %s", tokenID, err.Error())
	}
	if err := ex.UpdateDepositAddress(token, addr); err != nil {
		log.Printf("WARNING: Cant not update deposit address for token %s on exchange %s (%s), this will need to be manually update", tokenID, ex.ID(), err.Error())
	}
}

func getBittrexInterface(kyberENV string) bittrex.Interface {
	envInterface, ok := BittrexInterfaces[kyberENV]
	if !ok {
		envInterface = BittrexInterfaces[common.DevMode]
	}
	return envInterface
}

func getBinanceInterface(kyberENV string) binance.Interface {
	envInterface, ok := BinanceInterfaces[kyberENV]
	if !ok {
		envInterface = BinanceInterfaces[common.DevMode]
	}
	return envInterface
}

func getHuobiInterface(kyberENV string) huobi.Interface {
	envInterface, ok := HuobiInterfaces[kyberENV]
	if !ok {
		envInterface = HuobiInterfaces[common.DevMode]
	}
	return envInterface
}

func NewExchangePool(
	settingPaths SettingPaths,
	blockchain *blockchain.BaseBlockchain,
	kyberENV string, setting *settings.Settings) (*ExchangePool, error) {
	exchanges := map[common.ExchangeID]interface{}{}
	exparams := settings.RunningExchanges()
	for _, exparam := range exparams {
		switch exparam {
		case "stable_exchange":
			stableEx, err := exchange.NewStableEx(
				setting,
			)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange stable_exchange: (%s)", err.Error())
			}
			exchanges[stableEx.ID()] = stableEx
		case "bittrex":
			bittrexSigner := bittrex.NewSignerFromFile(settingPaths.secretPath)
			endpoint := bittrex.NewBittrexEndpoint(bittrexSigner, getBittrexInterface(kyberENV))
			bittrexStorage, err := bittrex.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "bittrex.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Bittrex storage: (%s)", err.Error())
			}
			bit, err := exchange.NewBittrex(
				endpoint,
				bittrexStorage,
				setting)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange Bittrex: (%s)", err.Error())
			}
			addrs, err := setting.GetDepositAddresses(settings.Bittrex)
			if err != nil {
				log.Printf("INFO: Can't get Bittrex Deposit Addresses from Storage (%s).", err.Error())
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			if err = bit.UpdatePairsPrecision(); err != nil {
				return nil, fmt.Errorf("Can not Update Bittrex Pairs Precision : (%s)", err.Error())
			}
			exchanges[bit.ID()] = bit
		case "binance":
			binanceSigner := binance.NewSignerFromFile(settingPaths.secretPath)
			endpoint := binance.NewBinanceEndpoint(binanceSigner, getBinanceInterface(kyberENV))
			storage, err := binance.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "binance.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Binance storage: (%s)", err.Error())
			}
			bin, err := exchange.NewBinance(
				endpoint,
				storage,
				setting)
			if err != nil {
				return nil, fmt.Errorf("Can not create exchange Binance: (%s)", err.Error())
			}
			addrs, err := setting.GetDepositAddresses(settings.Binance)
			if err != nil {
				log.Printf("INFO: Can't get Binance Deposit Addresses from Storage (%s)", err.Error())
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			if err = bin.UpdatePairsPrecision(); err != nil {
				return nil, fmt.Errorf("Can not Update Binance Pairs Precision: (%s)", err.Error())
			}
			exchanges[bin.ID()] = bin
		case "huobi":
			huobiSigner := huobi.NewSignerFromFile(settingPaths.secretPath)
			endpoint := huobi.NewHuobiEndpoint(huobiSigner, getHuobiInterface(kyberENV))
			storage, err := huobi.NewBoltStorage(filepath.Join(common.CmdDirLocation(), "huobi.db"))
			if err != nil {
				return nil, fmt.Errorf("Can not create Huobi storage: (%s)", err.Error())
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
				return nil, fmt.Errorf("Can not create exchange Huobi: (%s)", err.Error())
			}
			addrs, err := setting.GetDepositAddresses(settings.Huobi)
			if err != nil {
				log.Printf("INFO: Can't get Huobi Deposit Addresses from Storage (%s)", err.Error())
				addrs = make(common.ExchangeAddresses)
			}
			wait := sync.WaitGroup{}
			for tokenID, addr := range addrs {
				wait.Add(1)
				go AsyncUpdateDepositAddress(huobi, tokenID, addr.Hex(), &wait, setting)
			}
			wait.Wait()
			if err = huobi.UpdatePairsPrecision(); err != nil {
				return nil, fmt.Errorf("Can not Update Huobi Pairs Precision: (%s)", err.Error())
			}
			exchanges[huobi.ID()] = huobi
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
