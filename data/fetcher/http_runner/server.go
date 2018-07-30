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
func (hrs *HttpRunnerServer) register() {
	hrs.r.GET("/ping", pingHandler)

	hrs.r.GET("/otick", newTickerHandler(hrs.runner.oticker))
	hrs.r.GET("/atick", newTickerHandler(hrs.runner.aticker))
	hrs.r.GET("/rtick", newTickerHandler(hrs.runner.rticker))
	hrs.r.GET("/btick", newTickerHandler(hrs.runner.bticker))
	hrs.r.GET("/gtick", newTickerHandler(hrs.runner.globalDataTicker))
}

// Start creates the HTTP server if needed and starts it.
// The HTTP server is running in foreground.
// This function always return a non-nil error.
func (hrs *HttpRunnerServer) Start() error {
	if hrs.http == nil {
		hrs.http = &http.Server{
			Handler: hrs.r,
		}

		lis, err := net.Listen("tcp", hrs.host)
		if err != nil {
			return err
		}

		// if port is not provided, use a random one and set it back to runner.
		if hrs.runner.port == 0 {
			_, listenedPort, sErr := net.SplitHostPort(lis.Addr().String())
			if sErr != nil {
				return sErr
			}
			port, sErr := strconv.Atoi(listenedPort)
			if sErr != nil {
				return sErr
			}
			hrs.runner.port = port
		}

		hrs.notifyCh <- struct{}{}

		return hrs.http.Serve(lis)
	}
	return errors.New("server start already")
}

// Stop shutdowns the HTTP server and free the resources.
// It returns an error if the server is shutdown already.
func (hrs *HttpRunnerServer) Stop() error {
	if hrs.http != nil {
		err := hrs.http.Shutdown(nil)
		hrs.http = nil
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
