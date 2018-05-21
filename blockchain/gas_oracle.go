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
	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for {
			err := self.RunGasPricing()
			if err != nil {
				log.Printf("Error pricing gas from Gasstation: %v", err)
			}
		}
		<-ticker.C
	}()
}

func (self *GasOracle) RunGasPricing() error {
	client := &http.Client{Timeout: 3 * time.Second}
	r, err := client.Get(GasStationUrl)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	gasOracle := GasOracle{}
	err = json.Unmarshal(body, &gasOracle)
	if err != nil {
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
