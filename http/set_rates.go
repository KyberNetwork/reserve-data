package http

import (
	"fmt"
	"math/big"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
	v3common "github.com/KyberNetwork/reserve-data/reservesetting/common"
)

// RateRequest is request for a rate
type RateRequest struct {
	AssetID rtypes.AssetID `json:"asset_id"`
	Buy     string         `json:"buy"`
	Sell    string         `json:"sell"`
	Mid     string         `json:"mid"`
	Msg     string         `json:"msg"`
	Trigger bool           `json:"trigger"`
}

// SetRateEntry is input for set rate request
type SetRateEntry struct {
	Block uint64        `json:"block"`
	Rates []RateRequest `json:"rates"`
}

func tokenExisted(tokenAddr ethereum.Address, assets []v3common.Asset) bool {
	for _, asset := range assets {
		if tokenAddr == asset.Address {
			return true
		}
	}
	return false
}

func (s *Server) checkDelistedTokens(assets []v3common.Asset, bigBuys, bigSells, bigAfpMid []*big.Int, triggers []bool) ([]v3common.Asset, []*big.Int, []*big.Int, []*big.Int, []bool, error) {
	listedToken := s.blockchain.ListedTokens()
	for _, tokenAddr := range listedToken {
		if !tokenExisted(tokenAddr, assets) {
			assets = append(assets, v3common.Asset{
				Address: tokenAddr,
			})
			bigBuys = append(bigBuys, big.NewInt(0))
			bigSells = append(bigSells, big.NewInt(0))
			bigAfpMid = append(bigAfpMid, big.NewInt(0))
			triggers = append(triggers, false)
		}
	}
	return assets, bigBuys, bigSells, bigAfpMid, triggers, nil
}

func (s *Server) cancelSetRate(c *gin.Context) {
	id, err := s.core.CancelSetRate()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

// SetRate is for setting token rate
func (s *Server) SetRate(c *gin.Context) {
	var (
		input     SetRateEntry
		assets    []v3common.Asset
		bigBuys   = []*big.Int{}
		bigSells  = []*big.Int{}
		bigAfpMid = []*big.Int{}
		triggers  []bool
		msgs      []string
	)
	if err := c.ShouldBindJSON(&input); err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	for _, rates := range input.Rates {
		asset, err := s.settingStorage.GetAsset(rates.AssetID)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("failed to get asset from storage: %s", err.Error())))
			return
		}
		assets = append(assets, asset)
	}
	for _, rate := range input.Rates {
		rbuy, ok := big.NewInt(0).SetString(rate.Buy, 10)
		if !ok {
			httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("cannot parse rate number buy: %s", rate.Buy)))
			return
		}
		bigBuys = append(bigBuys, rbuy)
		rSell, ok := big.NewInt(0).SetString(rate.Sell, 10)
		if !ok {
			httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("cannot parse rate number sell: %s", rate.Sell)))
			return
		}
		bigSells = append(bigSells, rSell)
		rMid, ok := big.NewInt(0).SetString(rate.Mid, 10)
		if !ok {
			httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("cannot parse rate number mid: %s", rate.Mid)))
			return
		}
		bigAfpMid = append(bigAfpMid, rMid)
		msgs = append(msgs, rate.Msg)
		triggers = append(triggers, rate.Trigger)
	}
	var err error
	assets, bigBuys, bigSells, bigAfpMid, triggers, err = s.checkDelistedTokens(assets, bigBuys, bigSells, bigAfpMid, triggers)
	if err != nil {
		s.l.Warnw("failed to check delisted token", "error", err)
	}
	id, err := s.core.SetRates(assets, bigBuys, bigSells, big.NewInt(int64(input.Block)), bigAfpMid, msgs, triggers)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(fmt.Errorf("failed to set rates: %s", err.Error())))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}
