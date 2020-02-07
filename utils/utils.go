package utils

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
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
	READPRICEPERMB int64 = 1000000

	// Stored price 3$/TB*Month
	// 1 eth=0.01$
	// wei/MB*hour
	STOREPRICEPEDOLLAR int64 = 400000000000

	DefaultPassword = "123456789"
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
	if len(splitedIndex) < 4 {
		return "", 0, errors.New("String is too short")
	}
	offset, err := strconv.Atoi(splitedIndex[3])
	if err != nil {
		return "", 0, err
	}

	return strings.Join(splitedIndex[:3], "_"), offset, nil
}

// JoinStrings is
func JoinStrings(sep string, ops ...string) string {
	var res strings.Builder
	for _, op := range ops {
		res.WriteString(sep)
		res.WriteString(op)
	}
	return res.String()
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

// DisorderArray 对数组进行乱序操作，以便user随机选择providers
func DisorderArray(array []string) []string {
	var temp string
	var num int
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(array) - 1; i >= 0; i-- {
		num = r.Intn(i + 1)
		temp = array[i]
		array[i] = array[num]
		array[num] = temp
	}

	return array
}

func ValidIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if re.MatchString(ipAddress) {
		return true
	}
	return false
}
