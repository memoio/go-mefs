package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

const (
	KiB = 1024
	MiB = 1048576
	GiB = 1073741824
	TiB = 1099511627776

	KB = 1e3
	MB = 1e6
	GB = 1e9
	TB = 1e12

	Day    = 86400
	Hour   = 3600
	Minute = 60

	Wei   = 1
	GWei  = 1e9
	Token = 1e18
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
	DefaultCapacity int64 = 1024 //MB
	// DefaultDuration is default store days： 100 days
	DefaultDuration int64 = 100 // day
	// DefaultCycle is default cycle: 1 day
	DefaultCycle = 24 * 60 * 60 // seconds

	// offer options

	// DefaultOfferCapacity is provider default offer capacity: 1TB
	DefaultOfferCapacity int64 = 1048576 //MB
	// DefaultOfferDuration is provider； 100 days
	DefaultOfferDuration int64 = 100 // day
	// DepositCapacity is provider deposit capacity, 1TB
	DepositCapacity int64 = 1048576 // MB

	// 1token = 10^18 wei
	// samely Dollar is 10^18 WeiDollar

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

// GetDirSize gets dir size
func GetDirSize(path string) (uint64, error) {
	var size uint64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += uint64(info.Size())
		}
		return nil
	})
	return size, err
}

// FormatBytes convert bytes to human readable string. Like 2 MiB, 64.2 KiB, 52 B
func FormatBytes(i int64) (result string) {
	switch {
	case i >= TiB:
		result = fmt.Sprintf("%.02f TiB", float64(i)/TiB)
	case i >= GiB:
		result = fmt.Sprintf("%.02f GiB", float64(i)/GiB)
	case i >= MiB:
		result = fmt.Sprintf("%.02f MiB", float64(i)/MiB)
	case i >= KiB:
		result = fmt.Sprintf("%.02f KiB", float64(i)/KiB)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	return
}

// FormatBytesDec Convert bytes to base-10 human readable string. Like 2 MB, 64.2 KB, 52 B
func FormatBytesDec(i int64) (result string) {
	switch {
	case i >= TB:
		result = fmt.Sprintf("%.02f TB", float64(i)/TB)
	case i >= GB:
		result = fmt.Sprintf("%.02f GB", float64(i)/GB)
	case i >= MB:
		result = fmt.Sprintf("%.02f MB", float64(i)/MB)
	case i >= KB:
		result = fmt.Sprintf("%.02f KB", float64(i)/KB)
	default:
		result = fmt.Sprintf("%d B", i)
	}
	return
}

func FormatWei(i *big.Int) (result string) {
	f := new(big.Float).SetInt(i)
	res, _ := f.Float64()
	switch {
	case res >= Token:
		result = fmt.Sprintf("%.02f Token", res/Token)
	case res >= GWei:
		result = fmt.Sprintf("%.02f Gwei", res/GWei)
	default:
		result = fmt.Sprintf("%d Wei", i.Int64())
	}
	return
}

func FormatWeiDollar(i *big.Int) (result string) {
	f := new(big.Float).SetInt(i)
	res, _ := f.Float64()
	switch {
	case res >= Token:
		result = fmt.Sprintf("%.02f Dollar (For now, 1 Dollar = 100 Token)", res/Token)
	case res >= GWei:
		result = fmt.Sprintf("%.02f GweiDollar (For now, 1 GweiDollar = 100 Gwei)", res/GWei)
	default:
		result = fmt.Sprintf("%f WeiDollar (For now, 1 WeiDollar = 100 Wei)", res)
	}
	return
}

func FormatStorePrice(i *big.Int) (result string) {
	f := new(big.Float).SetInt(i)
	f.Mul(f, big.NewFloat(float64(1024*1024*24*30)))
	res, _ := f.Float64()
	switch {
	case res >= Token:
		result = fmt.Sprintf("%.02f Dollar/(TiB*Month) (For now, 1 Dollar = 100 Token)", res/Token)
	case res >= GWei:
		result = fmt.Sprintf("%.02f GweiDollar/(TiB*Month) (For now, 1 GweiDollar = 100 Gwei)", res/GWei)
	default:
		result = fmt.Sprintf("%f WeiDollar/(TiB*Month) (For now, 1 WeiDollar = 100 Wei)", res)
	}
	return
}

func FormatReadPrice(i *big.Int) (result string) {
	f := new(big.Float).SetInt(i)
	f.Mul(f, big.NewFloat(float64(1024)))
	res, _ := f.Float64()
	switch {
	case res >= Token:
		result = fmt.Sprintf("%.02f Dollar/GiB (For now, 1 Dollar = 100 Token)", res/Token)
	case res >= GWei:
		result = fmt.Sprintf("%.02f GweiDollar/GiB (For now, 1 GweiDollar = 100 Gwei)", res/GWei)
	default:
		result = fmt.Sprintf("%f WeiDollar/GiB (For now, 1 WeiDollar = 100 Wei)", res)
	}
	return
}

// FormatDuration convert time duration to human readable string
func FormatDuration(n int64) (result string) {
	d := time.Duration(n)
	if d > time.Hour*24 {
		result = fmt.Sprintf("%dd", d/24/time.Hour)
		d -= (d / time.Hour / 24) * (time.Hour * 24)
	}
	if d > time.Hour {
		result = fmt.Sprintf("%s%dh", result, d/time.Hour)
		d -= d / time.Hour * time.Hour
	}
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	result = fmt.Sprintf("%s%02dm%02ds", result, m, s)
	return
}

// FormatSecond convert seconds to human readable string
func FormatSecond(d int64) (result string) {
	if d > Day {
		result = fmt.Sprintf("%d day", d/Day)
		d -= (d / Day) * (Day)
	}

	if d == 0 {
		return
	}

	if d > Hour {
		result = fmt.Sprintf("%s %d hour", result, d/Hour)
		d -= (d / Hour) * Hour
	}

	if d == 0 {
		return
	}
	m := d / Minute
	d -= m * Minute
	result = fmt.Sprintf("%s %02d minute %02d second", result, m, d)
	return
}

// GetPassWord gets password from input
func GetPassWord() (string, error) {
	var password string
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	go func() {
		defer cancel()
		fmt.Printf("Please input your password (at least 8, or press 'Enter' to use default password): ")
		input := bufio.NewScanner(os.Stdin)
		ok := input.Scan()
		if ok {
			password = input.Text()
		}
	}()

	select {
	case <-ctx.Done():
	}

	if password == "" {
		fmt.Println("\nuse default password: ", DefaultPassword)
		password = DefaultPassword
	}

	if len(password) < 8 {
		fmt.Println("\n your password is too short, at least 8")
		return password, errors.New("Password is too short, length should be at least 8")
	}
	return password, nil
}

type DiskStats struct {
	Total uint64 `json:"all"`
	Used  uint64 `json:"used"`
	Free  uint64 `json:"free"`
}

// DiskStatus returns disk usage of path/disk
func DiskStatus(path string) (*DiskStats, error) {
	fs := syscall.Statfs_t{}
	err := syscall.Statfs(path, &fs)
	if err != nil {
		return nil, err
	}

	m := &DiskStats{
		Total: fs.Blocks * uint64(fs.Bsize),
		Free:  fs.Bfree * uint64(fs.Bsize),
	}
	m.Used = m.Total - m.Free
	return m, nil
}

// GetIntranetIP gets inter ip
func GetIntranetIP() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	var iAddrs []string
	for _, address := range addrs {
		// ip is loopback?
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				iAddrs = append(iAddrs, ipnet.IP.String())
			}
		}
	}

	return iAddrs, nil
}

// GetPulicIP gets public ip
func GetPulicIP() (string, error) {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(content)), nil
}

//IsReachable checks addr(ip:port) is reachable
func IsReachable(addr string) bool {
	cons, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return false
		}

		return true
	}

	cons.Close()

	return true
}
