package storage

import (
	"github.com/KyberNetwork/reserve-data/core"
	"github.com/KyberNetwork/reserve-data/data"
	"github.com/KyberNetwork/reserve-data/data/fetcher"
	"github.com/KyberNetwork/reserve-data/data/storage"
)

var (
	_ data.Storage          = &storage.PostgresStorage{}
	_ data.GlobalStorage    = &storage.PostgresStorage{}
	_ fetcher.Storage       = &storage.PostgresStorage{}
	_ fetcher.GlobalStorage = &storage.PostgresStorage{}
	_ core.ActivityStorage  = &storage.PostgresStorage{}
)
