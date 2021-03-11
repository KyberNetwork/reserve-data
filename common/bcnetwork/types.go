package bcnetwork

import (
	"github.com/ethereum/go-ethereum/common"
)

var (
	selectedNetwork NetworkConfig
)

// SetActiveNetwork set network config to use, this should be call on very soon as service get start
// need to set this correct before init DB(as init DB fill seed asset), and other service check for main asset be ETH/BNB
func SetActiveNetwork(nc NetworkConfig) {
	selectedNetwork = nc
}

// GetPreConfig get network config by name,
func GetPreConfig() NetworkConfig {
	if selectedNetwork.Network == "" { // this make sure if no network config set, use ETH as default
		return NetworkConfig{
			Network:     "ETH",
			MaxGasPrice: 1000,
			BootstrapAsset: BootstrapAsset{
				Symbol:   "ETH",
				Name:     "Ethereum",
				Address:  common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				Decimals: 18,
			},
		}
	}
	return selectedNetwork
}

// BootstrapAsset keep info to setup boot asset(like ETH)
type BootstrapAsset struct {
	Symbol   string         `json:"symbol"`
	Address  common.Address `json:"address"`
	Name     string         `json:"name"`
	Decimals uint64         `json:"decimals"`
}

// NetworkConfig store config to switch to BSC, this will use to seed say ETH/BNB on the network
type NetworkConfig struct {
	Network        string         `json:"network"`
	MaxGasPrice    float64        `json:"max_gas_price"` // fallback value, in case there no network contract
	BootstrapAsset BootstrapAsset `json:"bootstrap_asset"`
}
