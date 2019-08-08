package http

import (
	"log"
	"strings"

	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

const (
	validateAddressTag = "isAddress"
)

func (s *Server) createChangeAssetAddress(c *gin.Context) {
	var createChangeAssetAddress common.CreateChangeAssetAddress
	if err := c.ShouldBindJSON(&createChangeAssetAddress); err != nil {
		log.Printf("cannot bind data to create change_asset_addresses from request err=%s", err.Error())
		if strings.Contains(err.Error(), validateAddressTag) {
			err = common.ErrInvalidAddress
		}
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}

	for _, changeAssetAddressEntry := range createChangeAssetAddress.Assets {
		if err := s.checkChangeAssetAddressParams(changeAssetAddressEntry); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
	}

	id, err := s.storage.CreatePendingObject(createChangeAssetAddress, common.PendingTypeChangeAssetAddr)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}

func (s *Server) checkChangeAssetAddressParams(changeAssetAddressEntry common.ChangeAssetAddressEntry) error {
	asset, err := s.storage.GetAsset(changeAssetAddressEntry.ID)
	if err != nil {
		return err
	}
	if asset.Address == ethereum.HexToAddress(changeAssetAddressEntry.Address) {
		return common.ErrAddressExists
	}
	return nil
}
