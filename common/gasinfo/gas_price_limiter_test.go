package gasinfo

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/blockchain"
)

func TestGasPriceLimiter_MaxGasPrice(t *testing.T) {
	c, err := ethclient.Dial("https://mainnet.infura.io/v3/3243b06bf0334cff8e468bf90ce6e08c")
	require.NoError(t, err)
	kc, err := blockchain.NewNetworkProxy(common.HexToAddress("0x818E6FECD516Ecc3849DAf6845e3EC868087B755"), c)
	require.NoError(t, err)
	gp := NewNetworkGasPriceLimiter(kc, 300)
	mgp, err := gp.MaxGasPrice()
	require.NoError(t, err)
	t.Logf("max gasprice %v", mgp)

}
