package configuration

import (
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/common/config"
)

func mustGetDepositAddress(ac config.AppConfig) map[common.ExchangeID]common.ExchangeAddresses {
	res := make(map[common.ExchangeID]common.ExchangeAddresses)
	for k, v := range ac.DepositAddressesSet {
		res[common.ExchangeID(k)] = common.ExchangeAddresses(v)
	}
	return res
}
