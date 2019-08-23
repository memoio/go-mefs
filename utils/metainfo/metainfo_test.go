package metainfo

import (
	"fmt"
	"strings"
	"testing"
)

//Test1 这个测试构造key和解析key的功能是否正确
//主要是看输入参数的数量和keytype是否匹配
func TestKeyMeta(t *testing.T) {
	_, err := NewKeyMeta("testMainID", UserInitReq, "123")
	if err != nil {
		fmt.Println("NewKeyMeta错误测试：", err)
	}
	km, _ := NewKeyMeta("testMainID", UserInitReq, "4", "6")
	strKM := km.ToString()
	fmt.Println("NewKeyMeta正确测试:", strKM)
	_, err = GetKeyMeta("testMainID" + DELIMITER + "2" + DELIMITER + "123")
	if err != nil {
		fmt.Println("GetKetMeta 错误测试：", err)
	}
	km, _ = GetKeyMeta("testMainID" + DELIMITER + "2" + DELIMITER + "keeperCount" + DELIMITER + "ProviderCount")
	strKM = km.ToString()
	fmt.Println("NewKeyMeta正确测试:", strKM)
}

func TestBlockMeta(t *testing.T) {
	_, err := NewBlockMeta("testUID", "testgid", "testsid", "testbid")
	if err != nil {
		fmt.Println("NewBlockMeta错误测试:", err)
	}
	bm, _ := NewBlockMeta("8MGm9D1EVTQwZYrkRVY4xuDJYjwJZT", "testgid", "testsid", "testbid")
	strBm := bm.ToString()
	fmt.Println("NewBlockMeta正确测试：", strBm)
	_, err = GetBlockMeta(strings.Join([]string{"testUID", "testgid", "testsid", "testbid"}, BLOCK_DELIMITER))
	if err != nil {
		fmt.Println("GetKetMeta 错误测试：", err)
	}
	bm, _ = GetBlockMeta(strings.Join([]string{"8MGm9D1EVTQwZYrkRVY4xuDJYjwJZT", "testgid", "testsid", "testbid"}, BLOCK_DELIMITER))
	fmt.Println("GetBlockMeta正确测试", bm.ToString())
}
