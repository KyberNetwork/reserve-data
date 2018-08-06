package http

import (
	"log"
	"strconv"
	"strings"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/gin-gonic/gin"
)

//KYCInfo return kyc info of an user
func (hs *HTTPServer) KYCInfo(c *gin.Context) {
	postForm, ok := hs.Authenticated(c, []string{"user", "addresses", "timestamps"}, []Permission{UserDashboardPermission})
	if !ok {
		return
	}
	// Get KYC info
	email := postForm.Get("user")
	if email == "" {
		httputil.ResponseFailure(c, httputil.WithReason("email is empty"))
		return
	}
	addrStr := postForm.Get("addresses")
	if addrStr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("address list is empty"))
		return
	}
	timestampStr := postForm.Get("timestamps")
	if timestampStr == "" {
		httputil.ResponseFailure(c, httputil.WithReason("timestamps list is empty"))
		return
	}
	addresses := strings.Split(addrStr, "-")
	timestampstr := strings.Split(timestampStr, "-")
	// convert timestamp to uint64
	var timestamps []uint64
	for _, timestamp := range timestampstr {
		v, err := strconv.ParseUint(timestamp, 10, 64)
		if err != nil {
			httputil.ResponseFailure(c, httputil.WithError(err))
			return
		}
		timestamps = append(timestamps, v)
	}
	log.Printf("email: %s, addresses: %+v, timestamps: %+v", email, addresses, timestamps)
	err := hs.stat.UpdateUserKYCInfo(email, addresses, timestamps)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
