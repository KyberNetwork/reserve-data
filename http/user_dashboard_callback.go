package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KyberNetwork/reserve-data/http/httputil"
	ethereum "github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

// UpdateUserAddresses receive callback from userdashboard and save kycinfo
func (hs *HTTPServer) UpdateUserAddresses(c *gin.Context) {
	var err error
	postForm, ok := hs.Authenticated(c, []string{"user", "addresses", "timestamps"}, []Permission{ConfirmConfPermission})
	if !ok {
		return
	}
	user := postForm.Get("user")
	addresses := postForm.Get("addresses")
	times := postForm.Get("timestamps")
	addrs := []ethereum.Address{}
	timestamps := []uint64{}
	addrsStr := strings.Split(addresses, "-")
	timesStr := strings.Split(times, "-")
	if len(addrsStr) != len(timesStr) {
		httputil.ResponseFailure(c, httputil.WithReason("addresses and timestamps must have the same number of elements"))
		return
	}
	for i, addr := range addrsStr {
		var (
			t uint64
			a = ethereum.HexToAddress(addr)
		)
		t, err = strconv.ParseUint(timesStr[i], 10, 64)
		if a.Big().Cmp(ethereum.Big0) != 0 && err == nil {
			addrs = append(addrs, a)
			timestamps = append(timestamps, t)
		}
	}
	if len(addrs) == 0 {
		httputil.ResponseFailure(c, httputil.WithReason(fmt.Sprintf("user %s doesn't have any valid addresses in %s", user, addresses)))
		return
	}

	err = hs.stat.UpdateUserAddresses(user, addrs, timestamps)
	if err != nil {
		httputil.ResponseFailure(c, httputil.WithError(err))
		return
	}
	httputil.ResponseSuccess(c)
}
