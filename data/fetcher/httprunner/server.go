package httprunner

import (
	"context"
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

// Server is the HTTP ticker server.
type Server struct {
	runner *HTTPRunner
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
func (s *Server) register() {
	s.r.GET("/ping", pingHandler)

	s.r.GET("/otick", newTickerHandler(s.runner.oticker))
	s.r.GET("/atick", newTickerHandler(s.runner.aticker))
	s.r.GET("/rtick", newTickerHandler(s.runner.rticker))
	s.r.GET("/btick", newTickerHandler(s.runner.bticker))
	s.r.GET("/gtick", newTickerHandler(s.runner.globalDataTicker))
}

// Start creates the HTTP server if needed and starts it.
// The HTTP server is running in foreground.
// This function always return a non-nil error.
func (s *Server) Start() error {
	if s.http == nil {
		s.http = &http.Server{
			Handler: s.r,
		}

		lis, err := net.Listen("tcp", s.host)
		if err != nil {
			return err
		}

		// if port is not provided, use a random one and set it back to runner.
		if s.runner.port == 0 {
			_, listenedPort, sErr := net.SplitHostPort(lis.Addr().String())
			if sErr != nil {
				return sErr
			}
			port, sErr := strconv.Atoi(listenedPort)
			if sErr != nil {
				return sErr
			}
			s.runner.port = port
		}

		s.notifyCh <- struct{}{}

		return s.http.Serve(lis)
	}
	return errors.New("server start already")
}

// Stop shutdowns the HTTP server and free the resources.
// It returns an error if the server is shutdown already.
func (s *Server) Stop() error {
	if s.http != nil {
		err := s.http.Shutdown(context.Background())
		s.http = nil
		return err
	}
	return errors.New("server stop already")
}

// NewServer creates a new instance of HttpRunnerServer.
func NewServer(runner *HTTPRunner, host string) *Server {
	r := gin.Default()
	r.Use(sentry.Recovery(raven.DefaultClient, false))
	server := &Server{
		runner:   runner,
		host:     host,
		r:        r,
		http:     nil,
		notifyCh: make(chan struct{}, 1),
	}
	server.register()
	return server
}
