package http

import (
	"fmt"
	"log"
	"strings"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/gin-gonic/gin"
)

//KYCInfo return kyc info of an user
func (hs *HTTPServer) KYCInfo(c *gin.Context) {
	err := c.Request.ParseForm()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("Malformed request package: %s", err.Error())))
		return
	}
	postForm := c.Request.Form
	// Get KYC info
	email := postForm.Get("email")
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
	timestamps := strings.Split(timestampStr, "-")
	log.Printf("email: %s, addresses: %+v, timestamps: %+v", email, addresses, timestamps)
	httputil.ResponseSuccess(c)
}
