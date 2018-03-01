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

func (self *Checker) FetchAccStatusFromBlockChain(timepoint uint64) {
	pendings, err := self.storage.GetPendingActivitiesByDest(self.accaddr.String())
	if err != nil {
		log.Printf("Getting pending activites failed: %s\n", err)
		return
	}
	if len(pendings) < 1 {
		log.Println("There is no pending activites to the account")
		return
	}
	for pending := range pendings {
		if pending.MiningStatus == "mined" && pending.ExchangeStatus == "" {
			if self.GetAccountBalance() > 100 {
				go self.DepositToHuobi()
				//wait for the
				<-self.runner.com_channel
			}
		}
	}

}

func (self *checker) DepositToHuobi() {

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
