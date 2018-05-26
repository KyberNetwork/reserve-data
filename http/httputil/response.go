package httputil

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ResponseOption is the additional data to include in response.
type ResponseOption func(h gin.H)

// WithMultipleFields includes the extra data to the response.
func WithMultipleFields(extra gin.H) ResponseOption {
	return func(h gin.H) {
		for k, v := range extra {
			h[k] = v
		}
	}
}

// WithField includes the given field in the response.
func WithField(k string, v interface{}) ResponseOption {
	return WithMultipleFields(gin.H{k: v})
}

// WithData includes the given data to the response.
func WithData(data interface{}) ResponseOption {
	return WithField("data", data)
}

// WithReason includes the given reason to the response. It is
// intended to use for operation failed response.
func WithReason(reason string) ResponseOption {
	return WithField("reason", reason)
}

// WithError includes the error as a failure reason. It is
// usually used along with ResponseFailure.
// If err is nil, empty reason will be used.
func WithError(err error) ResponseOption {
	if err == nil {
		return WithField("reason", "")
	}
	return WithField("reason", err.Error())
}

// ResponseSuccess responses the request with 200 status code and a
// success message.
func ResponseSuccess(c *gin.Context, options ...ResponseOption) {
	h := gin.H{
		"success": true,
	}

	for _, option := range options {
		option(h)
	}

	c.JSON(
		http.StatusOK,
		h,
	)
}

// ResponseFailure responses the request with 200 status code and a
// failure message.
func ResponseFailure(c *gin.Context, options ...ResponseOption) {
	h := gin.H{
		"success": false,
	}

	for _, option := range options {
		option(h)
	}

	c.JSON(
		http.StatusOK,
		h,
	)
}
