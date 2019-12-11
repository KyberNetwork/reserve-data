package common

import (
	"os"
)

const (
	// modeEnv is the name environment variable that set the running mode of core.
	// See below constants for list of available modes.
	modeEnv        = "KYBER_ENV"
	ProductionMode = "production"
	SimulationMode = "simulation"
)

var validModes = map[string]struct{}{
	ProductionMode: {},
	SimulationMode: {},
}

// RunningMode returns the current running mode of application.
func RunningMode() string {
	mode, ok := os.LookupEnv(modeEnv)
	if !ok {
		return ProductionMode
	}
	_, valid := validModes[mode]
	if !valid {
		return ProductionMode
	}
	return mode
}
