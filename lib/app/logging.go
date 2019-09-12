package app

import (
	"fmt"

	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/reserve-data/cmd/mode"
)

const (
	modeFlag = "mode"
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
func NewLogger(c *cli.Context) (*zap.Logger, error) {
	modeConfig := c.GlobalString(modeFlag)
	switch modeConfig {
	case mode.Production.String():
		return zap.NewProduction()
	case mode.Development.String():
		return zap.NewDevelopment()
	default:
		return nil, fmt.Errorf("invalid running mode: %q", modeConfig)
	}
}

// NewSugaredLogger creates a new sugared logger and a flush function. The flush function should be
// called by consumer before quitting application.
// This function should be use most of the time unless
// the application requires extensive performance, in this case use NewLogger.
func NewSugaredLogger(c *cli.Context) (*zap.SugaredLogger, func(), error) {
	logger, err := NewLogger(c)
	if err != nil {
		return nil, nil, err
	}

	sugar := logger.Sugar()
	return sugar, NewFlusher(logger), nil
}