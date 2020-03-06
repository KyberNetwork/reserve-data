package coinbase

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/KyberNetwork/reserve-data/feed-provider/common"
)

type Price struct {
	Price    float64 `json:",string"`
	Size     float64 `json:",string"`
	NumOrder int
}

func (n *Price) UnmarshalJSON(buf []byte) error {
	var price, size json.Number
	tmp := []interface{}{&price, &size, &n.NumOrder}
	err := json.Unmarshal(buf, &tmp)
	if err != nil {
		return err
	}
	n.Price, err = price.Float64()
	if err != nil {
		return err
	}
	n.Size, err = size.Float64()
	if err != nil {
		return err
	}
	return nil
}

type Resp struct {
	Sequence int64   `json:"sequence"`
	Code     int     `json:"code"`
	Msg      string  `json:"msg"`
	Bids     []Price `json:"bids"`
	Asks     []Price `json:"asks"`
}

func (c *Resp) toFeed(amount float64) common.Feed {
	ask, err := getAfp(c.Asks, amount, false)
	if err != nil {
		return common.Feed{
			Error: err.Error(),
		}
	}
	bid, err := getAfp(c.Bids, amount, true)
	if err != nil {
		return common.Feed{
			Error: err.Error(),
		}
	}
	return common.Feed{
		Ask:   ask,
		Bid:   bid,
		Valid: true,
	}
}

func getAfp(orderbooks []Price, amount float64, isBid bool) (float64, error) {
	var (
		remain                = amount
		totalSize, totalValue float64
	)
	sort.Slice(orderbooks, func(p, q int) bool {
		if isBid {
			return orderbooks[p].Price > orderbooks[q].Price
		}
		return orderbooks[p].Price < orderbooks[q].Price
	})
	for _, o := range orderbooks {
		size := o.Size
		if remain < o.Price*size {
			size = remain / o.Price
		}
		value := o.Price * size
		totalValue += value
		totalSize += size
		remain -= value
		if remain <= 0 {
			break
		}
	}
	if remain > 0 {
		return 0, fmt.Errorf("could not collect %f of DAI, isBid is %v", amount, isBid)
	}
	if totalSize == 0 {
		return 0, fmt.Errorf("total size of orderbook is 0, isBid is %v", isBid)
	}
	return totalValue / totalSize, nil
}
