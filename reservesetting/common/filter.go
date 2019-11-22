package common

// AssetsHaveAddress filter return assets have address
func AssetsHaveAddress(assets []Asset) []Asset {
	var result []Asset
	for _, asset := range assets {
		if !IsZeroAddress(asset.Address) {
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
