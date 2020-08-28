package gasstation

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestClient_ETHGas(t *testing.T) {
	c := New(&http.Client{}, "", "", zap.L().Sugar())
	gas, err := c.ETHGas()
	require.NoError(t, err)
	require.True(t, gas.Fast > 0)
	require.True(t, gas.Average > 0)
	require.True(t, gas.Fastest > 0)
	require.True(t, gas.SafeLow > 0)
}

type roundTrip struct {
}

func (r roundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	if request.FormValue("api-key") != "abc" {
		return nil, errors.New("unexpect api-key")
	}
	return &http.Response{StatusCode: http.StatusOK, Body: ioutil.NopCloser(bytes.NewBufferString(`{"fast":310,"fastest":500,"safeLow":250,"average":280,"block_time":14.448275862068966,"blockNum":10338433,"speed":0.9972105008985533,"safeLowWait":22.7,"avgWait":2.8,"fastWait":0.5,"fastestWait":0.5,"gasPriceRange":{"4":240.8,"6":240.8,"8":240.8,"10":240.8,"20":240.8,"30":240.8,"40":240.8,"50":240.8,"60":240.8,"70":240.8,"80":240.8,"90":240.8,"100":240.8,"110":240.8,"120":240.8,"130":240.8,"140":240.8,"150":240.8,"160":240.8,"170":240.8,"180":240.8,"190":240.8,"200":240.8,"220":240.8,"240":240.8,"250":22.7,"260":9.6,"280":2.8,"300":0.7,"310":0.5,"320":0.5,"340":0.5,"360":0.5,"380":0.5,"400":0.5,"420":0.5,"440":0.5,"460":0.5,"480":0.5,"500":0.5}}`))}, nil
}

func TestClientWithKey(t *testing.T) {
	c := New(&http.Client{Transport: &roundTrip{}}, "abc", "", zap.L().Sugar())
	_, err := c.ETHGas()
	require.NoError(t, err)
}

func TestClientEthscanAPI(t *testing.T) {
	c := New(&http.Client{}, "", "", zap.L().Sugar())
	gas, err := c.getGasFromEtherscan()
	require.NoError(t, err)
	fast, err := strconv.ParseFloat(gas.Result.FastGasPrice, 64)
	require.NoError(t, err)
	require.True(t, fast > 0)
	safe, err := strconv.ParseFloat(gas.Result.SafeGasPrice, 64)
	require.NoError(t, err)
	require.True(t, safe > 0)
}
