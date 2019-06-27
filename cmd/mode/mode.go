package mode

import (
	"fmt"

	"github.com/urfave/cli"
)

const modeFlag = "mode"

// Mode is running mode of the application.
//go:generate stringer -type=Mode -linecomment
type Mode int

const (
	// ModeDevelopment is the mode to run in developer workstation.
	Development Mode = iota // develop
	// ModeProduction is the mode to run in production server environment.
	Production // production
)

// NewCliFlag returns new cli flag from context.
func NewCliFlag() cli.Flag {
	return cli.StringFlag{
		Name:   modeFlag,
		Usage:  "app running mode",
		EnvVar: "MODE",
		Value:  Development.String(),
	}
}

// NewModeFromContext returns the running mode from context.
func NewModeFromContext(c *cli.Context) (Mode, error) {
	modeStr := c.GlobalString(modeFlag)
	switch modeStr {
	case Development.String():
		return Development, nil
	case Production.String():
		return Production, nil
	}

	return 0, fmt.Errorf("unknown mode %s", modeStr)
}
