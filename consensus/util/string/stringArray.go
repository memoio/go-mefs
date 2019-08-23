package string

import "strings"

//字符串数组转字节数组
func StringArrayToByteArray(str []string) []byte {
	stringByte := strings.Join(str, ",")
	return []byte(stringByte)
}

//字节数组转字符串数组
func ByteArrayToStringArray(byteArr []byte) []string {
	str := string(byteArr)
	return strings.Split(str, ",")
}

//在字符串数组中是否包含指定字符串
//strArr 字符串数组
//str 指定字符串
//@return true 存在
func ContainsByArray(strArr []string, str string) bool {
	for _, s := range strArr {
		if s == str {
			return true
		}
	}
	return false
}

//删除数组中指定值
//@return 处理后的数组
func DeleteByArray(strArr []string, str string) []string {
	i := IndexByArray(strArr, str)
	if i != -1 {
		strArr = append(strArr[:i], strArr[i+1:]...)
	}
	return strArr
}

//获取数组中字符串的位置
//@return -1 表示没有找到
func IndexByArray(strArr []string, str string) int {
	for i, s := range strArr {
		if s == str {
			return i
		}
	}
	return -1
}
