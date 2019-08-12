package configuration

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/mode"
)

type syncer interface {
	Sync() error
}

// NewFlusher creates a new syncer from given syncer that log a error message if failed to sync.
func NewFlusher(s syncer) func() {
	return func() {
		// ignore the error as the sync function will always fail in Linux
		// https://github.com/uber-go/zap/issues/370
		_ = s.Sync()
	}
}

// NewLogger creates a new logger instance.
// The type of logger instance will be different with different application running modes.
func NewLogger(mod mode.Mode) (*zap.Logger, error) {
	switch mod {
	case mode.Production:
		zapConfig := zap.NewProductionConfig()
		zapConfig.OutputPaths = []string{"stdout"}
		return zapConfig.Build()
	case mode.Development:
		zapConfig := zap.NewDevelopmentConfig()
		zapConfig.OutputPaths = []string{"stdout"}
		return zapConfig.Build()
	default:
		return nil, fmt.Errorf("invalid running mod: %q", mod)
	}
}

// NewSugaredLogger creates a new sugared logger and a flush function. The flush function should be
// called by consumer before quitting application.
// This function should be use most of the time unless
// the application requires extensive performance, in this case use NewLogger.
func NewSugaredLogger(mode mode.Mode) (*zap.SugaredLogger, func(), error) {
	logger, err := NewLogger(mode)
	if err != nil {
		return nil, nil, err
	}

	sugar := logger.Sugar()
	return sugar, NewFlusher(logger), nil
}
