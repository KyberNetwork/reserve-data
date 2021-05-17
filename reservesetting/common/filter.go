package common

import (
	"github.com/KyberNetwork/reserve-data/common/ethutil"
)

// AssetsHaveAddress filter return assets have address
func AssetsHaveAddress(assets []Asset) []Asset {
	var result []Asset
	for _, asset := range assets {
		if !ethutil.IsZeroAddress(asset.Address) {
			result = append(result, asset)
		}
	}
	return result
}

// AssetsHaveSetRate filter return asset have setrate
func AssetsHaveSetRate(assets []Asset) []Asset {
	var result []Asset
	for _, asset := range assets {
		if asset.SetRate != SetRateNotSet {
			result = append(result, asset)
		}
	}
	return result
}
