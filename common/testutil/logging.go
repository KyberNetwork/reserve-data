package testutil

import (
	"go.uber.org/zap"
)

// MustNewDevelopmentSugaredLogger returns a development sugared logger.
func MustNewDevelopmentSugaredLogger() *zap.SugaredLogger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.Sugar()
}
