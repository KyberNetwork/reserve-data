package configuration

import (
	"os"
	"strings"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/exchange/huobi"
	"github.com/KyberNetwork/reserve-data/signer"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func AsyncUpdateDepositAddress(ex common.Exchange, tokenID, addr string, wait *sync.WaitGroup) {
	defer wait.Done()
	ex.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
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
	signer *signer.FileSigner,
	bittrexStorage exchange.BittrexStorage, kyberENV string) *ExchangePool {

	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			endpoint := bittrex.NewBittrexEndpoint(signer, getBittrexInterface(kyberENV))
			bit := exchange.NewBittrex(addressConfig.Exchanges["bittrex"], feeConfig.Exchanges["bittrex"], endpoint, bittrexStorage)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bit, tokenID, addr, &wait)
			}
			wait.Wait()
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			endpoint := binance.NewBinanceEndpoint(signer, getBinanceInterface(kyberENV))
			bin := exchange.NewBinance(addressConfig.Exchanges["binance"], feeConfig.Exchanges["binance"], endpoint)
			wait := sync.WaitGroup{}
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				wait.Add(1)
				go AsyncUpdateDepositAddress(bin, tokenID, addr, &wait)
			}
			wait.Wait()
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		case "huobi":
			endpoint := huobi.NewHuobiEndpoint(signer, getHuobiInterface(kyberENV))
			huobi := exchange.NewHuobi(addressConfig.Exchanges["huobi"], feeConfig.Exchanges["huobi"], endpoint)
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
