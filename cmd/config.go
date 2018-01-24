package main

import (
	"github.com/KyberNetwork/reserve-data/blockchain"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/http"
	"github.com/KyberNetwork/reserve-data/metric"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Config struct {
	ActivityStorage core.ActivityStorage
	DataStorage     data.Storage
	FetcherStorage  fetcher.Storage
	MetricStorage   metric.MetricStorage

	FetcherRunner    fetcher.FetcherRunner
	FetcherExchanges []fetcher.Exchange
	Exchanges        []common.Exchange
	BlockchainSigner blockchain.Signer

	EnableAuthentication bool
	AuthEngine           http.Authentication

	EthereumEndpoint string

	SupportedTokens []common.Token

	WrapperAddress   ethereum.Address
	PricingAddress   ethereum.Address
	ReserveAddress   ethereum.Address
	FeeBurnerAddress ethereum.Address
	NetworkAddress   ethereum.Address
}

func (self *Config) MapTokens() map[string]common.Token {
	result := map[string]common.Token{}
	for _, t := range self.SupportedTokens {
		result[t.ID] = t
	}
	return result
}

func (self *Config) GetAddress() *common.Addresses {
	exchanges := map[string]common.TokenAddresses{}
	for _, e := range self.Exchanges {
		exchanges[string(e.ID())] = e.TokenAddresses()
	}
	tokens := map[string]ethereum.Address{}
	for _, t := range self.SupportedTokens {
		tokens[t.ID] = ethereum.HexToAddress(t.Address)
	}

	return &common.Addresses{
		Tokens: tokens,
		Exchanges: exchanges,
		WrapperAddress: self.WrapperAddress,
		PricingAddress: self.PricingAddress,
		ReserveAddress: self.ReserveAddress,
		FeeBurnerAddress: self.FeeBurnerAddress,
		NetworkAddress: self.NetworkAddress,
	}
}
