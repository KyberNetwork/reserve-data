package intermediator

import (
	"github.com/KyberNetwork/reserve-data/common"
)

type Storage interface {
	GetPendingActivitiesByDest(dest string) ([]common.ActivityRecord, error)
}
