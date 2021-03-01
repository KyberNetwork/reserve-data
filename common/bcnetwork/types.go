package bcnetwork

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
)

var (
	selectedNetwork string

	networks = map[string]NetworkConfig{
		"ETH": {
			Network: "ETH",
			BootstrapAsset: BootstrapAsset{
				Symbol:   "ETH",
				Name:     "Ethereum",
				Address:  common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				Decimals: 18,
			},
		},
		"BSC": {
			Network: "BSC",
			BootstrapAsset: BootstrapAsset{
				Symbol:   "BNB",
				Name:     "BNB",
				Address:  common.HexToAddress("0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE"),
				Decimals: 18,
			},
		},
	}
)

// SetActiveNetwork set network config to use, this should be call on very soon as service get start
// need to set this correct before init DB(as init DB fill seed asset), and other service check for main asset be ETH/BNB
func SetActiveNetwork(name string) error {
	if _, ok := networks[name]; !ok {
		return fmt.Errorf("network not support %s", name)
	}
	selectedNetwork = name
	return nil
}

// GetPreConfig get network config by name,
func GetPreConfig() NetworkConfig {
	if selectedNetwork == "" { // default to "ETH" so test file will continue to work without change
		selectedNetwork = "ETH"
	}
	return networks[selectedNetwork]
}

// BootstrapAsset keep info to setup boot asset(like ETH)
type BootstrapAsset struct {
	Symbol   string         `json:"symbol"`
	Address  common.Address `json:"address"`
	Name     string         `json:"name"`
	Decimals uint64         `json:"decimal"`
}

// NetworkConfig store config to switch to BSC, this will use to seed say ETH/BNB on the network
type NetworkConfig struct {
	Network        string
	BootstrapAsset BootstrapAsset
}
