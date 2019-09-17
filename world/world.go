package world

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
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

	log.Printf("%s fetch %s", caller, url)

	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("request on %s failed, %v\n", caller, err)
		return err
	}
	defer func() {
		if cErr := resp.Body.Close(); cErr != nil {
			log.Printf("failed to close response body: %s", cErr.Error())
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected return code: %d", resp.StatusCode)
	}

	if err = json.NewDecoder(resp.Body).Decode(dst); err != nil {
		log.Printf("%s decode failed, %v\n", caller, err)
		return err
	}

	return nil
}

func (tw *TheWorld) getOneForgeGoldUSDInfo() common.OneForgeGoldData {
	url := tw.endpoint.OneForgeGoldUSDDataEndpoint()
	result := common.OneForgeGoldData{}
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
		result.Valid = false
		result.Error = err.Error()
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
		result.Valid = false
		result.Error = err.Error()
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
		result.Valid = false
		result.Error = err.Error()
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
		result.Valid = false
		result.Error = err.Error()
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
func NewTheWorld(dpl deployment.Deployment, keyfile string) (*TheWorld, error) {
	switch dpl {
	case deployment.Development,
		deployment.Production,
		deployment.Kovan,
		deployment.Staging,
		deployment.Ropsten,
		deployment.Analytic:
		// TODO: make key file a cli flag
		endpoint, err := NewRealEndpointFromFile(keyfile)
		if err != nil {
			return nil, err
		}
		return &TheWorld{endpoint}, nil
	case deployment.Simulation:
		return &TheWorld{SimulatedEndpoint{}}, nil
	}
	panic("unsupported environment")
}
