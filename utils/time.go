package utils

import "time"

// 获取当天0点时间戳
func GetDayZeroTime() int64 {
	now := time.Now()
	year, month, day := now.Date()

	zero := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	return zero.Unix()
}


// 获取本周0点时间戳
func GetWeekZeroTime() int64 {
	now := time.Now()
	year, month, day := now.Date()

	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}

	zero := time.Date(year, month, day, 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return zero.Unix()
}

// 获取本月0点时间戳
func GetMonthZeroTime() int64 {
	now := time.Now()
	year, month, _ := now.Date()

	zero := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	return zero.Unix()
}

/*
	判断当前时间是否在当天某个时间段内
	hour: 区间起始小时
	min: 区间起始分钟
	keepMin: 区间总共多少分钟
	如: ChkNowInTimeInner(9, 30, 90), 判断当前时间是否坐落于 9:30 ~ 11:00
*/
func ChkNowInTimeInner(hour int, min int, keepMin int64) bool {
	now := time.Now()
	year, month, day := now.Date()

	begin := time.Date(year, month, day, hour, min, 0, 0, time.Local)
	if now.Unix() >= begin.Unix() && now.Unix() <= begin.Unix()+keepMin*60 {
		return true
	}

	return false
}
