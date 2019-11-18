package world

import (
	"bytes"
	"encoding/json"
	"errors"
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

func (tw *TheWorld) getPublic(url string, dst interface{}) error {
	var (
		client = &http.Client{Timeout: 30 * time.Second}
	)
	caller := common.GetCallerFunctionName()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	log.Printf("get %s - %s", caller, url)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("request on %s failed, %v\n", caller, err)
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		errMsg := fmt.Sprintf("%s read response error %v", caller, err)
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("unexpected return code: %d, response text %s", resp.StatusCode, body)
		log.Println(errMsg)
		return errors.New(errMsg)
	}
	d := json.NewDecoder(bytes.NewBuffer(body))
	d.DisallowUnknownFields()
	if err = d.Decode(dst); err != nil {
		errMsg := fmt.Sprintf("%s unmarshal failed, err = %v, response text %s", caller, err, body)
		log.Println(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (tw *TheWorld) getBinanceInfo(ep string) common.BinanceData {
	var (
		url    = ep
		result = common.BinanceData{}
	)

	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getOneForgeGoldUSDInfo() common.OneForgeGoldData {

	url := tw.endpoint.OneForgeGoldUSDDataEndpoint()
	var result common.OneForgeGoldData
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = true
		result.Message = err.Error()
	}
	return result
}

func (tw *TheWorld) getOneForgeGoldETHInfo() common.OneForgeGoldData {
	url := tw.endpoint.OneForgeGoldETHDataEndpoint()
	result := common.OneForgeGoldData{}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = true
		result.Message = err.Error()
	}
	return result
}

func (tw *TheWorld) getDGXGoldInfo() common.DGXGoldData {
	url := tw.endpoint.GoldDataEndpoint()
	result := common.DGXGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getGDAXGoldInfo() common.GDAXGoldData {
	url := tw.endpoint.GDAXDataEndpoint()
	result := common.GDAXGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getKrakenGoldInfo() common.KrakenGoldData {
	url := tw.endpoint.KrakenDataEndpoint()
	result := common.KrakenGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) getGeminiGoldInfo() common.GeminiGoldData {
	url := tw.endpoint.GeminiDataEndpoint()
	result := common.GeminiGoldData{
		Valid: true,
	}
	err := tw.getPublic(url, &result)
	if err != nil {
		result.Error = err.Error()
		result.Valid = false
	} else {
		result.Valid = true
	}
	return result
}

func (tw *TheWorld) GetGoldInfo() (common.GoldData, error) {
	return common.GoldData{
		DGX:         tw.getDGXGoldInfo(),
		OneForgeETH: tw.getOneForgeGoldETHInfo(),
		OneForgeUSD: tw.getOneForgeGoldUSDInfo(),
		GDAX:        tw.getGDAXGoldInfo(),
		Kraken:      tw.getKrakenGoldInfo(),
		Gemini:      tw.getGeminiGoldInfo(),
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
