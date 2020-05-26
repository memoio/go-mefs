package utils

import (
	"errors"
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

const (
	//BASETIME 用于记录的时间
	BASETIME = time.RFC3339

	//SHOWTIME 用于输出给使用者
	SHOWTIME = "2006-01-02 Mon 15:04:05 MST"

	//IDLength is network ID length
	IDLength = 30

	// DefaultPassword is default password
	DefaultPassword = "memoriae"

	// KeeperSLA is keeper needed
	KeeperSLA = 3
	// ProviderSLA is provider needed
	ProviderSLA = 6

	// DefaultCapacity is default store capacity： 1GB
	DefaultCapacity int64 = 1000 //MB
	// DefaultDuration is default store days： 100 days
	DefaultDuration int64 = 100 // day
	// DefaultCycle is default cycle: 1 day
	DefaultCycle = 24 * 60 * 60 // seconds

	// offer options

	// DefaultOfferCapacity is provider offer capacity
	DefaultOfferCapacity int64 = 1000000 //MB
	// DefaultOfferDuration is provider； 100 days
	DefaultOfferDuration int64 = 100 // day
	// DepositCapacity is provider deposit capacity
	DepositCapacity int64 = 1000000 // MB

	// Token2Wei is 1token = 10^18 wei
	// samely Dollar is 10^18 WeiDollar
	Token2Wei = 1000000000000000000

	// Memo2Dollar is default Memo token Price, 1 token = 0.01 dollar
	Memo2Dollar float64 = 0.01

	// ProviderDeposit is provider deposit price, 3 dollar/TB
	ProviderDeposit = 3000000000000 // 3*10^12 WeiDollar/MB
	// KeeperDeposit is keeper deposit； 0.01 dollar for now
	KeeperDeposit = 10000000000000000 // WeiDollar

	// READPRICE is read price 0.00002 $/GB(0.25 rmb-0.5rmb/GB in aliyun oss)
	READPRICE int64 = 20000000000 // 2*10^10 WeiDollar/MB

	// STOREPRICE is stored price 3$/TB*Month (33 rmb/TB*Month in aliyun oss)
	STOREPRICE int64 = 4000000000 // 4 * 10^9 WeiDollar/MB*hour
)

//CheckDup checks duplicate, false 意味着有，true表示无重复
func CheckDup(strs []string, s string) bool {
	for _, str := range strs {
		if strings.Compare(str, s) == 0 {
			return false
		}
	}
	return true
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

func BytesToString(b []byte) string {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := reflect.StringHeader{bh.Data, bh.Len}
	return *(*string)(unsafe.Pointer(&sh))
}

func StringToBytes(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{sh.Data, sh.Len, 0}
	return *(*[]byte)(unsafe.Pointer(&bh))
}

func ByteSliceToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func GetDirSize(path string) (uint64, error) {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return err
	})
	return size, err
}
