package comparerates

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/KyberNetwork/reserve-data/cmd/configuration"
	"github.com/KyberNetwork/reserve-data/common"
)

const (
	BaseURL    string        = "https://ropsten-core.kyber.network"
	TweiAdjust float64       = 1000000000000000000
	SleepTime  time.Duration = 60 //sleep time for forever run mode
	DifferRate float64       = 0.001
)

//AllRateHTTPReply To hold all rate response and its request status
type AllRateHTTPReply struct {
	Data    []common.AllRateResponse
	Success bool
}

//AllActionHTTPReply To hold all activities response and its request status
type AllActionHTTPReply struct {
	Data    []common.ActivityRecord
	Success bool
}

/* GetActivitiesResponse :
 */
func GetActivitiesResponse(url string, params map[string]string, config configuration.Config) (AllActionHTTPReply, error) {
	timepoint := common.GetTimepoint()
	nonce := strconv.FormatUint(timepoint, 10)
	var allActionRep AllActionHTTPReply
	params["nonce"] = nonce
	data, err := GetResponse("GET", fmt.Sprintf("%s/%s", url, "activities"), params, true, config)

	if err != nil {
		fmt.Println("can't get response", err)
	} else {
		if err := json.Unmarshal(data, &allActionRep); err != nil {
			fmt.Println("can't decode the reply", err)
			return allActionRep, err
		}
	}
	return allActionRep, nil
}

func GetAllRateResponse(url string, params map[string]string, config configuration.Config) (AllRateHTTPReply, error) {
	var allRateRep AllRateHTTPReply
	data, err := GetResponse("GET", fmt.Sprintf("%s/%s", url, "get-all-rates"), params, false, config)

	if err != nil {
		fmt.Println("can't get response", err)
	} else {
		if err := json.Unmarshal(data, &allRateRep); err != nil {
			fmt.Println("can't decode the reply", err)
			return allRateRep, err
		}
	}
	return allRateRep, nil

}

func RateDifference(r1, r2 float64) float64 {
	return ((r2 - r1) / r1)
}

func printInterfaceMap(inf map[string]interface{}) {
	for k, v := range inf {
		switch v.(type) {
		case float64:
			log.Printf("\t\t %s %.5f \n", k, v.(float64))
		case []interface{}:
			log.Printf("\t\t %s \n", k)
			for _, vv := range v.([]interface{}) {
				log.Printf("\t\t\t %v", vv)
			}
		}
	}
}

func printAction(oneAct common.ActivityRecord) {
	i := int64(oneAct.Timestamp.ToUint64()) / 1000
	log.Printf("\t Time: %v \n", time.Unix(i, 0))
	log.Printf("\t Timestamp: %v", oneAct.Timestamp)
	log.Printf("\t Activity : %v\n", oneAct.Action)
	log.Printf("\t Destination : %v\n", oneAct.Destination)
	log.Printf("\t ExchangeStatus : %v\n", oneAct.ExchangeStatus)
	log.Printf("\t ID : %v\n", oneAct.ID)
	log.Printf("\t MiningStatus : %v\n", oneAct.MiningStatus)
	log.Printf("\t Params : \n")
	printInterfaceMap(oneAct.Params)
	log.Printf("\t Result : \n")
	printInterfaceMap(oneAct.Params)
}

func printRateResponse(oneRate common.AllRateResponse) {
	log.Printf("Data from get_all_rates:\n")
	log.Printf("All data were the same from Block number %v to Block number %v \n", oneRate.BlockNumber, oneRate.ToBlockNumber)
	i := int64(oneRate.Timestamp.ToUint64()) / 1000
	log.Printf("\t Time: %v\n", time.Unix(i, 0))
	log.Printf("\t TimeStamp %v\n", oneRate.Timestamp)
	log.Printf("\t Error: %v\n", oneRate.Error)
	log.Printf("\t Data: \n")
	for k, v := range oneRate.Data {
		log.Printf("\t Token\t\t BaseBuy\t\t BaseSell\t Error\t Block\t \n")
		log.Printf("\t %s \t %v \t %v\t %v\t %v ", k, v.BaseBuy, v.BaseSell, v.Error, v.Block)
		log.Printf("\t CompactBuy\t CompactSell\t Rate\t Valid\t TimeStamp\n")
		log.Printf("\t %v\t\t %v\t\t %v\t %v\t %v \n\n", v.CompactBuy, v.CompactSell, v.Rate, v.Valid, v.Timestamp)
	}
	log.Printf("\t Valid: %t\n", oneRate.Valid)
	log.Printf("\t Version: %v \n", oneRate.Version)
}

func CompareRate(oneAct common.ActivityRecord, oneRate common.AllRateResponse, blockID uint64) {
	tokenIDs, asrt := oneAct.Params["tokens"].([]interface{})
	buys, asrt1 := oneAct.Params["buys"].([]interface{})
	sells, asrt2 := oneAct.Params["sells"].([]interface{})
	warning := false
	if asrt && asrt1 && asrt2 {
		for idx, tokenID := range tokenIDs {
			tokenid, _ := tokenID.(string)
			val, ok := oneRate.Data[tokenid]
			if ok {
				differ := RateDifference(val.BaseBuy*(1+float64(val.CompactBuy)/1000.0)*TweiAdjust, buys[idx].(float64))
				if math.Abs(differ) > DifferRate {
					defer log.Printf("block %d set a buys rate differ %.5f%% than get rate at token %s \n", blockID, differ*100, tokenid)
					warning = true
				}
				differ = RateDifference(val.BaseSell*(1+float64(val.CompactSell)/1000.0)*TweiAdjust, sells[idx].(float64))
				if math.Abs(differ) > DifferRate {
					defer log.Printf("block %d set a sell rate differ %.5f%% than get rate at token %s \n", blockID, differ*100, tokenid)
					warning = true
				}
			}
		}
	}
	if warning {
		log.Printf("There was different in set rate at block %d \n", blockID)
		printAction(oneAct)
		printRateResponse(oneRate)
	}
}

func CompareRates(acts []common.ActivityRecord, rates []common.AllRateResponse) {
	idx := 0
	for _, oneAct := range acts {
		if (oneAct.Action == "set_rates") && (oneAct.MiningStatus == "mined") {
			_, ok := oneAct.Result["blockNumber"]
			if ok {
				curBlock := uint64(oneAct.Result["blockNumber"].(float64))
				for (idx < len(rates)) && (curBlock < rates[idx].ToBlockNumber) {
					idx++
				}
				if (idx < len(rates)) && (curBlock <= rates[idx].BlockNumber) && (curBlock >= rates[idx].ToBlockNumber) {
					fmt.Printf("\n Block %d is found between block %d to block %d \n", curBlock, rates[idx].BlockNumber, rates[idx].ToBlockNumber)
					CompareRate(oneAct, rates[idx], curBlock)
				} else {
					fmt.Printf("\n Block %d is not found\n", curBlock)
				}
			}
		}
	}
}

func DoQuery(url string, params map[string]string, config configuration.Config) {
	allActionRep, err := GetActivitiesResponse(url, params, config)
	if err != nil {
		log.Printf("couldn't get activites: %v", err)
		return
	}
	allRateRep, err := GetAllRateResponse(url, params, config)
	if err != nil {
		log.Printf("couldn't get all rates: %v", err)
		return
	}
	if (len(allActionRep.Data) < 1) || (len(allRateRep.Data) < 1) {
		log.Printf("One of the reply was empty")
		return
	}
	CompareRates(allActionRep.Data, allRateRep.Data)
}
