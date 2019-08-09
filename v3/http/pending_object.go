package http

import (
	"log"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
	"github.com/gin-gonic/gin"
)

func (s *Server) getPendingObjects(objectType common.PendingObjectType) func(c *gin.Context) {
	return func(c *gin.Context) {
		result, err := s.storage.GetPendingObjects(objectType)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		httputil.ResponseSuccess(c, httputil.WithData(result))
	}
}

func (s *Server) getPendingObject(objectType common.PendingObjectType) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input struct {
			ID uint64 `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&input); err != nil {
			log.Println(err)
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		result, err := s.storage.GetPendingObject(input.ID, objectType)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		httputil.ResponseSuccess(c, httputil.WithData(result))
	}
}

func (s *Server) getCreateAsset(c *gin.Context) {
	var input struct {
		ID uint64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&input); err != nil {
		log.Println(err)
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	result, err := s.storage.GetPendingObject(input.ID, common.PendingTypeCreateAsset)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithData(result))
}

func (s *Server) rejectPendingObject(objectType common.PendingObjectType) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input struct {
			ID uint64 `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&input); err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		err := s.storage.RejectPendingObject(input.ID, objectType)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		httputil.ResponseSuccess(c)
	}
}

func (s *Server) confirmPendingObject(objectType common.PendingObjectType) func(c *gin.Context) {
	return func(c *gin.Context) {
		var input struct {
			ID uint64 `uri:"id" binding:"required"`
		}
		if err := c.ShouldBindUri(&input); err != nil {
			log.Println(err)
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		err := s.storage.ConfirmPendingObject(input.ID, objectType)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		httputil.ResponseSuccess(c)
	}
}
