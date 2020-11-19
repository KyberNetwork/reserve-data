package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/gin-contrib/httpsign"
	"github.com/gin-gonic/gin"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"

	"github.com/KyberNetwork/httpsign-utils/authenticator"
	"github.com/KyberNetwork/reserve-data/cmd/mode"
	"github.com/KyberNetwork/reserve-data/gateway/http"
	libapp "github.com/KyberNetwork/reserve-data/lib/app"
	"github.com/KyberNetwork/reserve-data/lib/httputil"
)

const (
	coreEndpointFlag    = "core-endpoint"
	settingEndpointFlag = "setting-endpoint"

	noAuthFlag = "no-auth"

	configFileFlag    = "config"
	defaultConfigFile = "config.json"
)

var (
	coreEndpointDefaultValue    = "http://127.0.0.1:8000"
	settingEndpointDefaultValue = "http://127.0.0.1:8002"
)

func main() {
	app := cli.NewApp()
	app.Name = "HTTP gateway for reserve core"
	app.Action = run
	app.Flags = append(app.Flags, cli.StringFlag{
		Name:   coreEndpointFlag,
		Usage:  "core endpoint url",
		EnvVar: "CORE_ENDPOINT",
		Value:  coreEndpointDefaultValue,
	},
		cli.StringFlag{
			Name:   settingEndpointFlag,
			Usage:  "setting endpoint url",
			EnvVar: "SETTING_ENDPOINT",
			Value:  settingEndpointDefaultValue,
		},
		cli.BoolFlag{
			Name:   noAuthFlag,
			Usage:  "no authenticate",
			EnvVar: "NO_AUTH",
		},
		cli.StringFlag{
			Name:   configFileFlag,
			Usage:  "path to config file",
			EnvVar: "CONFIG",
			Value:  defaultConfigFile,
		},
	)

	app.Flags = append(app.Flags, httputil.NewHTTPCliFlags(httputil.GatewayPort)...)
	app.Flags = append(app.Flags, mode.NewCliFlag())

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

type keyConfig struct {
	ReadRole      []authenticator.KeyPair `json:"read_role"`
	ManageRole    []authenticator.KeyPair `json:"manage_role"`
	RebalanceRole []authenticator.KeyPair `json:"rebalance_role"`
}

func configFromFile(path string) (keyConfig, error) {
	var c keyConfig
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(f, &c); err != nil {
		return c, err
	}
	return c, nil
}

func run(c *cli.Context) error {
	var (
		keyPairs []authenticator.KeyPair
		err      error
		auth     *httpsign.Authenticator
		perm     gin.HandlerFunc
	)
	logger, err := libapp.NewLogger(c)
	if err != nil {
		return err
	}
	defer libapp.NewFlusher(logger)()
	zap.ReplaceGlobals(logger)
	if err := validation.Validate(c.String(coreEndpointFlag),
		validation.Required,
		is.URL); err != nil {
		return errors.Wrapf(err, "app names API URL error: %s", c.String(coreEndpointFlag))
	}

	if err := validation.Validate(c.String(settingEndpointFlag),
		validation.Required,
		is.URL); err != nil {
		return errors.Wrapf(err, "app names API URL error: %s", c.String(settingEndpointFlag))
	}

	noAuth := c.Bool(noAuthFlag)
	if !noAuth {
		config, err := configFromFile(c.String(configFileFlag))
		if err != nil {
			return errors.Wrap(err, "cannot read config file")
		}
		for _, kp := range config.ReadRole {
			if err := validation.Validate(kp.AccessKeyID, validation.Required); err != nil {
				return errors.Wrap(err, "read access key error")
			}

			if err := validation.Validate(kp.SecretAccessKey, validation.Required); err != nil {
				return errors.Wrap(err, "read secret key error")
			}
			keyPairs = append(keyPairs, kp)
		}
		for _, kp := range config.ManageRole {
			if err := validation.Validate(kp.AccessKeyID, validation.Required); err != nil {
				return errors.Wrap(err, "manage access key error")
			}

			if err := validation.Validate(kp.SecretAccessKey, validation.Required); err != nil {
				return errors.Wrap(err, "manage secret key error")
			}
			keyPairs = append(keyPairs, kp)
		}
		for _, kp := range config.RebalanceRole {
			if err := validation.Validate(kp.AccessKeyID, validation.Required); err != nil {
				return errors.Wrap(err, "rebalance access key error")
			}

			if err := validation.Validate(kp.SecretAccessKey, validation.Required); err != nil {
				return errors.Wrap(err, "rebalance secret key error")
			}
			keyPairs = append(keyPairs, kp)
		}
		auth, err = authenticator.NewAuthenticator(keyPairs...)
		if err != nil {
			return errors.Wrap(err, "authentication object creation error")
		}
		perm, err = http.NewPermissioner(config.ReadRole, config.ManageRole, config.RebalanceRole)
		if err != nil {
			return errors.Wrap(err, "permission object creation error")
		}
	}

	svr, err := http.NewServer(httputil.NewHTTPAddressFromContext(c),
		auth,
		perm,
		noAuth,
		logger,
		http.WithCoreEndpoint(c.String(coreEndpointFlag)),
		http.WithSettingEndpoint(c.String(settingEndpointFlag)),
	)
	if err != nil {
		return errors.Wrap(err, "create new server error")
	}
	return svr.Start()
}
