package data

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type TestStorage struct{}

func (ts TestStorage) CurrentPriceVersion() (common.Version, error) {
	return common.Version(10), nil
}

func (ts TestStorage) GetAllPrices(version common.Version) (map[common.TokenPairID]common.OnePrice, error) {
	return map[common.TokenPairID]common.OnePrice{}, nil
}

func (ts TestStorage) GetOnePrice(pairID common.TokenPairID, version common.Version) (common.OnePrice, error) {
	return common.OnePrice{}, nil
}

func NewTestStorage() TestStorage {
	return TestStorage{}
}
