package world

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/KyberNetwork/reserve-data/common"
)

//TheWorld is the concrete implementation of fetcher.TheWorld interface.
type TheWorld struct {
	endpoint Endpoint
}

func (self *TheWorld) getOneForgeGoldUSDInfo() common.OneForgeGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.OneForgeGoldUSDDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	} else {
		defer func() {
			if cErr := resp.Body.Close(); cErr != nil {
				log.Printf("Close http body error: %s", cErr.Error())
			}
		}()
		if resp.StatusCode != 200 {
			err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
			return common.OneForgeGoldData{
				Error:   true,
				Message: err.Error(),
			}
		}
		respBody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return common.OneForgeGoldData{
				Error:   true,
				Message: err.Error(),
			}
		}
		log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
		result := common.OneForgeGoldData{}
		err = json.Unmarshal(respBody, &result)
		if err != nil {
			result.Error = true
			result.Message = err.Error()
		}
		return result
	}
}

func (self *TheWorld) getOneForgeGoldETHInfo() common.OneForgeGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.OneForgeGoldETHDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.OneForgeGoldData{
			Error:   true,
			Message: err.Error(),
		}
	}
	log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
	result := common.OneForgeGoldData{}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		result.Error = true
		result.Message = err.Error()
	}
	return result
}

func (self *TheWorld) getDGXGoldInfo() common.DGXGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.GoldDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Close reponse body error: %s", cErr.Error())
		}
	}()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.DGXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
	result := common.DGXGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (self *TheWorld) getGDAXGoldInfo() common.GDAXGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.GDAXDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.GDAXGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
	result := common.GDAXGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (self *TheWorld) getKrakenGoldInfo() common.KrakenGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.KrakenDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.KrakenGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
	result := common.KrakenGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (self *TheWorld) getGeminiGoldInfo() common.GeminiGoldData {
	client := &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
	url := self.endpoint.GeminiDataEndpoint()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Accept", "application/json")
	var err error
	var respBody []byte
	log.Printf("request to gold feed endpoint: %s", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("Response body close error: %s", cErr.Error())
		}
	}()
	if resp.StatusCode != 200 {
		err = fmt.Errorf("Gold feed returned with code: %d", resp.StatusCode)
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.GeminiGoldData{
			Valid: false,
			Error: err.Error(),
		}
	}
	log.Printf("request to %s, got response from gold feed %s", req.URL, respBody)
	result := common.GeminiGoldData{
		Valid: true,
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		result.Valid = false
		result.Error = err.Error()
	}
	return result
}

func (self *TheWorld) GetGoldInfo() (common.GoldData, error) {
	return common.GoldData{
		DGX:         self.getDGXGoldInfo(),
		OneForgeETH: self.getOneForgeGoldETHInfo(),
		OneForgeUSD: self.getOneForgeGoldUSDInfo(),
		GDAX:        self.getGDAXGoldInfo(),
		Kraken:      self.getKrakenGoldInfo(),
		Gemini:      self.getGeminiGoldInfo(),
	}, nil
}

//NewTheWorld return new world instance
func NewTheWorld(env string, keyfile string) (*TheWorld, error) {
	switch env {
	case common.DevMode, common.KovanMode, common.MainnetMode, common.ProductionMode, common.StagingMode, common.RopstenMode, common.AnalyticDevMode:
		endpoint, err := NewRealEndpointFromFile(keyfile)
		if err != nil {
			return nil, err
		}
		return &TheWorld{endpoint}, nil
	case common.SimulationMode:
		return &TheWorld{SimulatedEndpoint{}}, nil
	}
	panic("unsupported environment")
}
