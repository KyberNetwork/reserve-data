package http_runner

import (
	"errors"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http/httputil"
	raven "github.com/getsentry/raven-go"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"
)

// maxTimeSpot is the default time point to return in case the
// timestamp parameter in request is omit or malformed.
const maxTimeSpot uint64 = math.MaxUint64

// HttpRunnerServer is the HTTP ticker server.
type HttpRunnerServer struct {
	runner *HttpRunner
	host   string
	r      *gin.Engine
	http   *http.Server

	// notifyCh is notified when the HTTP server is ready.
	notifyCh chan struct{}
}

// getTimePoint returns the timepoint from query parameter.
// If no timestamp parameter is supplied, or it is invalid, returns the default one.
func getTimePoint(c *gin.Context) uint64 {
	timestamp := c.DefaultQuery("timestamp", "")
	timepoint, err := strconv.ParseUint(timestamp, 10, 64)
	if err != nil {
		log.Printf("Interpreted timestamp(%s) to default - %d\n", timestamp, maxTimeSpot)
		return maxTimeSpot
	}
	log.Printf("Interpreted timestamp(%s) to %d\n", timestamp, timepoint)
	return timepoint
}

// newTickerHandler creates a new HTTP handler for given channel.
func newTickerHandler(ch chan time.Time) gin.HandlerFunc {
	return func(c *gin.Context) {
		timepoint := getTimePoint(c)
		ch <- common.TimepointToTime(timepoint)
		httputil.ResponseSuccess(c)
	}
}

// pingHandler always returns to client a success status.
func pingHandler(c *gin.Context) {
	httputil.ResponseSuccess(c)
}

// register setups the gin.Engine instance by registers HTTP handlers.
func (self *HttpRunnerServer) register() {
	self.r.GET("/ping", pingHandler)

	self.r.GET("/otick", newTickerHandler(self.runner.oticker))
	self.r.GET("/atick", newTickerHandler(self.runner.aticker))
	self.r.GET("/rtick", newTickerHandler(self.runner.rticker))
	self.r.GET("/btick", newTickerHandler(self.runner.bticker))
	self.r.GET("/gtick", newTickerHandler(self.runner.globalDataTicker))
}

// Start creates the HTTP server if needed and starts it.
// The HTTP server is running in foreground.
// This function always return a non-nil error.
func (self *HttpRunnerServer) Start() error {
	if self.http == nil {
		self.http = &http.Server{
			Handler: self.r,
		}

		lis, err := net.Listen("tcp", self.host)
		if err != nil {
			return err
		}

		// if port is not provided, use a random one and set it back to runner.
		if self.runner.port == 0 {
			_, listenedPort, sErr := net.SplitHostPort(lis.Addr().String())
			if sErr != nil {
				return sErr
			}
			port, sErr := strconv.Atoi(listenedPort)
			if sErr != nil {
				return sErr
			}
			self.runner.port = port
		}

		self.notifyCh <- struct{}{}

		return self.http.Serve(lis)
	}
	return errors.New("server start already")
}

// Stop shutdowns the HTTP server and free the resources.
// It returns an error if the server is shutdown already.
func (self *HttpRunnerServer) Stop() error {
	if self.http != nil {
		err := self.http.Shutdown(nil)
		self.http = nil
		return err
	}
	return errors.New("server stop already")
}

// NewHttpRunnerServer creates a new instance of HttpRunnerServer.
func NewHttpRunnerServer(runner *HttpRunner, host string) *HttpRunnerServer {
	r := gin.Default()
	r.Use(sentry.Recovery(raven.DefaultClient, false))
	server := &HttpRunnerServer{
		runner:   runner,
		host:     host,
		r:        r,
		http:     nil,
		notifyCh: make(chan struct{}, 1),
	}
	server.register()
	return server
}
