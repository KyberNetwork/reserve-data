package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/http"
)

func main() {
	app := cli.NewApp()
	app.Name = "Reserve Core"
	app.Usage = "Kyber Reserve core component that helps manage reserves of tokens"
	app.Version = "0.11.0"
	app.Action = run

	app.Flags = configuration.NewCliFlags()

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func run(c *cli.Context) error {
	configuration.SetupLogging()

	dpl, err := deployment.NewDeploymentFromContext(c)
	if err != nil {
		return err
	}

	conf, err := configuration.NewConfigurationFromContext(c)
	if err != nil {
		return err
	}

	bc, err := configuration.CreateBlockchain(conf)
	if err != nil {
		log.Printf("Can not create blockchain: (%s)", err)
		return err
	}

	dryRun := configuration.NewDryRunFromContext(c)

	rData, rCore := configuration.CreateDataCore(conf, dpl, bc)
	if !dryRun {
		if dpl != deployment.Simulation {
			if err = rData.RunStorageController(); err != nil {
				log.Printf("failed to run storage controller err=%s", err.Error())
				return err
			}
		}
		if err = rData.Run(); err != nil {
			log.Printf("failed to run data service err=%s", err.Error())
			return err
		}
	}

	for _, ex := range conf.Exchanges {
		common.SupportedExchanges[ex.ID()] = ex
	}

	port := configuration.NewPortFromContext(c)
	servPortStr := fmt.Sprintf(":%d", port)
	server := http.NewHTTPServer(
		rData, rCore,
		conf.MetricStorage,
		servPortStr,
		conf.EnableAuthentication,
		conf.AuthEngine,
		dpl,
		bc,
		conf.Setting,
		conf.ContractAddresses,
	)

	if !dryRun {
		server.Run()
	} else {
		log.Printf("Dry run finished. All configs are corrected")
	}

	return err
}
