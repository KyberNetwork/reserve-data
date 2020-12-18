package http

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	libhttputil "github.com/KyberNetwork/reserve-data/lib/httputil"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/httpsign"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Server is HTTP server of gateway service.
type Server struct {
	r    *gin.Engine
	addr string
}

func newReverseProxyMW(target string, noAuth bool) (gin.HandlerFunc, error) {
	parsedURL, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	proxy := httputil.NewSingleHostReverseProxy(parsedURL)

	return func(c *gin.Context) {
		if !noAuth {
			c.Request.Header.Del("UserKeyID")
			signedHeader, err := NewSignatureHeader(c.Request)
			if err != nil {
				libhttputil.ResponseFailure(c, http.StatusBadRequest, err)
				return
			}
			c.Request.Header.Add("UserKeyID", signedHeader.KeyID)
		}
		proxy.ServeHTTP(c.Writer, c.Request)
	}, nil
}

// NewServer creates new instance of gateway HTTP server.
// TODO: add logger
func NewServer(addr string,
	auth *httpsign.Authenticator,
	perm gin.HandlerFunc,
	noAuth bool,
	logger *zap.Logger,
	options ...Option,
) (*Server, error) {
	r := gin.Default()
	r.Use(libhttputil.MiddlewareHandler) // TODO: remove this as we have already have zap logger?
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AddAllowHeaders("Digest", "Authorization", "Signature", "Nonce")
	corsConfig.MaxAge = 5 * time.Minute
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(cors.New(corsConfig))
	if !noAuth {
		r.Use(auth.Authenticated())
		r.Use(perm)
	}

	server := Server{
		addr: addr,
		r:    r,
	}

	for _, opt := range options {
		if err := opt(&server); err != nil {
			return nil, err
		}
	}
	return &server, nil
}

// Start runs the HTTP gateway server.
func (svr *Server) Start() error {
	return svr.r.Run(svr.addr)
}
