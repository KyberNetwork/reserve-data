package http

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
)

type HTTPServer struct {
	app  Huobi
	host string
	r    *gin.Engine
}

func IsIntime(nonce string) bool {
	serverTime := common.GetTimepoint()
	log.Printf("Server time: %d, None: %s", serverTime, nonce)
	nonceInt, err := strconv.ParseInt(nonce, 10, 64)
	if err != nil {
		log.Printf("IsIntime returns false, err: %v", err)
		return false
	}
	difference := nonceInt - int64(serverTime)
	if difference < -30000 || difference > 30000 {
		log.Printf("IsIntime returns false, nonce: %d, serverTime: %d, difference: %d", nonceInt, int64(serverTime), difference)
		return false
	}
	return true
}

func (self *HTTPServer) PendingIntermediateTxs(c *gin.Context) {
	data, err := self.app.PendingIntermediateTxs()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}

}

func (self *HTTPServer) Run() {
	if self.app != nil {
		self.r.GET("/pending_intermediate_tx", self.PendingIntermediateTxs)
	}

	self.r.Run(self.host)
}

func NewHuobiHTTPServer(app Huobi) *HTTPServer {
	huobihost := fmt.Sprintf(":12221")
	r := gin.Default()
	sentryCli, err := raven.NewWithTags(
		"https://bf15053001464a5195a81bc41b644751:eff41ac715114b20b940010208271b13@sentry.io/228067",
		map[string]string{
			"env": "huobi",
		},
	)
	if err != nil {
		panic(err)
	}
	r.Use(sentry.Recovery(
		sentryCli,
		false,
	))
	corsConfig := cors.DefaultConfig()
	corsConfig.AddAllowHeaders("signed")
	corsConfig.AllowAllOrigins = true
	corsConfig.MaxAge = 5 * time.Minute
	r.Use(cors.New(corsConfig))

	return &HTTPServer{
		app, huobihost, r,
	}
}
