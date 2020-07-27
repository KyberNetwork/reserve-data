package configuration

import (
	"github.com/urfave/cli"
	"go.uber.org/zap"
)

const (
	secretConfigFileFlag = "secret-file"
	configFileFlag       = "config"
)

// NewSecretConfigCliFlag returns the cli flag to configure secret config file flag.
func NewSecretConfigCliFlag() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   secretConfigFileFlag,
			Usage:  "location of secret config file",
			EnvVar: "SECRET_FILE",
			Value:  "secret.json",
		},
		cli.StringFlag{
			Name:   configFileFlag,
			Usage:  "location of config file",
			EnvVar: "CONFIG_FILE",
			Value:  "config.json",
		},
	}
}

// NewConfigFilesFromContext returns the configured secret config file location.
func NewConfigFilesFromContext(c *cli.Context) (string, string) {
	configFile := c.GlobalString(configFileFlag)
	secretConfigFile := c.GlobalString(secretConfigFileFlag)
	l := zap.S()
	l.Infow("using secret config file", "file", secretConfigFile)
	return configFile, secretConfigFile
}
