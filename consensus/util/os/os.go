package os

import (
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

//获取应用执行路径
func GetAppPath() string {
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	index := strings.LastIndex(path, string(os.PathSeparator))

	return path[:index]
}

// 把srcName文件复制到dstName
func CopyFile(srcName, dstName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

//创建文件
//fileName 文件路径
func CreateFile(fileName string) {
	file, err := os.Create(fileName)
	if err != nil {
		log.Println(err)
	}
	defer file.Close()
}

//获取n个可用的端口
func GetUsablePort(n int) []int {
	lp := GetListenPort()
	return GetNumberByList(lp, n)
}

// 获取一组数字中没有的指定个数个数字
func GetNumberByList(list []string, n int) (res []int) {
	// 转成数字
	var numList []int
	for i := 0; i < len(list); i++ {
		num, _ := strconv.Atoi(list[i])
		numList = append(numList, num)
	}
	// 排序
	sort.Ints(numList)
	// 查找两个数字之间是否有n个可用数字
	for i := 0; i < len(numList)-1; i++ {
		//只取端口>2000的端口
		if numList[i] > 2000 && numList[i+1]-numList[i]-1 >= n {
			for j := 1; j <= n; j++ {
				res = append(res, numList[i]+j)
			}
			return res
		}
	}
	return []int{}
}

// 获取监听的端口
func GetListenPort() []string {
	// 获取所有监听的端口
	cmd := exec.Command("cmd", "/C", "netstat -an")
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
	}
	str := string(out)
	// 格式化数据
	reg := regexp.MustCompile(" +")
	res := reg.ReplaceAllString(str, " ")
	strArr := strings.Split(res, "\r\n")
	var tmp []string
	for i := 4; i < len(strArr)-1; i++ {
		tmp = append(tmp, strings.Split(strArr[i], " ")[2])
	}
	// 获取所有端口
	for i := 0; i < len(tmp); i++ {
		tmp[i] = tmp[i][strings.LastIndex(tmp[i], ":")+1:]
	}

	return tmp
}
