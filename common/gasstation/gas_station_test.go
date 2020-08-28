package gasstation

import (
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
