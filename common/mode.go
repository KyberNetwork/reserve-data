package common

import (
	"os"
)

const (
	// modeEnv is the name environment variable that set the running mode of core.
	// See below constants for list of available modes.
	modeEnv = "KYBER_ENV"

	// DevMode is default running mode. It is used for development.
	DevMode = "dev"
	// StagingMode is the running mode in staging environment.
	StagingMode = "staging"
	// ProductionMode is the running mode in production environment.
	ProductionMode = "production"
	// MainnetMode is the same as production mode, just for backward compatibility.
	MainnetMode = "mainnet"
	// KovanMode is the running mode for testing kovan network.
	KovanMode = "kovan"
	// RopstenMode is the running mode for testing ropsten network.
	RopstenMode = "ropsten"
	// SimulationMode is running mode in simulation.
	SimulationMode = "simulation"
	// AnalyticDevMode is running mode for analytic development.
	AnalyticDevMode = "analytic_dev"
)

var validModes = map[string]struct{}{
	DevMode:         {},
	StagingMode:     {},
	ProductionMode:  {},
	MainnetMode:     {},
	KovanMode:       {},
	RopstenMode:     {},
	SimulationMode:  {},
	AnalyticDevMode: {},
}

// RunningMode returns the current running mode of application.
func RunningMode() string {
	mode, ok := os.LookupEnv(modeEnv)
	if !ok {
		return DevMode
	}
	_, valid := validModes[mode]
	if !valid {
		return DevMode
	}
	return mode
}
