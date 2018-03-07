package intermediator

import (
	"log"
	"math/big"
	"strconv"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Intermediator struct {
	storage       fetcher.Storage
	runner        IntermediatorRunner
	currentStatus []common.ActivityRecord
	intaddr       ethereum.Address
	blockchain    Blockchain
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

func (self *Intermediator) FetchAccountBalanceFromBlockchain(timepoint uint64) (map[string]common.BalanceEntry, error) {
	return self.blockchain.FetchBalanceData(self.intaddr, nil, timepoint)
}

func (self *Intermediator) CheckAccStatusFromBlockChain(timepoint uint64) {
	log.Printf("Intermediator: Get pending....")
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

	log.Printf("Intermediator: check pending....")

	//loop through the pendings, only concern the action which statisfy the following conditions:
	// 1. is deposit.
	// 2. mining status is mined.
	// 3. exchange status is empty
	// 4. changed comapre to the last get pending activities.
	for _, pending := range pendings {
		log.Printf("Intermediator: action is %v", pending)
		log.Printf("Intermediator: exchange status is %v", pending.ExchangeStatus)
		log.Printf("Intermediator: mining status is %v", pending.MiningStatus)
		log.Printf("Intermediator: amount is: %v ", pending.Params["amount"])
		if pending.Action == "deposit" && pending.MiningStatus == "mined" && pending.ExchangeStatus == "" && !unchanged(compare_map, pending) {
			tokenID, ok1 := pending.Params["token"].(string)
			exchangeID, ok2 := pending.Params["exchange"].(string)
			sentAmountStr, _ := pending.Params["amount"].(string)
			log.Printf("Intermediator: sentAmountStr= %s", sentAmountStr)
			sentAmount, ok4 := strconv.ParseFloat(sentAmountStr, 64)
			if ok4 != nil {
				log.Println("Intermediator: Activity record is malformed, cannot read the exchange amount")
			}
			if (!ok1) || (!ok2) {
				log.Println("Intermediator: Activity record is malformed, cannot read the exchange/ token")
			}
			log.Printf("Intermediator: Found a status change at activity %v, which deposit token %v \n", pending.ID, tokenID)
			accBalance, err := self.FetchAccountBalanceFromBlockchain(timepoint)
			if err != nil {
				log.Printf("Intermediator: can not get account balance %v", err)
			}
			token, err := common.GetToken(tokenID)
			if err != nil {
				log.Printf("Intermediator: Token is not supported: %v", err)
			}
			exchange, err := common.GetExchange(exchangeID)
			if err != nil {
				log.Printf("Intermediator: Exchange is not supported: %v", err)
			}
			log.Printf("Intermediator: base sentAmount is %.5f, token decimal is %d", sentAmount, token.Decimal)
			log.Printf("Intermediator: Sent amount is %.5f , balance is %.5f", sentAmount, accBalance[tokenID].ToBalanceResponse(token.Decimal).Balance)

			if accBalance[tokenID].ToBalanceResponse(token.Decimal).Balance >= sentAmount {
				//get token and exchange object from IDs in the activity
				self.DepositToExchange(token, exchange, sentAmount)
			} else {
				log.Printf("Intermediator: Insufficient Amount ")
			}
		}
	}

	self.currentStatus = pendings
}

func (self *Intermediator) DepositToExchange(token common.Token, exchange common.Exchange, amount float64) {
	exchangeAddress, supported := exchange.Address(token)
	log.Println()
	if !supported {
		log.Printf("ERROR: Intermediator: Token %s is not supported on Exchange %v", token.ID, exchange.ID)
		return
	}
	FAmount := big.NewFloat(amount)
	FDecimal := (big.NewFloat(0)).SetInt(big.NewInt(token.Decimal))
	FAmount.Mul(FAmount, FDecimal)
	IAmount := big.NewInt(0)
	FAmount.Int(IAmount)
	var tx *types.Transaction
	var err error
	if token.ID == "ETH" {
		tx, err = self.blockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress)
	} else {
		tx, err = self.blockchain.SendTokenFromAccountToExchange(IAmount, exchangeAddress)
	}
	if err != nil {
		log.Printf("ERROR: Intermediator: Can not send transaction to exchange: %v", err)
		return
	}
	log.Printf("Intermediator: Transaction submitted. Tx is: \n %v ", tx)
}

func (self *Intermediator) RunStatusChecker() {
	for {
		log.Printf("Intermediator: waiting for signal from status checker channel")
		t := <-self.runner.GetStatusTicker()
		log.Printf("Intermediator: got signal from status checker with timestamp %d", common.TimeToTimepoint(t))
		/*
			accBalance, err := self.FetchAccountBalanceFromBlockchain(common.TimeToTimepoint(t))
			if err != nil {
				log.Printf("Intermediator: can not get account balance %v", err)
			}
			log.Printf("Intermediator: account balance %.10f", accBalance["ETH"].ToBalanceResponse(10).Balance)
		*/
		self.CheckAccStatusFromBlockChain(common.TimeToTimepoint(t))
	}
}

func (self *Intermediator) Run() error {
	log.Printf("Intermediator: deposit_huobi: Account Status checker is running... \n")
	//log.Printf("Intermediator: Blockchain client is: %v", self.blockchain.client )
	self.runner.Start()
	go self.RunStatusChecker()
	return nil
}

func NewIntermediator(storage fetcher.Storage, runner IntermediatorRunner, address ethereum.Address, blockchain Blockchain) *Intermediator {
	return &Intermediator{storage, runner, nil, address, blockchain}
}
