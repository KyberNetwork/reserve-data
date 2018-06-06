package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
	ihttp "github.com/KyberNetwork/reserve-data/http"
)

//Verification use for verify api related to core activities
//which interact with exchanges and blockchain
//so require auth for authentication api call
//and exchange for exchange call
//baseURL use for verify in different env
type Verification struct {
	auth      ihttp.Authentication
	exchanges []string
	baseURL   string
}

//DepositWithdrawResponse object using for parse response from deposit api
type DepositWithdrawResponse struct {
	Success bool              `json:"success"`
	ID      common.ActivityID `json:"id"`
	Reason  string            `json:"reason"`
}

var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

//UpdateBaseURL function to update base url when run via cmd
func (v *Verification) UpdateBaseURL(baseURL string) {
	v.baseURL = baseURL
}

//InitLogger categorize log
func InitLogger(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func (v *Verification) fillRequest(req *http.Request, signNeeded bool, timepoint uint64) {
	if req.Method == "POST" {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	req.Header.Add("Accept", "application/json")
	if signNeeded {
		q := req.URL.Query()
		q.Set("nonce", fmt.Sprintf("%d", timepoint))
		req.URL.RawQuery = q.Encode()
		req.Header.Add("signed", v.auth.KNSign(q.Encode()))
	}
}

//GetResponse get response from http request
func (v *Verification) GetResponse(
	method string, url string,
	params map[string]string, signNeeded bool, timepoint uint64) ([]byte, error) {

	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	req, _ := http.NewRequest(method, url, nil)
	req.Header.Add("Accept", "application/json")

	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	v.fillRequest(req, signNeeded, timepoint)
	var err error
	var respBody []byte
	resp, err := client.Do(req)
	if err != nil {
		return respBody, err
	}
	defer func() {
		if vErr := resp.Body.Close(); vErr != nil {
			log.Printf("Response body close error: %s", vErr.Error())
		}
	}()
	respBody, err = ioutil.ReadAll(resp.Body)
	return respBody, err
}

//GetPendingActivities return pending activities
func (v *Verification) GetPendingActivities(timepoint uint64) ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	respBody, err := v.GetResponse(
		"GET",
		v.baseURL+"/immediate-pending-activities",
		map[string]string{},
		true,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

//GetActivities return list of activities
func (v *Verification) GetActivities(timepoint, fromTime, toTime uint64) ([]common.ActivityRecord, error) {
	result := []common.ActivityRecord{}
	respBody, err := v.GetResponse(
		"GET",
		v.baseURL+"/activities",
		map[string]string{
			"fromTime": strconv.FormatUint(fromTime, 10),
			"toTime":   strconv.FormatUint(toTime, 10),
		},
		true,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

//GetAuthData return authentication data
func (v *Verification) GetAuthData(timepoint uint64) (common.AuthDataResponse, error) {
	result := common.AuthDataResponse{}
	respBody, err := v.GetResponse(
		"GET",
		v.baseURL+"/authdata",
		map[string]string{},
		true,
		timepoint,
	)
	if err == nil {
		err = json.Unmarshal(respBody, &result)
	}
	return result, err
}

//Deposit deposit token/eth to an exchange
func (v *Verification) Deposit(
	exchange, token, amount string, timepoint uint64) (common.ActivityID, error) {
	result := DepositWithdrawResponse{}
	log.Println("Start deposit")
	respBody, err := v.GetResponse(
		"POST",
		v.baseURL+"/deposit/"+exchange,
		map[string]string{
			"amount": amount,
			"token":  token,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result.ID, err
	}
	if vErr := json.Unmarshal(respBody, &result); vErr != nil {
		return result.ID, vErr
	}
	if result.Success != true {
		err = fmt.Errorf("Cannot deposit: %s", result.Reason)
	}
	return result.ID, err
}

//Withdraw withdraw token/eth from an exchange
func (v *Verification) Withdraw(
	exchange, token, amount string, timepoint uint64) (common.ActivityID, error) {
	result := DepositWithdrawResponse{}
	respBody, err := v.GetResponse(
		"POST",
		v.baseURL+"/withdraw/"+exchange,
		map[string]string{
			"amount": amount,
			"token":  token,
		},
		true,
		timepoint,
	)
	if err != nil {
		return result.ID, err
	}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return result.ID, err
	}
	if result.Success != true {
		err = fmt.Errorf("Cannot withdraw: %s", result.Reason)
	}
	return result.ID, nil
}

//CheckPendingActivities check pending activities
func (v *Verification) CheckPendingActivities(activityID common.ActivityID, timepoint uint64) {
	pendingActivities, err := v.GetPendingActivities(timepoint)
	if err != nil {
		Error.Println(err.Error())
		return
	}
	available := false
	for _, pending := range pendingActivities {
		if pending.ID == activityID {
			available = true
			break
		}
	}
	if !available {
		Error.Println("Deposit activity did not store")
		return
	}
	Info.Println("Pending activities stored success")
}

//CheckPendingAuthData check pending activity in authdata
func (v *Verification) CheckPendingAuthData(activityID common.ActivityID, timepoint uint64) {
	authData, err := v.GetAuthData(timepoint)
	if err != nil {
		Error.Println(err.Error())
	}
	available := false
	for _, pending := range authData.Data.PendingActivities {
		if activityID == pending.ID {
			available = true
			break
		}
	}
	if !available {
		Error.Println("Activity cannot find in authdata pending activity")
	}
	Info.Println("Stored pending auth data success")
}

//CheckActivities check activities api if they are working fine or not
func (v *Verification) CheckActivities(activityID common.ActivityID, timepoint uint64) {
	toTime := common.GetTimepoint()
	fromTime := toTime - 3600000
	activities, err := v.GetActivities(timepoint, fromTime, toTime)
	if err != nil {
		Error.Println(err.Error())
	}
	available := false
	for _, activity := range activities {
		if activity.ID == activityID {
			available = true
			break
		}
	}
	if !available {
		Error.Printf("Cannot find activity: %v", activityID)
	}
	Info.Printf("Activity %v stored successfully", activityID)
}

//VerifyDeposit verify deposit action to ensure it works as expected
func (v *Verification) VerifyDeposit() error {
	var err error
	timepoint := common.GetTimepoint()
	token, err := common.GetInternalToken("ETH")
	amount := getTokenAmount(0.5, token)
	Info.Println("Start deposit to exchanges")
	for _, exchange := range v.exchanges {
		activityID, vErr := v.Deposit(exchange, token.ID, amount, timepoint)
		if vErr != nil {
			return vErr
		}
		Info.Printf("Deposit id: %s", activityID)
		go v.CheckPendingActivities(activityID, timepoint)
		go v.CheckPendingAuthData(activityID, timepoint)
		go v.CheckActivities(activityID, timepoint)
	}
	return err
}

//VerifyWithdraw to ensure withdraw action works as expected
func (v *Verification) VerifyWithdraw() error {
	var err error
	timepoint := common.GetTimepoint()
	token, err := common.GetInternalToken("ETH")
	amount := getTokenAmount(0.5, token)
	for _, exchange := range v.exchanges {
		activityID, vErr := v.Withdraw(exchange, token.ID, amount, timepoint)
		if vErr != nil {
			return vErr
		}
		Info.Printf("Withdraw ID: %s", activityID)
		go v.CheckPendingActivities(activityID, timepoint)
		go v.CheckPendingAuthData(activityID, timepoint)
		go v.CheckPendingAuthData(activityID, timepoint)
	}
	return err
}

//RunVerification run package
func (v *Verification) RunVerification() error {
	Info.Println("Start verification")
	if err := v.VerifyDeposit(); err != nil {
		return err
	}
	if err := v.VerifyWithdraw(); err != nil {
		return err
	}
	return nil
}

// NewVerification init new verification object
func NewVerification(
	auth ihttp.Authentication) *Verification {
	params := os.Getenv("KYBER_EXCHANGES")
	exchanges := strings.Split(params, ",")
	return &Verification{
		auth,
		exchanges,
		"http://localhost:8000",
	}
}
