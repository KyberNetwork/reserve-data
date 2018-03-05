package ratelog

import (
	"time"
)

const TIME_DISTANCE uint64 = 10800000 // 3 hours in milisecond

func GetTimeStamp(year int, month time.Month, day int, hour int, minute int, sec int, nanosec int, loc *time.Location) uint64 {
	return uint64(time.Date(year, month, day, hour, minute, sec, nanosec, loc).Unix() * 1000)
}

func GetMonthTimeStamp(timePoint uint64) uint64 {
	t := time.Unix(int64(timePoint/1000), 0).UTC()
	month, year := t.Month(), t.Year()
	return GetTimeStamp(year, month, 1, 0, 0, 0, 0, time.UTC)
}

func IsRecent(timePoint uint64) bool {
	milisecNow := time.Now().Unix() * 1000
	if uint64(milisecNow) - timePoint < TIME_DISTANCE {
		return true
	}
	return false
}

func GetNextMonth(month, year int) (int, int) {
	var toMonth, toYear int
	if int(month) == 12 {
		toMonth = 1
		toYear = year + 1
	} else {
		toMonth = int(month) + 1
		toYear = year
	}
	return toMonth, toYear
}
