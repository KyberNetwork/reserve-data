package blockchain

import (
	"log"
	"time"
	"sync"
	"net/http"
	"io/ioutil"
	"encoding/json"
)

const GasStationUrl = "https://ethgasstation.info/json/ethgasAPI.json"

type GasOracle struct {
	mu       *sync.RWMutex     
	Standard float64  `json:"average"`
	Fast     float64  `json:"fast"`
	SafeLow  float64  `json:"safeLow"`
}

func NewGasOracle() *GasOracle {
	gasOracle := &GasOracle{
		mu: &sync.RWMutex{},
	}
	gasOracle.GasPricing()
	return gasOracle
}

func (self *GasOracle) GasPricing() {
	self.RunGasPricing()
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			select {
			case <- ticker.C:
				err := self.RunGasPricing()
				if err != nil {
					log.Printf("Error get GasOracle: %s", err.Error())
				}
			}
		}
	}()
}

func (self *GasOracle) RunGasPricing() error {
	client := &http.Client{Timeout: 3 * time.Second}
	r, err := client.Get(GasStationUrl)
	if err != nil {
		log.Println("Cannot get GasOracle from Gasstation")
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println("Failed to read response of Gasstation")
		return err
	}
	gasOracle := GasOracle{}
	err = json.Unmarshal(body, &gasOracle)
	if err != nil {
		log.Println("Failed to map data to GasOracle")
		return err
	}
	self.Set(gasOracle)
	return nil
}

func (self *GasOracle) Set(gasOracle GasOracle) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.SafeLow = gasOracle.SafeLow
	self.Standard = gasOracle.Standard
	self.Fast = gasOracle.Fast
}
