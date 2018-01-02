package main

import (
	"os"
	"strings"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/exchange"
	"github.com/KyberNetwork/reserve-data/exchange/binance"
	"github.com/KyberNetwork/reserve-data/exchange/bittrex"
	"github.com/KyberNetwork/reserve-data/signer"
)

type ExchangePool struct {
	Exchanges map[common.ExchangeID]interface{}
}

func NewSimulationExchangePool(
	addressConfig common.AddressConfig,
	signer *signer.FileSigner,
	bittrexStorage exchange.BittrexStorage) *ExchangePool {

	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			bit := exchange.NewBittrex(bittrex.NewSimulatedBittrexEndpoint(signer), bittrexStorage)
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				bit.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bit.ID()] = bit
		case "binance":
			bin := exchange.NewBinance(binance.NewSimulatedBinanceEndpoint(signer))
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				bin.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewDevExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			bit := exchange.NewBittrex(bittrex.NewDevBittrexEndpoint(signer), bittrexStorage)
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				bit.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			bit.UpdatePairsPrecision()
			exchanges[bit.ID()] = bit
		case "binance":
			bin := exchange.NewBinance(binance.NewDevBinanceEndpoint(signer))
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				bin.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			bin.UpdatePairsPrecision()
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewKovanExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			bit := exchange.NewBittrex(bittrex.NewKovanBittrexEndpoint(signer), bittrexStorage)
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				bit.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bit.ID()] = bit
		case "binance":
			bin := exchange.NewBinance(binance.NewKovanBinanceEndpoint(signer))
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				bin.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bin.ID()] = bin
		}
	}
	return &ExchangePool{exchanges}
}

func NewMainnetExchangePool(addressConfig common.AddressConfig, signer *signer.FileSigner, bittrexStorage exchange.BittrexStorage) *ExchangePool {
	exchanges := map[common.ExchangeID]interface{}{}
	params := os.Getenv("KYBER_EXCHANGES")
	exparams := strings.Split(params, ",")
	for _, exparam := range exparams {
		switch exparam {
		case "bittrex":
			bit := exchange.NewBittrex(bittrex.NewRealBittrexEndpoint(signer), bittrexStorage)
			for tokenID, addr := range addressConfig.Exchanges["bittrex"] {
				bit.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bit.ID()] = bit
		case "binance":
			bin := exchange.NewBinance(binance.NewRealBinanceEndpoint(signer))
			for tokenID, addr := range addressConfig.Exchanges["binance"] {
				bin.UpdateDepositAddress(common.MustGetToken(tokenID), addr)
			}
			exchanges[bin.ID()] = bin
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
