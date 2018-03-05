package ratelog

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
	"errors"

	"github.com/KyberNetwork/reserve-data/common"
)

type RateLogResponse struct {
	PriceUSD [][]float64 `json:"price_usd"`
}

type RateLog struct {
	storage Storage
}

func NewRateLog(storage Storage) *RateLog {
	return &RateLog{
		storage: storage,
	}
}

func (self *RateLog) GetEthRate(timePoint uint64) (float64, error) {
	if IsRecent(timePoint) {
		return 0, errors.New("time point is too near with current time")
	}
	ethRateLog := self.storage.GetEthRateLog(GetMonthTimeStamp(timePoint))
	if len(ethRateLog) != 0 {
		ethRate := findEthRate(ethRateLog, timePoint)
		log.Println("ethRate: ", ethRate)
		if ethRate != 0 {
			return ethRate, nil
		} else {
			return self.GetRateFromCoinMC(timePoint)
		}
	} else {
		return self.GetRateFromCoinMC(timePoint)
	}
	return 0, errors.New("Cannot get ether rate from rate log")
}

func (self *RateLog) GetRateFromCoinMC(timePoint uint64) (float64, error) {
	bulkEthRateLog, ethRate, monthTimePoint, err := fetchRate(timePoint)
	if err == nil && ethRate != 0 {
		err = self.storage.StoreEthRateLog(bulkEthRateLog, monthTimePoint)
		if err != nil {
			log.Println("failed to save ether rate log: ", err)
		} else {
			log.Println("save new rate log")
		}
		return ethRate, nil
	} else {
		log.Println("Cannot get rate from coinmarketcap: ", err)
		return 0, err
	}
	return 0, err
}

func findEthRate(ethRateLog []common.EthRateLog, timePoint uint64) float64 {
	var ethRate float64
	for _, e := range ethRateLog {
		if e.Timepoint >= timePoint {
			ethRate = e.Usd
			break
		}
	}
	return ethRate
}

func fetchRate(timePoint uint64) ([]common.EthRateLog, float64, uint64, error) {
	bulkEthRateLog := []common.EthRateLog{}
	t := time.Unix(int64(timePoint/1000), 0).UTC()
	month, year := t.Month(), t.Year()
	fromTime := GetTimeStamp(year, month, 1, 0, 0, 0, 0, time.UTC)
	toMonth, toYear := GetNextMonth(int(month), year)
	toTime := GetTimeStamp(toYear, time.Month(toMonth), 1, 0, 0, 0, 0, time.UTC)
	api := "https://graphs2.coinmarketcap.com/currencies/ethereum/" + strconv.FormatInt(int64(fromTime), 10) + "/" + strconv.FormatInt(int64(toTime), 10) + "/"
	resp, err := http.Get(api)
	if err != nil {
		return bulkEthRateLog, 0, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bulkEthRateLog, 0, 0, err
	}
	rateResponse := RateLogResponse{}
	err = json.Unmarshal(body, &rateResponse)
	if err != nil {
		return bulkEthRateLog, 0, 0, err
	}
	var ethRate float64
	bulkPriceUsd := rateResponse.PriceUSD
	for _, p := range bulkPriceUsd {
		tickTimeStamp := uint64(p[0])
		if ethRate == 0 && timePoint <= tickTimeStamp {
			ethRate = p[1]
		}
		ethRateLog := common.EthRateLog{
			Timepoint: tickTimeStamp,
			Usd:       p[1],
		}
		bulkEthRateLog = append(bulkEthRateLog, ethRateLog)
	}
	if uint64(ethRate) == 0 {
		return []common.EthRateLog{}, 0, 0, nil
	}
	return bulkEthRateLog, ethRate, fromTime, nil
}
