package common

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

// AssetPriceFactor present for price factor from set side
type AssetPriceFactor struct {
	AssetID rtypes.AssetID `json:"id"`
	AfpMid  float64        `json:"afp_mid"`
	Spread  float64        `json:"spread"`
}

// AssetPriceFactorList is a list of price factor that set side will send to server
type AssetPriceFactorList []AssetPriceFactor

// Scan scan from DB result
func (p *AssetPriceFactorList) Scan(src interface{}) error {
	var source []byte
	switch v := src.(type) {
	case string:
		source = []byte(v)
	case []uint8:
		source = v
	default:
		return errors.New("incompatible type")
	}
	return json.Unmarshal(source, p)
}

// Value format to DB data
func (p AssetPriceFactorList) Value() (driver.Value, error) {
	data, err := json.Marshal(p)
	return data, err
}

// PriceFactorAtTime present for a row of set price factor in DB.
type PriceFactorAtTime struct {
	ID        uint64               `json:"id"`
	Timestamp uint64               `json:"timestamp"`
	Data      AssetPriceFactorList `json:"data"`
}

// AssetPriceFactorResponse is on element in asset price factor list output in getPriceFactors
type AssetPriceFactorResponse struct {
	Timestamp uint64  `json:"timestamp"`
	AfpMid    float64 `json:"afp_mid"`
	Spread    float64 `json:"spread"`
}

// AssetPriceFactorListResponse present for price factor list of an asset.
type AssetPriceFactorListResponse struct {
	AssetID rtypes.AssetID             `json:"id"`
	Data    []AssetPriceFactorResponse `json:"data"`
}

// PriceFactorResponse present for out getPriceFactors result.
type PriceFactorResponse struct {
	Timestamp  uint64
	ReturnTime uint64
	Data       []*AssetPriceFactorListResponse
}
