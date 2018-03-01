package deposithuobi

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
)

type Checker struct {
	storage fetcher.Storage
	runner  deposit_huobi_Runner
}

func (self *Checker) RunStatusChecker() {
	log.Printf("deposit_huobi:waiting for signal from status checker channel")
	t <- self.runner.GetStatusTicker()
	log.Printf("deposit_huobi: got signal from status checker with timestamp %d", common.TimeToTimepoint(t))
	statuses := self.AccountStatusCheck()

}

func (self *Checker) Run() error {
	log.Printf("deposit_huobi: Account Status checker is running... \n")
	go self.RunStatusChecker()
	return nil
}
