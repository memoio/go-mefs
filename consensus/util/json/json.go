package json

import (
	"encoding/json"
	"io/ioutil"
)

//修改json文件
//jsonFile 输入json文件路径
//outFile 输出文件路径
//handle 处理函数
func HandleJson(jsonFile string, outFile string, handle func(map[string]interface{})) error {
	byteValue, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return err
	}

	var result map[string]interface{}
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return err
	}

	handle(result)

	byteValue, err = json.Marshal(result)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(outFile, byteValue, 0644)
	return err
}
