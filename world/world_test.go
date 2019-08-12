package world

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/KyberNetwork/reserve-data/cmd/deployment"
	"github.com/KyberNetwork/reserve-data/common/testutil"
)

//This test require external resources
func TestTheWorld_GetGoldInfo(t *testing.T) {
	t.Skip()
	sugar := testutil.MustNewDevelopmentSugaredLogger()
	world, err := NewTheWorld(sugar, deployment.Development, "../cmd/config.json")
	require.NoError(t, err)
	_, err = world.GetGoldInfo()
	require.NoError(t, err)
}
