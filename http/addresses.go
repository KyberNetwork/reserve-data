package http

import (
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/lib/rtypes"
)

const (
	pricingOPAddressName       = "pricing_operator"
	depositOPAddressName       = "deposit_operator"
	intermediateOPAddressName  = "intermediate_operator"
	wrapper                    = "wrapper"
	network                    = "network"
	reserveAddress             = "reserve"
	rateQueryHelperAddressName = "rate_query_helper"
)

// GetAddresses get address config from core
func (s *Server) GetAddresses(c *gin.Context) {
	var (
		addresses = make(map[string]ethereum.Address)
	)
	addresses[pricingOPAddressName] = s.blockchain.GetPricingOPAddress()
	addresses[depositOPAddressName] = s.blockchain.GetDepositOPAddress()
	if h := common.SupportedExchanges[rtypes.Huobi]; h != nil {
		addresses[intermediateOPAddressName] = s.blockchain.GetIntermediatorOPAddress()
	}
	addresses[wrapper] = s.rcf.ContractAddresses.Wrapper
	addresses[network] = s.rcf.ContractAddresses.Proxy
	addresses[reserveAddress] = s.rcf.ContractAddresses.Reserve
	addresses[rateQueryHelperAddressName] = s.rcf.ContractAddresses.RateQueryHelper
	result := common.NewAddressesResponse(addresses, s.rcf.Nodes.Main)
	httputil.ResponseSuccess(c, httputil.WithData(result))
}
