package utils

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	//BASETIME 用于记录的时间
	BASETIME = time.RFC3339

	//SHOWTIME 用于输出给使用者
	SHOWTIME = "2006-01-02 Mon 15:04:05 MST"

	//utils.IDLength  目前ID的长度
	IDLength = 30

	//READPRICEPERMB 读支付中1MB内容需要支付的金额
	READPRICEPERMB = 100

	//BlockSize 暂定一个块中纯data的大小，1k
	BlockSize = 1024 * 1024
)

//false 意味着有，true表示无重复
func CheckDup(strs []string, s string) bool {
	for _, str := range strs {
		if strings.Compare(str, s) == 0 {
			return false
		}
	}
	return true
}

func SplitIndex(index string) (string, int, error) {
	splitedIndex := strings.Split(index, "_")
	blockID := splitedIndex[0] + "_" + splitedIndex[1] + "_" + splitedIndex[2] + "_" + splitedIndex[3]
	offset, err := strconv.Atoi(splitedIndex[4])
	if err != nil {
		return "", 0, err
	}

	return blockID, offset, nil
}

//获取当前时间的函数，保证格式一致
func GetTimeNow() time.Time {
	return StringToTime(TimeToString(time.Now()))
}

//将记录中的时间变量转换为字符串的操作，保证格式一致
func TimeToString(time time.Time) string {
	return time.Format(BASETIME)
}

//将字符串转换成时间变量，保证格式一致
func StringToTime(stringTime string) time.Time {
	local, err := time.LoadLocation("Local")
	if err != nil {
		fmt.Println("time.LoadLocation error!:", err)
	}
	timerec, err := time.ParseInLocation(BASETIME, stringTime, local)
	if err != nil {
		fmt.Println("time.ParseInLocation error!:", err)
	}
	return timerec
}

func GetUnixNow() int64 {
	return time.Now().Unix()
}

//将unix时间戳转换为time格式，时区使用当前节点的时区
func UnixToTime(timeStamp int64) time.Time {
	return time.Unix(timeStamp, 0).In(time.Local)
}

func UnixToString(timeStamp int64) string {
	return strconv.FormatInt(timeStamp, 10)
}
func StringToUnix(stringTime string) int64 {
	unix, err := strconv.ParseInt(stringTime, 10, 64)
	if err != nil {
		fmt.Println("stringToUnix err:", err)
		fmt.Println("string:", stringTime)
		return 0
	}
	return unix
}

//HexToBigInt transfer hexString to BigInt
func HexToBigInt(hex string) *big.Int {
	s, _ := new(big.Int).SetString(hex[2:], 16)
	return s
}

func IntMin(x, y int) int {
	if x < y {
		return x
	}
	return y
}

// Writable ensures the directory exists and is writable
func Writable(path string) error {
	// Construct the path if missing
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	// Check the directory is writable
	if f, err := os.Create(filepath.Join(path, "._check_writable")); err == nil {
		f.Close()
		os.Remove(f.Name())
	} else {
		return errors.New("'" + path + "' is not writable")
	}
	return nil
}
