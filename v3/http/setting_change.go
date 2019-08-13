package http

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/KyberNetwork/reserve-data/v3/common"
)

type settingChangeEntry struct {
	Type common.ChangeType `json:"type"`
	Data json.RawMessage   `json:"data"`
}

type settingChange struct {
	ChangeList []settingChangeEntry `json:"change_list" binding:"required"`
}

func (s *Server) validateChangeEntry(e common.SettingChangeType) error {
	// TODO: validate each change, also fill LiveInfo for trading pair here.
	return nil
}

func (s *Server) createSettingChange(c *gin.Context) {
	var createSettingChange settingChange
	if err := c.ShouldBindJSON(&createSettingChange); err != nil {
		log.Printf("cannot bind data to create setting_change from request err=%s", err.Error())
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	var settingChangeRequest = common.SettingChange{
		ChangeList: []common.SettingChangeEntry{},
	}
	for i, o := range createSettingChange.ChangeList {
		obj, err := common.SettingChangeFromType(o.Type)
		if err != nil {
			msg := fmt.Sprintf("change type must set at %d\n", i)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
			return
		}
		if err = json.Unmarshal(o.Data, obj); err != nil {
			msg := fmt.Sprintf("decode error at %d, err=%s", i, err)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
			return
		}
		if err = s.validateChangeEntry(obj); err != nil {
			msg := fmt.Sprintf("validate error at %d, err=%s", i, err)
			log.Println(msg)
			httputil.ResponseFailure(c, httputil.WithError(err), httputil.WithReason(msg))
		}
		settingChangeRequest.ChangeList = append(settingChangeRequest.ChangeList, common.SettingChangeEntry{
			Type: o.Type,
			Data: obj,
		})
	}

	id, err := s.storage.CreateSettingChange(settingChangeRequest)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c, httputil.WithField("id", id))
}
