package exchange

import (
	"math/big"
	"sync"

	"github.com/KyberNetwork/reserve-data/common"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Bitfinex struct {
	interf      BitfinexInterface
	pairs       []common.TokenPair
	addresses   map[string]ethereum.Address
	databusType string
}

func (self *Bitfinex) MarshalText() (text []byte, err error) {
	return []byte(self.ID()), nil
}

func (self *Bitfinex) Address(token common.Token) (ethereum.Address, bool) {
	addr, supported := self.addresses[token.ID]
	return addr, supported
}

func (self *Bitfinex) UpdateAllDepositAddresses(address string) {
	for k, _ := range self.addresses {
		self.addresses[k] = ethereum.HexToAddress(address)
	}
}

func (self *Bitfinex) UpdateDepositAddress(token common.Token, address string) {
	self.addresses[token.ID] = ethereum.HexToAddress(address)
}

func (self *Bitfinex) UpdateFetcherDatabusType(databusType string) {
	self.databusType = databusType
}

func (self *Bitfinex) ID() common.ExchangeID {
	return common.ExchangeID("bitfinex")
}

func (self *Bitfinex) DatabusType() string {
	return self.databusType
}

func (self *Bitfinex) TokenPairs() []common.TokenPair {
	return self.pairs
}

func (self *Bitfinex) Name() string {
	return "bitfinex"
}

func (self *Bitfinex) Trade(tradeType string, base common.Token, quote common.Token, rate float64, amount float64, timepoint uint64) (done float64, remaining float64, finished bool, err error) {
	return self.interf.Trade(tradeType, base, quote, rate, amount, timepoint)
}

func (self *Bitfinex) Withdraw(token common.Token, amount *big.Int, address ethereum.Address, timepoint uint64) error {
	return self.interf.Withdraw(token, amount, address, timepoint)
}

func (self *Bitfinex) FetchPriceData(timepoint uint64) (map[common.TokenPairID]common.ExchangePrice, error) {
	wait := sync.WaitGroup{}
	data := sync.Map{}
	pairs := self.pairs
	for _, pair := range pairs {
		wait.Add(1)
		go self.interf.FetchOnePairData(&wait, pair, &data, timepoint)
	}
	wait.Wait()
	result := map[common.TokenPairID]common.ExchangePrice{}
	data.Range(func(key, value interface{}) bool {
		result[key.(common.TokenPairID)] = value.(common.ExchangePrice)
		return true
	})
	return result, nil
}

func (self *Bitfinex) FetchPriceDataUsingSocket() (map[common.TokenPairID]common.ExchangePrice, error) {
	// TODO: add support for socket later
	panic("Socket has not implemented yet")
}

func (self *Bitfinex) FetchEBalanceData(timepoint uint64) (common.EBalanceEntry, error) {
	result := common.EBalanceEntry{}
	return result, nil
}

func NewBitfinex(interf BitfinexInterface) *Bitfinex {
	return &Bitfinex{
		interf,
		[]common.TokenPair{
			common.MustCreateTokenPair("OMG", "ETH"),
			common.MustCreateTokenPair("EOS", "ETH"),
		},
		map[string]ethereum.Address{
			"ETH": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"OMG": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"DGD": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"CVC": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"MCO": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"GNT": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"ADX": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"EOS": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"PAY": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"BAT": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
			"KNC": ethereum.HexToAddress("0x5b082bc7928e1fd5b757426fe7225cc7a6a75c55"),
		},
		"http",
	}
}
