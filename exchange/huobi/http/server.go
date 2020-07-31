package http

import (
	"log"
	"time"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	raven "github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
)

//Server for huobi which including
//app stand for huobi exchange instance in reserve data
//host is for api calling
//r for http engine
type Server struct {
	app  Huobi
	host string
	r    *gin.Engine
}

//PendingIntermediateTxs get pending transaction
func (s *Server) PendingIntermediateTxs(c *gin.Context) {
	data, err := s.app.PendingIntermediateTxs()
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithReason(err.Error()))
	} else {
		httputil.ResponseSuccess(c, httputil.WithData(data))
	}

}

//Run run http server for huobi
func (s *Server) Run() {
	if s.app != nil {
		s.r.GET("/pending_intermediate_tx", s.PendingIntermediateTxs)
	}

	if err := s.r.Run(s.host); err != nil {
		log.Fatalf("Http server run error: %s", err.Error())
	}
}

//NewHuobiHTTPServer return new http instance
func NewHuobiHTTPServer(app Huobi) *Server {
	huobihost := ":12221"
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
	corsConfig.MaxAge = time.Minute * 5
	r.Use(cors.New(corsConfig))

	return &Server{
		app, huobihost, r,
	}
}
