package intermediator

import (
	"errors"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/KyberNetwork/reserve-data/common"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type DepositAction struct {
	actRecord common.ActivityRecord
	tx        *types.Transaction
}

type Intermediator struct {
	storage              fetcher.Storage
	runner               IntermediatorRunner
	CurrentDepositStatus map[common.ActivityID]DepositAction
	intaddr              ethereum.Address
	blockchain           Blockchain
}

func unchanged(pre map[common.ActivityID]DepositAction, post common.ActivityRecord) bool {
	log.Printf("Intermediator: pending ID is %v", post.ID)
	_, found := pre[post.ID]
	if !found {
		return false
	}
	act := pre[post.ID].actRecord
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

func (self *Intermediator) updateCurrentDepositStatus(pendings []common.ActivityRecord) {
	newStatus := make(map[common.ActivityID]DepositAction)
	var newDepositAction DepositAction
	for _, act := range pendings {

		_, ok := self.CurrentDepositStatus[act.ID]
		if ok {
			newDepositAction = DepositAction{
				act,
				self.CurrentDepositStatus[act.ID].tx,
			}
		} else {
			newDepositAction = DepositAction{
				act,
				nil}
		}
		newStatus[act.ID] = newDepositAction
	}
	self.CurrentDepositStatus = newStatus
}

func (self *Intermediator) updateAction(ID common.ActivityID, action *common.ActivityRecord, tx *types.Transaction, exchangeID string) {
	log.Printf("Intermediator: old Action is : %v ", *action)
	oldID := common.ActivityID{
		action.ID.Timepoint,
		action.ID.EID,
	}
	idParts := strings.Split(action.ID.EID, "|")
	nuID := tx.Hash().Hex() + "|" + idParts[1] + "|" + idParts[2]
	nID := common.ActivityID{
		action.ID.Timepoint,
		nuID,
	}
	action.ID = nID
	action.Destination = exchangeID
	log.Printf("Intermediator: new action is : %v", *action)
	self.storage.UpdateActivity(oldID, *action)
	log.Printf("Intermediator: update action finished")
}

func (self *Intermediator) CheckAccStatusFromBlockChain(timepoint uint64) {
	//get pending
	pendings, err := self.storage.GetPendingActivities()
	if err != nil {
		log.Printf("Intermediator: Getting pending activites failed: %s\n", err)
		return
	}
	if len(pendings) < 1 {
		log.Println("Intermediator: There is no pending activites to the account")
		return
	}

	log.Printf("Intermediator: check pending....")

	//loop through the pendings, only concern the action which statisfy the following conditions:
	// 1. is deposit.
	// 2. mining status is mined.
	// 3. exchange status is empty
	// 4. changed comapre to the last get pending activities.
	for idx, pending := range pendings {
		/*log.Printf("Intermediator: action is %v", pending)
		log.Printf("Intermediator: exchange status is %v", pending.ExchangeStatus)
		log.Printf("Intermediator: mining status is %v", pending.MiningStatus)
		log.Printf("Intermediator: amount is: %v ", pending.Params["amount"])*/
		if pending.Action == "deposit" && pending.MiningStatus == "mined" && pending.ExchangeStatus == "" {
			if !unchanged(self.CurrentDepositStatus, pending) {
				//get token and exchange from the status
				tokenID, ok1 := pending.Params["token"].(string)
				exchangeID, ok2 := pending.Params["exchange"].(string)
				if (!ok1) || (!ok2) {
					log.Println("Intermediator: Activity record is malformed, cannot read the exchange/ token")
				}
				//get amount deposit
				sentAmountStr, _ := pending.Params["amount"].(string)
				sentAmount, ok4 := strconv.ParseFloat(sentAmountStr, 64)
				if ok4 != nil {
					log.Println("Intermediator: Activity record is malformed, cannot read the exchange amount")
				}

				log.Printf("Intermediator: Found a status change at activity %v, which deposit token %v \n", pending.ID, tokenID)

				//get account token
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
				log.Printf("Intermediator: Sent amount is %.5f , balance is %.5f", sentAmount, accBalance[tokenID].ToBalanceResponse(token.Decimal).Balance)

				if accBalance[tokenID].ToBalanceResponse(token.Decimal).Balance >= sentAmount {
					//get token and exchange object from IDs in the activity
					tx, err := self.DepositToExchange(token, exchange, sentAmount)
					if err != nil {
						log.Printf("Intermediator: cannot send transaction %v", err)
					} else {
						newDepositAction := DepositAction{
							pending,
							tx,
						}
						self.CurrentDepositStatus[pending.ID] = newDepositAction
					}
				} else {
					log.Printf("Intermediator: Insufficient Amount ")
				}
			} else {
				//check tx status of the pending action and update
				tx := self.CurrentDepositStatus[pending.ID].tx
				if tx == nil {
					log.Printf("Intermediator: There was no transfer from account to exchanged submitted ")
				}
				accBalance, err := self.FetchAccountBalanceFromBlockchain(timepoint)
				if err != nil {
					log.Printf("Intermediator: can not get account balance 2nd time %v", err)
				}

				status, blockNumber, err := self.blockchain.TxStatus(tx.Hash())
				if err != nil {
					log.Printf("Intermediator: can not get tx Status ")
				} else if (status == "mined") || (status == "failed") {
					log.Printf("Intermediator: tx status was % s at block no %d", status, blockNumber)
					log.Printf("\nIntermediator: before pending is : %v", pending)
					exchangeID, _ := pending.Params["exchange"].(string)
					self.updateAction(pending.ID, &pending, tx, exchangeID)
					log.Printf("\nIntermediator: after pending is : %v", pendings[0])
					pendings[idx] = pending
					self.CurrentDepositStatus[pending.ID] = DepositAction{
						pending,
						tx,
					}
				}
				tokenID, _ := pending.Params["token"].(string)
				token, _ := common.GetToken(tokenID)
				log.Printf("Intermediator: balance after is %.5f", accBalance[tokenID].ToBalanceResponse(token.Decimal).Balance)
			}

		}
	}
	self.updateCurrentDepositStatus(pendings)
	log.Printf("current deposit status is %v", self.CurrentDepositStatus)
}

func getBigIntFromFloat(amount float64, decimal int64) *big.Int {
	FAmount := big.NewFloat(amount)

	power := math.Pow10(int(decimal))

	FDecimal := (big.NewFloat(0)).SetFloat64(power)
	FAmount.Mul(FAmount, FDecimal)
	IAmount := big.NewInt(0)
	FAmount.Int(IAmount)
	return IAmount
}

func (self *Intermediator) DepositToExchange(token common.Token, exchange common.Exchange, amount float64) (*types.Transaction, error) {
	exchangeAddress, supported := exchange.Address(token)
	if !supported {
		err := errors.New("Token is not supported on exchange")
		log.Printf("ERROR: Intermediator: Token %s is not supported on Exchange %v", token.ID, exchange.ID)
		return nil, err
	}
	log.Printf("Intermediator: exchange address to receive token is %s", exchangeAddress.Hex())
	IAmount := getBigIntFromFloat(amount, token.Decimal)
	var tx *types.Transaction
	var err error
	if token.ID == "ETH" {
		tx, err = self.blockchain.SendETHFromAccountToExchange(IAmount, exchangeAddress)
	} else {
		tx, err = self.blockchain.SendTokenFromAccountToExchange(IAmount, exchangeAddress, ethereum.HexToAddress(token.Address))
	}
	if err != nil {
		log.Printf("ERROR: Intermediator: Can not send transaction to exchange: %v", err)
		return nil, err
	}
	log.Printf("Intermediator: Transaction submitted. Tx is: \n %v", tx)
	return tx, nil
}

func (self *Intermediator) RunStatusChecker() {
	for {
		log.Printf("Intermediator: waiting for signal from status checker channel")
		t := <-self.runner.GetStatusTicker()
		log.Printf("Intermediator: got signal from status checker with timestamp %d", common.TimeToTimepoint(t))
		accBalance, err := self.FetchAccountBalanceFromBlockchain(common.TimeToTimepoint(t))
		if err != nil {
			log.Printf("Intermediator: can not get account balance %v", err)
		}
		for k, v := range accBalance {
			token, _ := common.GetToken(k)
			log.Printf("Intermediator: Account balance of token %v after is %.5f", k, v.ToBalanceResponse(token.Decimal).Balance)
		}
		depositBalance, err := self.blockchain.FetchBalanceData(ethereum.HexToAddress("0xf2121ce5dfabbae1056f9d219a88304a81d1190c"), nil, common.TimeToTimepoint(t))
		for k, v := range depositBalance {
			token, _ := common.GetToken(k)
			log.Printf("Intermediator: Deposit balance of token %v after is %.5f", k, v.ToBalanceResponse(token.Decimal).Balance)
		}
		self.CheckAccStatusFromBlockChain(common.TimeToTimepoint(t))
	}
}

func (self *Intermediator) RunTxChecker() {
	for {
		log.Printf("Intermediator: waiting for signal from status checker channel")
		t := <-self.runner.GetStatusTicker()
		log.Printf("Intermediator: got signal from status checker with timestamp %d", common.TimeToTimepoint(t))
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
	depositstatus := make(map[common.ActivityID]DepositAction)
	return &Intermediator{storage, runner, depositstatus, address, blockchain}
}
