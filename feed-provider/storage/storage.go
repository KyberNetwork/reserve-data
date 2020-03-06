package storage

import (
	"github.com/KyberNetwork/reserve-data/feed-provider/common"
)

type Storage interface {
	Save(feedName string, data common.Feed) error
	Load(feedName string) common.Feed
}
