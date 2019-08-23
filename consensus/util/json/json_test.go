package json

import (
	"testing"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

func Test1(t *testing.T) {
	HandleJson("genesis.json","b.json",alter)
}

func alter(result map[string]interface{}){
	result["genesis_time"] = "2019-03-15 09:15:39"
	result["chain_id"] = "group-666"

	byteValue, err := ioutil.ReadFile("data.json")
	if err != nil {
		fmt.Println(err)
	}

	var data []interface{}
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		fmt.Println(err)
	}
	result["validators"] = data
}