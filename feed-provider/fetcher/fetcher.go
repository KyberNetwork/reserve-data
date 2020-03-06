package fetcher

import (
	"github.com/KyberNetwork/reserve-data/feed-provider/common"
)

type Fetcher interface {
	GetData() common.Feed
}
