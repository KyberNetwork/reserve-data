package gaspricedataclient

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGasPriceDataClient_GetGas(t *testing.T) {
	t.Skip()
	c := New(&http.Client{}, "http://localhost:8088/api/v1/gas")
	r, err := c.GetGas()
	require.NoError(t, err)
	t.Logf("%+v", r)
}
