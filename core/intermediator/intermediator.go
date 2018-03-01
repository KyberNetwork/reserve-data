package intermediator

import (
	"log"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	ethereum "github.com/ethereum/go-ethereum/common"
)

type Intermediator struct {
	storage       fetcher.Storage
	runner        IntermediatorRunner
	currentStatus []common.ActivityRecord
	intaddr       ethereum.Address
}

func unchanged(pre map[common.ActivityID]common.ActivityRecord, post common.ActivityRecord) bool {
	act, found := pre[post.ID]
	if !found {
		return false
	}
	if post.ExchangeStatus != act.ExchangeStatus {
		return false
	}
	if post.MiningStatus != act.MiningStatus {
		return false
	}

	return true
}

func (self *Intermediator) GetAccountBalance() float64 {
	return (0.0)
}

func (self *Intermediator) CheckAccStatusFromBlockChain() {
	pendings, err := self.storage.GetPendingActivities()
	if err != nil {
		log.Printf("Intermediator: Getting pending activites failed: %s\n", err)
		return
	}
	if len(pendings) < 1 {
		log.Println("Intermediator: There is no pending activites to the account")
		return
	}
	//make a map based on the old pending activities, only concern the destination fit to intermediator's address.
	compare_map := make(map[common.ActivityID]common.ActivityRecord)
	for _, act := range self.currentStatus {
		if act.Destination == self.intaddr.String() {
			compare_map[act.ID] = act
		}
	}
	//loop through the pendings, only concern the action which statisfy the following conditions:
	// 1. is deposit.
	// 2. mining status is mined.
	// 3. exchange status is empty
	// 4. changed comapre to the last get pending activities.
	for _, pending := range pendings {
		if pending.Action == "deposit" && pending.MiningStatus == "mined" && pending.ExchangeStatus == "" && !unchanged(compare_map, pending) {
			token := pending.Params["token"]
			log.Printf("Intermediator: Found a status change at activity %v, which deposit token %v \n", pending.ID, token)
			if self.GetAccountBalance() > 100 {
				self.DepositToHuobi()
			}
		}
	}

}

func (self *Intermediator) DepositToHuobi() {

}

func (self *Intermediator) RunStatusChecker() {
	log.Printf("Intermediator: deposit_huobi:waiting for signal from status checker channel")
	t := <-self.runner.GetStatusTicker()
	log.Printf("Intermediator: deposit_huobi: got signal from status checker with timestamp %d", common.TimeToTimepoint(t))
	self.CheckAccStatusFromBlockChain()

}

func (self *Intermediator) Run() error {
	log.Printf("Intermediator: deposit_huobi: Account Status checker is running... \n")
	go self.RunStatusChecker()
	return nil
}

func NewIntermediator(storage fetcher.Storage, runner IntermediatorRunner, address ethereum.Address) *Intermediator {
	return &Intermediator{storage, runner, nil, address}
}
