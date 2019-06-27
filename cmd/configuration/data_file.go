package configuration

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
)

const (
	dataFileFlag        = "data-file"
	settingDataFileFlag = "setting-data-file"
)

// NewDataFileCliFlags returns all cli flags to use to configure data files location.
func NewDataFileCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:   dataFileFlag,
			Usage:  "location of main boltdb data file",
			EnvVar: "DATA_FILE",
		},
		cli.StringFlag{
			Name:   settingDataFileFlag,
			Usage:  "location of setting data file",
			EnvVar: "SETTING_DATA_FILE",
		},
	}
}

// NewDataFileFromContext returns the configured main data file from cli context.
func NewDataFileFromContext(c *cli.Context) (string, error) {
	var dataFile = c.GlobalString(dataFileFlag)
	if len(dataFile) == 0 {
		dpl, err := deployment.NewDeploymentFromContext(c)
		if err != nil {
			return "", err
		}
		defaultDataFile := filepath.Join(common.CmdDirLocation(), fmt.Sprintf("%s.db", dpl))
		log.Printf("using default data file location %s", defaultDataFile)
		return defaultDataFile, nil
	}
	return dataFile, nil
}

// NewSettingDataFileFromContext returns the configured setting data file from cli context.
func NewSettingDataFileFromContext(c *cli.Context) (string, error) {
	var settingDataFile = c.GlobalString(settingDataFileFlag)

	if len(settingDataFile) == 0 {
		dpl, err := deployment.NewDeploymentFromContext(c)
		if err != nil {
			return "", err
		}
		defaultSettingDataFile := filepath.Join(common.CmdDirLocation(), fmt.Sprintf("%s_setting.db", dpl))
		log.Printf("using default setting data file location %s", defaultSettingDataFile)
		return defaultSettingDataFile, nil
	}
	return settingDataFile, nil
}
