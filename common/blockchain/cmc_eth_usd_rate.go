package blockchain

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
)

const (
	cmcEthereumPricingAPIEndpoint = "https://graphs2.coinmarketcap.com/currencies/ethereum/"
	cmcTopUSDPricingAPIEndpoint   = "https://api.coinmarketcap.com/v1/ticker/?convert=USD&limit=10"
)

type CoinCapRateResponse []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Rank     string `json:"rank"`
	PriceUSD string `json:"price_usd"`
}

type CMCEthUSDRate struct {
	mu                *sync.RWMutex
	cachedRates       [][]float64
	currentCacheMonth uint64
	realtimeTimepoint uint64
	realtimeRate      float64
}

type RateLogResponse struct {
	PriceUSD [][]float64 `json:"price_usd"`
}

func GetTimeStamp(year int, month time.Month, day int, hour int, minute int, sec int, nanosec int, loc *time.Location) uint64 {
	return uint64(time.Date(year, month, day, hour, minute, sec, nanosec, loc).Unix() * 1000)
}

func GetMonthTimeStamp(timepoint uint64) uint64 {
	t := time.Unix(int64(timepoint/1000), 0).UTC()
	month, year := t.Month(), t.Year()
	return GetTimeStamp(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func GetNextMonth(month, year int) (int, int) {
	var toMonth, toYear int
	if int(month) == 12 {
		toMonth = 1
		toYear = year + 1
	} else {
		toMonth = int(month) + 1
		toYear = year
	}
	return toMonth, toYear
}

func (ethUSDRate *CMCEthUSDRate) GetUSDRate(timepoint uint64) float64 {
	if timepoint >= ethUSDRate.realtimeTimepoint {
		return ethUSDRate.realtimeRate
	}
	return ethUSDRate.rateFromCache(timepoint)
}

func (ethUSDRate *CMCEthUSDRate) rateFromCache(timepoint uint64) float64 {
	ethUSDRate.mu.Lock()
	defer ethUSDRate.mu.Unlock()
	monthTimeStamp := GetMonthTimeStamp(timepoint)
	if monthTimeStamp != ethUSDRate.currentCacheMonth {
		ethRates, err := fetchRate(timepoint)
		if err != nil {
			log.Println("Cannot get rate from coinmarketcap")
			return ethUSDRate.realtimeRate
		}
		rate, err := findEthRate(ethRates, timepoint)
		if err != nil {
			log.Println(err)
			return ethUSDRate.realtimeRate
		}
		ethUSDRate.currentCacheMonth = monthTimeStamp
		ethUSDRate.cachedRates = ethRates
		return rate
	}
	rate, err := findEthRate(ethUSDRate.cachedRates, timepoint)
	if err != nil {
		return ethUSDRate.realtimeRate
	}
	return rate
}

func fetchRate(timepoint uint64) ([][]float64, error) {
	t := time.Unix(int64(timepoint/1000), 0).UTC()
	month, year := t.Month(), t.Year()
	fromTime := GetTimeStamp(year, month, 1, 0, 0, 0, 0, time.UTC)
	toMonth, toYear := GetNextMonth(int(month), year)
	toTime := GetTimeStamp(toYear, time.Month(toMonth), 1, 0, 0, 0, 0, time.UTC)
	api := cmcEthereumPricingAPIEndpoint + strconv.FormatInt(int64(fromTime), 10) + "/" + strconv.FormatInt(int64(toTime), 10) + "/"
	resp, err := http.Get(api)
	if err != nil {
		return [][]float64{}, err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return [][]float64{}, err
	}
	rateResponse := RateLogResponse{}
	err = json.Unmarshal(body, &rateResponse)
	if err != nil {
		return [][]float64{}, err
	}
	ethRates := rateResponse.PriceUSD
	return ethRates, nil
}

func findEthRate(ethRateLog [][]float64, timepoint uint64) (float64, error) {
	var ethRate float64
	for _, e := range ethRateLog {
		if uint64(e[0]) >= timepoint {
			ethRate = e[1]
			return ethRate, nil
		}
	}
	return 0, errors.New("Cannot find ether rate corresponding with the timepoint")
}

func (ethUSDRate *CMCEthUSDRate) RunGetEthRate() {
	tick := time.NewTicker(10 * time.Minute)
	go func() {
		for {
			err := ethUSDRate.FetchEthRate()
			if err != nil {
				log.Println(err)
			}
			<-tick.C
		}
	}()
}

func (ethUSDRate *CMCEthUSDRate) FetchEthRate() (err error) {
	resp, err := http.Get(cmcTopUSDPricingAPIEndpoint)
	if err != nil {
		return err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	rateResponse := CoinCapRateResponse{}
	err = json.Unmarshal(body, &rateResponse)
	if err != nil {
		log.Printf("Getting eth-usd rate failed: %+v", err)
	} else {
		for _, rate := range rateResponse {
			if rate.Symbol == "ETH" {
				newrate, err := strconv.ParseFloat(rate.PriceUSD, 64)
				if err != nil {
					log.Printf("Cannot get usd rate: %s", err.Error())
					return err
				} else {
					if ethUSDRate.realtimeRate == 0 {
						// set realtimeTimepoint to the timepoint that realtime rate is updated for the
						// first time
						ethUSDRate.realtimeTimepoint = common.GetTimepoint()
					}
					ethUSDRate.realtimeRate = newrate
					return nil
				}
			}
		}
	}
	return nil
}

//Run run get ETH-USD rate from CoinMarketCap
func (ethUSDRate *CMCEthUSDRate) Run() {
	// run real time fetcher
	ethUSDRate.RunGetEthRate()
}

//NewCMCEthUSDRate return a new CMCEthUSDRate instance
func NewCMCEthUSDRate() *CMCEthUSDRate {
	result := &CMCEthUSDRate{
		mu: &sync.RWMutex{},
	}
	result.Run()
	return result
}
