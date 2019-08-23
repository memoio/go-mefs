package time

import (
	"strconv"
	"time"
)

//获取时间戳
func GetTimestamp() string {
	return strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}
