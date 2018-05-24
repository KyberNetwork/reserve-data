package configuration

import (
	"os"
	"strings"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/blockchain"
	"github.com/KyberNetwork/reserve-data/common/blockchain/nonce"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func AsyncUpdateDepositAddress(ex common.Exchange, tokenID, addr string, wait *sync.WaitGroup) {
	defer wait.Done()
	ex.UpdateDepositAddress(common.MustGetInternalToken(tokenID), addr)
}
func getBittrexInterface(kyberENV string) bittrex.Interface {
	envInterface, err := BittrexInterfaces[kyberENV]
	if !err {
		envInterface = BittrexInterfaces["dev"]
	}
	return envInterface
}

func getBinanceInterface(kyberENV string) binance.Interface {
	envInterface, err := BinanceInterfaces[kyberENV]
	if !err {
		envInterface = BinanceInterfaces["dev"]
	}
	return envInterface
}

func getHuobiInterface(kyberENV string) huobi.Interface {
	envInterface, err := HuobiInterfaces[kyberENV]
	if !err {
		envInterface = HuobiInterfaces["dev"]
	}
	return envInterface
}

func NewExchangePool(
	feeConfig common.ExchangeFeesConfig,
	addressConfig common.AddressConfig,
	settingPaths SettingPaths,
	blockchain *blockchain.BaseBlockchain,
	minDeposit common.ExchangesMinDepositConfig,
	kyberENV string) *ExchangePool {

	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "stable_exchange":
			stableEx := exchange.NewStableEx(
				addressConfig.Exchanges["stable_exchange"],
				feeConfig.Exchanges["stable_exchange"],
				minDeposit.Exchanges["stable_exchange"],
			)
			exchanges[stableEx.ID()] = stableEx
		case "bittrex":
			bittrexSigner := bittrex.NewSignerFromFile(settingPaths.secretPath)
			endpoint := bittrex.NewBittrexEndpoint(bittrexSigner, getBittrexInterface(kyberENV))
			bittrexStorage, err := bittrex.NewBoltStorage("/go/src/github.com/KyberNetwork/reserve-data/cmd/bittrex.db")
			if err != nil {
				panic(err)
			}
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], feeConfig.Exchanges["bittrex"], endpoint, bittrexStorage, minDeposit.Exchanges["bittrex"])
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			binanceSigner := binance.NewSignerFromFile(settingPaths.secretPath)
			endpoint := binance.NewBinanceEndpoint(binanceSigner, getBinanceInterface(kyberENV))
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], feeConfig.Exchanges["binance"], endpoint, minDeposit.Exchanges["binance"])
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		case "huobi":
			huobiSigner := huobi.NewSignerFromFile(settingPaths.secretPath)
			endpoint := huobi.NewHuobiEndpoint(huobiSigner, getHuobiInterface(kyberENV))
			storage, err := huobi.NewBoltStorage("/go/src/github.com/KyberNetwork/reserve-data/cmd/huobi.db")
			intermediatorSigner := HuobiIntermediatorSignerFromFile(settingPaths.secretPath)
			intermediatorNonce := nonce.NewTimeWindow(intermediatorSigner.GetAddress(), 10000)
			if err != nil {
				panic(err)
			}
			huobi := exchange.NewHuobi(
				addressConfig.Exchanges["huobi"],
				feeConfig.Exchanges["huobi"],
				endpoint,
				blockchain,
				intermediatorSigner,
				intermediatorNonce,
				storage,
				minDeposit.Exchanges["huobi"],
			)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["huobi"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(huobi, tokenID, addr, &wait)
			}
			wait.Wait()
			huobi.UpdatePairsPrecision()
			exchanges[huobi.ID()] = huobi
		}
	}
	return &ExchangePool{exchanges}
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
